resource "aws_ec2_transit_gateway_vpc_attachment" "tgw_attachment" {
  subnet_ids = [
    aws_subnet.private_1a.id,
    aws_subnet.private_1b.id,
    aws_subnet.private_1c.id,
    aws_subnet.private_1d.id
  ]
  transit_gateway_id = var.transit_gateway_id
  vpc_id             = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = format("%s-%s", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags
  )
}


