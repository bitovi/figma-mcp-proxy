resource "aws_security_group" "rdp" {
  name        = "win-rdp-${var.client_name}-${local.short_environment}"
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
  name        = "${local.fully_qualified_name}-sg-${local.short_environment}"
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
  instance_type               = "t3.large"
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
  name = "ec2-ssm-role-${var.client_name}-${local.short_environment}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = { Service = "ec2.amazonaws.com" },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.ssm.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}
resource "aws_iam_role_policy_attachment" "ssm_cloudwatch" {
  role       = aws_iam_role.ssm.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_iam_instance_profile" "ssm" {
  name = "ec2-ssm-profile-${var.client_name}-${local.short_environment}"
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
        "if (!(Test-Path 'C:\\ProgramData\\chocolatey\\bin\\choco.exe')) { iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1')) }",
        "$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
        "& 'C:\\ProgramData\\chocolatey\\bin\\choco.exe' install git golang nssm -y --force",
        "$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
        "New-NetFirewallRule -DisplayName \"Allow App Port 3846\" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 3846 -ErrorAction SilentlyContinue",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' stop FigmaProxy -ErrorAction SilentlyContinue",
        "Start-Sleep -Seconds 5",
        "cd C:\\",
        "if (Test-Path 'C:\\figma-mcp-proxy') { Remove-Item -Recurse -Force 'C:\\figma-mcp-proxy' -ErrorAction SilentlyContinue }",
        "& 'C:\\Program Files\\Git\\cmd\\git.exe' clone https://github.com/bitovi/figma-mcp-proxy.git",
        "cd C:\\figma-mcp-proxy",
        "& 'C:\\Program Files\\Go\\bin\\go.exe' build -o figma-proxy.exe main.go",
        "New-Item -ItemType Directory -Force -Path 'C:\\figma-mcp-proxy\\logs'",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' remove FigmaProxy confirm",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' install FigmaProxy 'C:\\figma-mcp-proxy\\figma-proxy.exe'",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppDirectory 'C:\\figma-mcp-proxy'",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppEnvironmentExtra API_KEY=${var.proxy_api_token} EXTERNAL_DNS_NAME=${aws_route53_record.app.fqdn}",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppStdout 'C:\\figma-mcp-proxy\\logs\\stdout.log'",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppStderr 'C:\\figma-mcp-proxy\\logs\\stderr.log'",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppStdoutCreationDisposition 4",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppStderrCreationDisposition 4",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppRotateFiles 1",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppRotateOnline 1",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' set FigmaProxy AppRotateBytes 1048576",
        "& 'C:\\ProgramData\\chocolatey\\bin\\nssm.exe' start FigmaProxy",
        "$cwAgentUrl = 'https://s3.amazonaws.com/amazoncloudwatch-agent/windows/amd64/latest/amazon-cloudwatch-agent.msi'",
        "$cwAgentPath = Join-Path $env:TEMP 'amazon-cloudwatch-agent.msi'",
        "if (Test-Path $cwAgentPath) { Remove-Item $cwAgentPath -Force -ErrorAction SilentlyContinue }",
        "Invoke-WebRequest -Uri $cwAgentUrl -OutFile $cwAgentPath",
        "Start-Process msiexec.exe -ArgumentList '/i', $cwAgentPath, '/qn' -Wait",
        "Start-Sleep -Seconds 10",
        "$cwAgentDir = 'C:\\Program Files\\Amazon\\AmazonCloudWatchAgent'",
        "if (!(Test-Path $cwAgentDir)) { New-Item -ItemType Directory -Force -Path $cwAgentDir }",
        "$cwConfigPath = Join-Path $cwAgentDir 'config.json'",
        "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/bitovi/figma-mcp-proxy/main/infra/cloudwatch-config.json' -OutFile $cwConfigPath",
        "$cwAgentCtl = Join-Path $cwAgentDir 'amazon-cloudwatch-agent-ctl.ps1'",
        "if (Test-Path $cwAgentCtl) { & $cwAgentCtl -a fetch-config -m ec2 -s -c file:$cwConfigPath }"
    ])
  }
}
