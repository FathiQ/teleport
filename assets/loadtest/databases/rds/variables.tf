variable "prefix" {
  type        = string
  description = "Prefix added to all resources."
  default     = "loadtest"
}

variable "region" {
  type        = string
  description = "The region to manage resources in."
  default     = "us-east-1"
}

variable "instance_class" {
  type        = string
  description = "Database instance machine class."
  default     = "db.t3.medium"
}

variable "create_postgres" {
  type        = bool
  description = "Determines if the PostgreSQL instance will be created."
  default     = true
}

variable "postgres_port" {
  type        = number
  description = "PostgreSQL database port"
  default     = 5432
}

variable "mysql_port" {
  type        = number
  description = "MySQL database port"
  default     = 3306
}

variable "create_mysql" {
  type        = bool
  description = "Determines if the MySQL instance will be created."
  default     = true
}

variable "local_ip" {
  type        = string
  description = "Local IP address. It is used to create a security group with external access, which is required when Terraform creates the Teleport database user."
}

variable "database_master_username" {
  type        = string
  description = "Database master username"
  default     = "postgres"
}

variable "database_name" {
  type        = string
  description = "Database name"
  default     = "postgres"
}

variable "teleport_database_user" {
  type        = string
  description = "Teleport database username"
  default     = "alice"
}

variable "eks_cluster_name" {
  type        = string
  description = "EKS cluster name"
}

variable "database_access_namespace" {
  type        = string
  description = "Database agents EKS cluster namespace"
  default     = "database-agents"
}

variable "database_access_svc_account_name" {
  type        = string
  description = "Database agent service account name"
  default     = "database-agents"
}
