resource "aws_subnet" "private_1a" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[0], 2, 0)
  availability_zone = var.vpc_azs[0]
  tags = merge(
    {
      "Name"       = format("%s-%s-private-1a", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "private"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "private_1b" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[0], 2, 1)
  availability_zone = var.vpc_azs[1]
  tags = merge(
    {
      "Name"       = format("%s-%s-private-1b", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "private"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "private_1c" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[0], 2, 2)
  availability_zone = var.vpc_azs[2]
  tags = merge(
    {
      "Name"       = format("%s-%s-private-1c", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "private"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "private_1d" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[0], 2, 3)
  availability_zone = var.vpc_azs[3]
  tags = merge(
    {
      "Name"       = format("%s-%s-private-1d", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "private"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}



resource "aws_subnet" "public_1a" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[1], 2, 0)
  availability_zone = var.vpc_azs[0]
  tags = merge(
    {
      "Name"       = format("%s-%s-public-1a", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "public"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "public_1b" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[1], 2, 1)
  availability_zone = var.vpc_azs[1]
  tags = merge(
    {
      "Name"       = format("%s-%s-public-1b", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "public"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "public_1c" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[1], 2, 2)
  availability_zone = var.vpc_azs[2]
  tags = merge(
    {
      "Name"       = format("%s-%s-public-1c", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "public"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

resource "aws_subnet" "public_1d" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(cidrsubnets(var.vpc_cidr, 1, 1)[1], 2, 3)
  availability_zone = var.vpc_azs[3]
  tags = merge(
    {
      "Name"       = format("%s-%s-public-1d", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
      "SubnetType" = "public"
    },
    var.tags
  )
  lifecycle {
    ignore_changes = [
      tags,
    ]
  }
}

