resource "aws_security_group" "alb_sg" {
  name        = "${local.fully_qualified_name}-alb-sg"
  description = "ALB SG"
  vpc_id      = data.aws_vpc.default_vpc.id

  ingress {
    description      = "HTTP from anywhere"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "HTTPS from anywhere"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_lb" "app" {
  name               = "${local.fully_qualified_name}-alb"
  load_balancer_type = "application"
  internal           = false
  security_groups    = [aws_security_group.alb_sg.id]
  subnets            = data.aws_subnets.default_subnets.ids
}

resource "aws_lb_target_group" "app" {
    name        = "${local.fully_qualified_name}-app-tg"
    port        = var.container_port
    protocol    = "HTTP"
    target_type = "instance"
    vpc_id      = data.aws_vpc.default_vpc.id

    health_check {
        path                = "/health"
        protocol            = "HTTP"
        matcher             = "200-399"
    }
}

resource "aws_lb_target_group_attachment" "win" {
    target_group_arn = aws_lb_target_group.app.arn
    target_id        = aws_instance.win2025.id
    port             = var.container_port
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.app.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.app.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.acm_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

resource "aws_route53_record" "app_ipv6" {
  zone_id = var.hosted_zone_id
  name    = "${local.fully_qualified_name}"
  type    = "AAAA"

    alias {
    name                   = aws_lb.app.dns_name
    zone_id                = aws_lb.app.zone_id
    evaluate_target_health = true
    }
}
