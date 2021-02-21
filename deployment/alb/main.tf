data "aws_vpc" "impart-backend" {
  id  = "vpc-0bfb2f75636067f54"
}

data "aws_route53_zone" "impart_domain" {
  name         = "impartwealth.com"
  private_zone = false
}

resource "aws_lb" "backend" {
  name               = "backend-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [module.sg.id]
  subnets            = ["subnet-06f29f59259015320", "subnet-0bde91099ab9f4553"]

  enable_deletion_protection = false //todo - change this when we get more clarity

//  access_logs {
//    bucket  = aws_s3_bucket.lb_logs.bucket
//    prefix  = "test-lb"
//    enabled = true
//  }

  tags = {
    Environment = "production"
  }
  lifecycle {
    prevent_destroy = false
  }
}

resource "aws_route53_record" "backend_services" {
  zone_id = data.aws_route53_zone.impart_domain.zone_id
  name    = var.app_dns_entry
  type    = "CNAME"
  ttl     = "300"
  records = [aws_lb.backend.dns_name]
  lifecycle {
    prevent_destroy = false
  }
}

locals {
  private_cidr_block = cidrsubnet(data.aws_vpc.impart-backend.cidr_block, 1, 1)
}

module "sg" {
  source  = "cloudposse/security-group/aws"
  version = "0.1.3"
  vpc_id = data.aws_vpc.impart-backend.id
  delimiter = ""
  name = "443-ingress-private-egress"
  rules = [
    {
      type        = "egress"
      from_port   = 0
      to_port     = 65535
      protocol    = "TCP"
      cidr_blocks = [local.private_cidr_block]
    },
    {
      type        = "ingress"
      from_port   = 443
      to_port     = 443
      protocol    = "TCP"
      cidr_blocks = ["0.0.0.0/0"]
    },
    {
      type        = "ingress"
      from_port   = 80
      to_port     = 80
      protocol    = "TCP"
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]
}

resource "aws_lb_listener" "redirect_non_ssl" {
  load_balancer_arn = aws_lb.backend.arn
  port              = "80"
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

resource "aws_acm_certificate" "cert" {
  domain_name       = var.app_dns_entry
  validation_method = "DNS"

  lifecycle {
    prevent_destroy = false
  }
}

resource "aws_lb_listener" "app" {
  load_balancer_arn = aws_lb.backend.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.cert.arn

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Not Found\n"
      status_code  = "404"
    }
  }

  lifecycle {
    prevent_destroy = false
  }
}
