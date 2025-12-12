resource "aws_cloudwatch_log_group" "figma_proxy" {
  name              = "/aws/ec2/figma-proxy-${var.target_environment}"
  retention_in_days = 7

  tags = {
    Name        = "${local.fully_qualified_name}-logs"
    Environment = var.target_environment
    Client      = var.client_name
  }
}
