resource "aws_internet_gateway" "internet_gtw" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = format("%s-%s", var.name, join("", split(".", split("/", var.vpc_cidr)[0]))),
    },
    var.tags
  )
}
