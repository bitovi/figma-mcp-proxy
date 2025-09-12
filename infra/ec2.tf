resource "aws_security_group" "rdp" {
  name        = "win-rdp-${var.client_name}"
  description = "Allow RDP from my IP"
  vpc_id      = data.aws_vpc.default_vpc.id

  ingress {
    description = "RDP"
    protocol    = "tcp"
    from_port   = 3389
    to_port     = 3389
    cidr_blocks = local.allowed_cidrs
  }

  egress {
    description = "All outbound"
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "win-rdp-${var.client_name}"
  }
}

resource "aws_security_group" "lb_to_ec2_sg" {
  name        = "${local.fully_qualified_name}-sg"
  description = "EC2 tasks behind ALB"
  vpc_id      = data.aws_vpc.default_vpc.id

  ingress {
    description     = "From ALB only"
    from_port       = var.container_port
    to_port         = var.container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.alb_sg.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_instance" "win2025" {
  ami                         = data.aws_ssm_parameter.win2025.value
  iam_instance_profile        = aws_iam_instance_profile.ssm.name
  instance_type               = "t3.medium"
  subnet_id                   = element(data.aws_subnets.default_subnets.ids, 0)
  vpc_security_group_ids      = [aws_security_group.rdp.id, aws_security_group.lb_to_ec2_sg.id]
  associate_public_ip_address = true

  # Provide an existing key pair name so you can decrypt the Windows password
  key_name          = var.key_name
  get_password_data = true

  tags = {
    Name = "${local.fully_qualified_name}-ec2"
  }
}

resource "aws_iam_role" "ssm" {
  name = "ec2-ssm-role-${var.client_name}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect = "Allow",
      Principal = { Service = "ec2.amazonaws.com" },
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.ssm.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "ssm" {
  name = "ec2-ssm-profile-${var.client_name}"
  role = aws_iam_role.ssm.name
}

resource "aws_ssm_association" "install_tools" {
  name = "AWS-RunPowerShellScript"

  targets {
    key    = "InstanceIds"
    values = [aws_instance.win2025.id]
  }

  parameters = {
    commands = join("\n", [
        "Set-ExecutionPolicy Bypass -Scope Process -Force",
        "[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072",
        "iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))",
        "choco install git -y",
        "choco install golang -y",
        "New-NetFirewallRule -DisplayName \"Allow App Port 3846\" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 3846"
    ])
  }
}
