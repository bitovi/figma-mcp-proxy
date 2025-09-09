locals {
  administrator_password = var.private_key_path != "" ? rsadecrypt(aws_instance.win2025.password_data, local.private_key_pem) : ""
}

output "instance_id" {
  value       = aws_instance.win2025.id
  description = "EC2 instance ID"
}

output "public_ip" {
  value       = aws_instance.win2025.public_ip
  description = "Public IP"
}

output "public_dns" {
  value       = aws_instance.win2025.public_dns
  description = "Public DNS"
}

output "ami_id" {
  value       = aws_instance.win2025.ami
  description = "AMI ID used"
  sensitive = true
}

output "administrator_password" {
  value       = local.administrator_password
  sensitive   = true
  description = "Decrypted Windows Administrator password (populated only if private_key_path is provided)"
}

output "windows_username" {
    value       = "Administrator"
    description = "Default Windows local Administrator username"
}

output "fqdn" {
    value       = aws_route53_record.app_ipv6.fqdn
    description = "Fully qualified domain name for the application"
}

