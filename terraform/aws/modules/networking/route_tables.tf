
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = format("%s-%s-public-rtb", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags
  )
}


resource "aws_route_table" "private" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = format("%s-%s-private-rtb", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags
  )
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gtw.id

  timeouts {
    create = "5m"
  }
}


resource "aws_route" "private_nat_gateway" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.nat_gtw.id

  timeouts {
    create = "5m"
  }
}

resource "aws_route" "transit_gateway_private" {
  for_each = toset(var.transit_gtw_route_destinations)

  route_table_id         = aws_route_table.public.id
  destination_cidr_block = each.value
  transit_gateway_id     = var.transit_gateway_id
  depends_on             = [aws_ec2_transit_gateway_vpc_attachment.tgw_attachment]
}

resource "aws_route" "transit_gateway_public" {
  for_each = toset(var.transit_gtw_route_destinations)

  route_table_id         = aws_route_table.public.id
  destination_cidr_block = each.value
  transit_gateway_id     = var.transit_gateway_id
  depends_on             = [aws_ec2_transit_gateway_vpc_attachment.tgw_attachment]
}

resource "aws_route_table_association" "private_1a" {
  subnet_id      = aws_subnet.private_1a.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "private_1b" {
  subnet_id      = aws_subnet.private_1b.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "private_1c" {
  subnet_id      = aws_subnet.private_1c.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "private_1d" {
  subnet_id      = aws_subnet.private_1d.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "public_1a" {
  subnet_id      = aws_subnet.public_1a.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_1b" {
  subnet_id      = aws_subnet.public_1b.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_1c" {
  subnet_id      = aws_subnet.public_1c.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_1d" {
  subnet_id      = aws_subnet.public_1d.id
  route_table_id = aws_route_table.public.id
}

