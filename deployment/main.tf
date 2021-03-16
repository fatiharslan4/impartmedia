locals {
  enabled_deployments = tomap({ for k, v in var.deployments :  k => v if v.enabled == true })
}


data "aws_ecs_cluster" "backend" {
  cluster_name = "impart-default-backend"
}

data "aws_lb_target_group" "all" {
  for_each = local.enabled_deployments
  name     = "backend-${each.key}"
}

resource "aws_cloudwatch_log_group" "all" {
  for_each = local.enabled_deployments
  name     = "impart/backend/${each.key}"
  retention_in_days = 90
}

resource "aws_iam_role" "all" {
  for_each = local.enabled_deployments
  name     = "task-role-backend-${each.key}"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "dynamo" {
  for_each = local.enabled_deployments
  name     = "dynamo-task-backend-${each.key}"
  role = aws_iam_role.all[each.key].id

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem",
          "dynamodb:UpdateTimeToLive",
          "dynamodb:PutItem",
          "dynamodb:ListTables",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
          "dynamodb:Query",
          "dynamodb:UpdateItem",
          "dynamodb:ListGlobalTables",
          "dynamodb:DescribeTable",
          "dynamodb:GetItem",
          "dynamodb:UpdateTable"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:dynamodb:*:*:table/${each.key}*"
      },
      {
        Action = [
          "ec2:AuthorizeSecurityGroupIngress",
          "ec2:Describe*",
          "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
          "elasticloadbalancing:DeregisterTargets",
          "elasticloadbalancing:Describe*",
          "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
          "elasticloadbalancing:RegisterTargets"
        ]
        Effect = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "ec2:DescribeTags",
          "ecs:CreateCluster",
          "ecs:DeregisterContainerInstance",
          "ecs:DiscoverPollEndpoint",
          "ecs:Poll",
          "ecs:RegisterContainerInstance",
          "ecs:StartTelemetrySession",
          "ecs:UpdateContainerInstancesState",
          "ecs:Submit*",
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "ecs:DescribeServices",
          "ecs:UpdateService",
          "cloudwatch:DescribeAlarms",
          "cloudwatch:PutMetricAlarm"
        ]
        Effect = "Allow"
        Resource = "*"
      }
    ]
  })
}

resource "aws_ecs_service" "all" {
  for_each        = local.enabled_deployments
  name            = "backend-${each.key}"
  cluster         = data.aws_ecs_cluster.backend.id
  task_definition = aws_ecs_task_definition.all[each.key].arn
  desired_count   = each.value.desired_count
  launch_type = "EC2"
  scheduling_strategy = "REPLICA"
  ordered_placement_strategy {
    type  = "spread"
    field = "instanceId"
  }

  load_balancer {
    target_group_arn = data.aws_lb_target_group.all[each.key].arn
    container_name   = "api"
    container_port   = var.container_port
  }
  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [${var.region}a, ${var.region}b]"
  }
  depends_on = [data.aws_lb_target_group.all, aws_iam_role.all]
}

resource "aws_ecs_task_definition" "all" {
  for_each = local.enabled_deployments
  family   = "service-${each.key}"
  task_role_arn        = aws_iam_role.all[each.key].arn
  network_mode = "bridge"
  container_definitions = templatefile("./taskdef.json",
    {
      env    = each.key
      region = var.region
      image  = "${var.image_repo}:${each.value.image_tag}"
      env_variables = join(",", flatten([
        for obj in each.value.environment_variables : format("{\"name\":\"IMPART_%s\", \"value\":\"%s\"}", obj.key, obj.value)
      ])),
      log_group = aws_cloudwatch_log_group.all[each.key].name
      container_port = var.container_port
    })
  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [${var.region}a, ${var.region}b]"
  }
}
