resource "aws_eip" "nat_eip" {
  vpc = true
  tags = merge(
    {
      "Name" = format("%s-%s", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags,
  )
}

resource "aws_nat_gateway" "nat_gtw" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = aws_subnet.public_1a.id
  tags = merge(
    {
      "Name" = format("%s-%s", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags,
  )

  depends_on = [aws_internet_gateway.internet_gtw]
}
