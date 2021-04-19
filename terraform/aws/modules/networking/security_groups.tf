resource "aws_security_group" "master_sg" {
  name        = format("%s-%s-master-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0])))
  description = "Master Nodes Security Group"
  vpc_id      = aws_vpc.vpc.id

  tags = merge(
    {
      "Name"     = format("%s-%s-master-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "NodeType" = "master"
    },
    var.tags
  )
}

resource "aws_security_group" "worker_sg" {
  name        = format("%s-%s-worker-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0])))
  description = "Worker Nodes Security Group"
  vpc_id      = aws_vpc.vpc.id

  tags = merge(
    {
      "Name"     = format("%s-%s-worker-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "NodeType" = "worker"
    },
    var.tags
  )
}

# PostgreSQL database security group
resource "aws_security_group" "db_sg_postgresql" {
  name        = format("%s-%s-db-postgresql-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0])))
  description = "RDS Database PostgreSQL Security Group"
  vpc_id      = aws_vpc.vpc.id
  tags = merge(
    {
      "Name"                                = format("%s-%s-db-sg", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "MattermostCloudInstallationDatabase" = "PostgreSQL/Aurora"
    },
    var.tags
  )
}

# Master Rules
resource "aws_security_group_rule" "master_egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = "Outbound Traffic"
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.master_sg.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "master_ingress_worker" {
  source_security_group_id = aws_security_group.worker_sg.id
  description              = "Ingress Traffic from Worker Nodes"
  from_port                = 443
  protocol                 = "TCP"
  security_group_id        = aws_security_group.master_sg.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "master_ingress_teleport" {
  type              = "ingress"
  from_port         = 3022
  to_port           = 3022
  protocol          = "tcp"
  cidr_blocks       = var.teleport_cidr
  security_group_id = aws_security_group.master_sg.id
}

# Worker Rules
resource "aws_security_group_rule" "worker_egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = "Outbound Traffic"
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.worker_sg.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "worker_ingress_worker" {
  self              = true
  description       = "Ingress Traffic from Worker Nodes"
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.worker_sg.id
  to_port           = 0
  type              = "ingress"
}

resource "aws_security_group_rule" "worker_ingress_master" {
  source_security_group_id = aws_security_group.master_sg.id
  description              = "Ingress Traffic from Master Nodes"
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.worker_sg.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "worker_ingress_teleport" {
  type              = "ingress"
  from_port         = 3022
  to_port           = 3022
  protocol          = "tcp"
  cidr_blocks       = var.teleport_cidr
  security_group_id = aws_security_group.worker_sg.id
}

# PostgreSQL DB Rules
resource "aws_security_group_rule" "db_ingress_worker_postgresql" {
  source_security_group_id = aws_security_group.worker_sg.id
  description              = "Ingress Traffic from Worker Nodes"
  from_port                = 5432
  protocol                 = "TCP"
  security_group_id        = aws_security_group.db_sg_postgresql.id
  to_port                  = 5432
  type                     = "ingress"
}

resource "aws_security_group_rule" "db_ingress_worker_command_control_postgresql" {
  cidr_blocks       = var.command_and_control_private_subnet_cidrs
  description       = "Ingress Traffic from Command and Control Private Subnets"
  from_port         = 5432
  protocol          = "TCP"
  security_group_id = aws_security_group.db_sg_postgresql.id
  to_port           = 5432
  type              = "ingress"
}
