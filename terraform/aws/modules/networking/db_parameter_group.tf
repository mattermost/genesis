resource "aws_db_parameter_group" "db_parameter_group" {
  name   = "mattermost-provisioner-rds-pg"
  family = "aurora-postgresql11"

  parameter {
    name = "random_page_cost"
    value = var.random_page_cost
  }

  parameter {
    name = "tcp_keepalives_count"
    value = var.tcp_keepalives_count
  }

  parameter {
    name = "tcp_keepalives_idle"
    value = var.tcp_keepalives_idle
  }

  parameter {
    name = "tcp_keepalives_interval"
    value = var.tcp_keepalives_interval
  }

  tags = merge(
    {
      "MattermostCloudInstallationDatabase" = "PostgreSQL/Aurora"
    },
    var.tags
  )
}

resource "aws_rds_cluster_parameter_group" "cluster_parameter_group" {
  name   = "mattermost-provisioner-rds-cluster-pg"
  family = "aurora-postgresql11"
  
  parameter {
    name = "random_page_cost"
    value = var.random_page_cost
  }

  parameter {
    name = "tcp_keepalives_count"
    value = var.tcp_keepalives_count
  }

  parameter {
    name = "tcp_keepalives_idle"
    value = var.tcp_keepalives_idle
  }

  parameter {
    name = "tcp_keepalives_interval"
    value = var.tcp_keepalives_interval
  }

  tags = merge(
    {
      "MattermostCloudInstallationDatabase" = "PostgreSQL/Aurora"
    },
    var.tags
  )

}
