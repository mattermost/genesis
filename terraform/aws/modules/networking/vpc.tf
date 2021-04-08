
resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = var.enable_dns_hostnames
  tags = merge(
    {
      "Name"              = format("%s-%s", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "Available"         = "true",
      "CloudClusterID"    = "none",
      "CloudClusterOwner" = "none",
      "Size"              = split("/", var.vpc_cidr)[1]
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      # Ignore changes to tag Available
      tags["Available"],
      tags["CloudClusterID"],
      tags["CloudClusterOwner"]
    ]
  }
}



