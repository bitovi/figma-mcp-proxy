data "aws_vpc" "default_vpc" {
    default = true
}

data "aws_subnets" "default_subnets" {
    filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default_vpc.id]
    }
}

data "aws_ssm_parameter" "win2025" {
  name = "/aws/service/ami-windows-latest/Windows_Server-2025-English-Full-Base"
}
