// Copyright 2022 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/httplib/csrf"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/web"
	websession "github.com/gravitational/teleport/lib/web/session"
	"github.com/gravitational/teleport/lib/web/ui"
)

// WebClientPack is an authenticated HTTP Client for Teleport.
type WebClientPack struct {
	clt         *http.Client
	host        string
	webCookie   string
	bearerToken string
	clusterName string
}

// LoginWebClient receives the host url, the username and a password.
// It will login into that host and return a WebClientPack.
func LoginWebClient(t *testing.T, host, username, password string) *WebClientPack {
	csReq, err := json.Marshal(web.CreateSessionReq{
		User: username,
		Pass: password,
	})
	require.NoError(t, err)

	// Create POST request to create session.
	u := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/v1/webapi/sessions/web",
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(csReq))
	require.NoError(t, err)

	// Attach CSRF token in cookie and header.
	csrfToken, err := utils.CryptoRandomHex(32)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  csrf.CookieName,
		Value: csrfToken,
	})
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set(csrf.HeaderName, csrfToken)

	// Issue request.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read in response.
	var csResp *web.CreateSessionResponse
	err = json.NewDecoder(resp.Body).Decode(&csResp)
	require.NoError(t, err)

	// Extract session cookie and bearer token.
	require.Len(t, resp.Cookies(), 1)
	cookie := resp.Cookies()[0]
	require.Equal(t, cookie.Name, websession.CookieName)

	webClient := &WebClientPack{
		clt:         client,
		host:        host,
		webCookie:   cookie.Value,
		bearerToken: csResp.Token,
	}

	respStatusCode, bs := webClient.DoRequest(t, http.MethodGet, "sites", nil)
	require.Equal(t, http.StatusOK, respStatusCode, string(bs))

	var clusters []ui.Cluster
	require.NoError(t, json.Unmarshal(bs, &clusters), string(bs))
	require.NotEmpty(t, clusters)

	webClient.clusterName = clusters[0].Name
	return webClient
}

// DoRequest receives a method, endpoint and payload and sends an HTTP Request to the Teleport API.
// The endpoint must not contain the host neither the base path ('/v1/webapi/').
// Status Code and Body are returned.
// "$site" in the endpoint is substituted by the current site.
func (w *WebClientPack) DoRequest(t *testing.T, method, endpoint string, payload any) (int, []byte) {
	endpoint = fmt.Sprintf("https://%s/v1/webapi/%s", w.host, endpoint)
	endpoint = strings.ReplaceAll(endpoint, "$site", w.clusterName)
	u, err := url.Parse(endpoint)
	require.NoError(t, err)

	bs, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(bs))
	require.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  websession.CookieName,
		Value: w.webCookie,
	})
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", w.bearerToken))
	req.Header.Add("Content-Type", "application/json")

	resp, err := w.clt.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}

// OpenWebsocket opens a websocket on a given Teleport API endpoint.
// The endpoint must not contain the host neither the base path ('/v1/webapi/').
// Raw websocket and HTTP response are returned.
// "$site" in the endpoint is substituted by the current site.
func (w *WebClientPack) OpenWebsocket(t *testing.T, endpoint string, params any) (*websocket.Conn, *http.Response, error) {
	path, err := url.JoinPath("v1", "webapi", strings.ReplaceAll(endpoint, "$site", w.clusterName))
	require.NoError(t, err)

	u := url.URL{
		Host:   w.host,
		Scheme: client.WSS,
		Path:   path,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, nil, err
	}

	q := u.Query()
	q.Set("params", string(data))
	q.Set(roundtrip.AccessTokenQueryParam, w.bearerToken)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{}
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	cookie := &http.Cookie{
		Name:  websession.CookieName,
		Value: w.webCookie,
	}

	header := http.Header{}
	header.Add("Origin", "http://localhost")
	header.Add("Cookie", cookie.String())

	ws, resp, err := dialer.Dial(u.String(), header)
	return ws, resp, trace.Wrap(err)
}
