// creates a target group for each environment in the target groups variable
resource "aws_lb_target_group" "map" {
  for_each = var.target_groups
  name     = "backend-${each.key}"
  vpc_id   = data.aws_vpc.impart-backend.id
  port = 8080
  protocol = "HTTP"
  target_type = "instance"
  health_check {
    enabled = true
    interval = 60
    port = "traffic-port"
    path = "/ping"
    protocol = "HTTP"
    timeout = 5
    healthy_threshold = 2
    unhealthy_threshold = 3
    matcher = "200"
  }
}

locals {

}

// creates a listener rule to allow traffic only if it has the correct api key
resource "aws_lb_listener_rule" "map" {
  for_each      = var.target_groups
  listener_arn = aws_lb_listener.app.arn
  priority     = each.value.priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.map[each.key].arn
  }

  condition {
    path_pattern {
      values = concat(["/${each.key}/v1/*"], each.key == "prod" ? ["/v1/*"] : [])
    }
  }

  condition {
    http_header {
      http_header_name = "x-api-key"
      values           = [each.value.api_key]
    }
  }
  depends_on = [aws_lb_target_group.map]
}

