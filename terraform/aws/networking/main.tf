terraform {
  required_version = ">= 0.13"
  backend "s3" {
    bucket = "terraform-genesis-state-bucket-test"
    key    = "us-east-1/cloud-enterprise-networking"
    region = "us-east-1"
    profile = "central-monitoring-test"
  }
  required_providers {
    aws = "~> 3.15.0"
  }
}


provider "aws" {
  region = var.region
  profile = "mattermost-control-tower"
  assume_role {
    role_arn     = "arn:aws:iam::${var.account_id}:role/MattermostAccountProvisioningRole"
    session_name = "account-provisioning"
  }
}

module "networking" {
  source                                   = "../modules/networking"
  environment                              = var.environment
  vpc_cidr                                 = var.vpc_cidr
  vpc_azs                                  = var.vpc_azs
  name                                     = "mattermost-cloud-${var.environment}-enterprise"
  enable_dns_hostnames                     = true
  transit_gateway_id                       = var.transit_gateway_id
  transit_gtw_route_destinations           = var.transit_gtw_route_destinations
  region                                   = var.region
  teleport_cidr                            = var.teleport_cidr
  command_and_control_private_subnet_cidrs = var.command_and_control_private_subnet_cidrs
  tcp_keepalives_count                     = var.tcp_keepalives_count
  tcp_keepalives_idle                      = var.tcp_keepalives_idle
  tcp_keepalives_interval                  = var.tcp_keepalives_interval
  random_page_cost                         = var.random_page_cost
  private_dns_ips                          = var.private_dns_ips

  tags = {
    Owner       = "cloud-team"
    Terraform   = "true"
    Environment = var.environment
    Purpose     = "provisioning"
  }
}

