// creates a target group for each environment in the target groups variable
resource "aws_lb_target_group" "map" {
  for_each      = var.target_groups
  name     = "backend-${each.key}"
  port     = 80
  protocol = "HTTP"
  vpc_id   = data.aws_vpc.impart-backend.id
  health_check {
    enabled = true
    interval = 20
    path = "/ping"
    protocol = "HTTP"
    timeout = 5
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200"
  }
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
      values = ["/${each.key}/*"]
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

