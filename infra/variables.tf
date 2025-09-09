variable "app_name" {
    description = "Name of the application"
    type        = string
    default = "figma-proxy"
}

variable "aws_region" {
    description = "AWS region for deployment"
    type        = string
}

variable "client_name" {
    description = "Name of the client"
    type        = string
}

variable "target_environment" {
    description = "Name of the target environment"
    type        = string
}

variable "acm_certificate_arn" {
    description = "ARN of the ACM certificate for HTTPS"
    type        = string
}

variable "container_port" {
    description = "Port on which the container listens"
    type        = number
    default     = 3846
}

variable "hosted_zone_id" {
    description = "ID of the hosted zone"
    type        = string
}

variable "key_name" {
    description = "Name of an existing EC2 key pair in the target region (for Windows password decryption)"
    type        = string
}

variable "private_key_path" {
description = "Path to your private key (.pem) that matches var.key_name (optional)"
type        = string
default     = ""
}

variable "allowed_cidrs_file" {
  description = "Path to a text file with one CIDR per line (e.g., 203.0.113.5/32). Lines starting with # are treated as comments."
  type        = string
  default     = "allowed_cidrs.txt"
}