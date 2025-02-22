/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { wait } from 'shared/utils/wait';

import { MainProcessClient } from 'teleterm/mainProcess/types';
import {
  Cluster,
  CreateConnectMyComputerRoleResponse,
  Server,
  TshClient,
} from 'teleterm/services/tshd/types';

import type * as uri from 'teleterm/ui/uri';

export class ConnectMyComputerService {
  constructor(
    private mainProcessClient: MainProcessClient,
    private tshClient: TshClient
  ) {}

  async downloadAgent(): Promise<void> {
    await this.mainProcessClient.downloadAgent();
  }

  createRole(
    rootClusterUri: uri.RootClusterUri
  ): Promise<CreateConnectMyComputerRoleResponse> {
    return this.tshClient.createConnectMyComputerRole(rootClusterUri);
  }

  async createAgentConfigFile(cluster: Cluster): Promise<{
    token: string;
  }> {
    const { token, labelsList } =
      await this.tshClient.createConnectMyComputerNodeToken(cluster.uri);

    await this.mainProcessClient.createAgentConfigFile({
      rootClusterUri: cluster.uri,
      proxy: cluster.proxyHost,
      token: token,
      labels: labelsList,
    });

    return { token };
  }

  runAgent(rootClusterUri: uri.RootClusterUri): Promise<void> {
    return this.mainProcessClient.runAgent({
      rootClusterUri,
    });
  }

  killAgent(rootClusterUri: uri.RootClusterUri): Promise<void> {
    return this.mainProcessClient.killAgent({ rootClusterUri });
  }

  isAgentConfigFileCreated(
    rootClusterUri: uri.RootClusterUri
  ): Promise<boolean> {
    return this.mainProcessClient.isAgentConfigFileCreated({ rootClusterUri });
  }

  deleteToken(
    rootClusterUri: uri.RootClusterUri,
    token: string
  ): Promise<void> {
    return this.tshClient.deleteConnectMyComputerToken(rootClusterUri, token);
  }

  async waitForNodeToJoin(
    rootClusterUri: uri.RootClusterUri,
    abortSignal: AbortSignal
  ): Promise<Server> {
    // TODO(gzdunek): Replace with waiting for the node to join.
    await wait(1_000, abortSignal);
    return {
      uri: `${rootClusterUri}/servers/178ef081-259b-4aa5-a018-449b5ea7e694`,
      tunnel: false,
      name: '178ef081-259b-4aa5-a018-449b5ea7e694',
      hostname: 'foo',
      addr: '127.0.0.1:3022',
      labelsList: [
        {
          name: 'hostname',
          value: 'mbp.home',
        },
        {
          name: 'teleport.dev/connect-my-computer',
          value: 'abcd@goteleport.com',
        },
      ],
    };
  }
}
