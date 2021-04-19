resource "aws_db_subnet_group" "provisioner_db_subnet_group_postgresql" {

  name = "mattermost-provisioner-db-${aws_vpc.vpc.id}-postgresql"

  subnet_ids = [
    aws_subnet.private_1a.id,
    aws_subnet.private_1b.id,
    aws_subnet.private_1c.id,
    aws_subnet.private_1d.id
  ]

  tags = merge(
    {
      "MattermostCloudInstallationDatabase" = "PostgreSQL/Aurora"
    },
    var.tags
  )
}
