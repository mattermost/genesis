resource "aws_vpc_dhcp_options" "dhcp_options" {
  domain_name_servers = var.private_dns_ips
  domain_name         = "ec2.internal"

  tags = {
    Name = "${var.name}-dhcp-options"
  }
}

resource "aws_vpc_dhcp_options_association" "command_control_association" {
  vpc_id            = aws_vpc.vpc.id
  dhcp_options_id   = aws_vpc_dhcp_options.dhcp_options.id
}
