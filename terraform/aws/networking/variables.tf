variable "environment" {
  default = ""
  type    = string
}

variable "region" {
  default = ""
  type    = string
}

variable "vpc_cidr" {
  default = ""
  type    = string
}

variable "vpc_azs" {
  default = ["us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d"]
  type    = list(string)
}

variable "transit_gateway_id" {
  default = ""
  type    = string
}

variable "transit_gtw_route_destinations" {
  default     = [""]
  type        = list(string)
  description = "The destinations of the transit gateway route table entry for monitoring and provisioning tools"
}

variable "teleport_cidr" {
  default     = [""]
  type        = list(string)
  description = "The Teleport CIDR for ssh access"
}

variable "command_and_control_private_subnet_cidrs" {
  default     = [""]
  type        = list(string)
  description = "The CIDRs of the command and control private subnets to be allowed by the provisioning DBs"
}

variable "tcp_keepalives_count" {
  default = 5
  description = "Maximum number of TCP keepalive retransmits.Specifies the number of TCP keepalive messages that can be lost before the server's connection to the client is considered dead. A value of 0 (the default) selects the operating system's default."
  type = number
}

variable "random_page_cost" {
  default = 1.1
  description = "Sets the planner's estimate of the cost of a non-sequentially-fetched disk page. The default is 4.0. This value can be overridden for tables and indexes in a particular tablespace by setting the tablespace parameter of the same name."
  type = number
}

variable "tcp_keepalives_idle" {
  default = 5
  description = "Time between issuing TCP keepalives.Specifies the amount of time with no network activity after which the operating system should send a TCP keepalive message to the client. If this value is specified without units, it is taken as seconds. A value of 0 (the default) selects the operating system's default."
  type = number
}

variable "tcp_keepalives_interval" {
  default = 1
  description = "Time between TCP keepalive retransmits. Specifies the amount of time after which a TCP keepalive message that has not been acknowledged by the client should be retransmitted. If this value is specified without units, it is taken as seconds. A value of 0 (the default) selects the operating system's default."
  type = number
}

variable "private_dns_ips" {
  default     = [""]
  type        = list(string)
  description = "Private DNS IPs used by the Bind server"
}

variable "account_id" {
  default = ""
  type    = string
  description = "The account ID that will be provisioned"
}
