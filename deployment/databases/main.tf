locals {
  private_cidr_block = cidrsubnet(var.cidr_block, 1, 1)
}


module "rds_cluster_aurora_mysql" {
  for_each = var.deployments
  source              = "cloudposse/rds-cluster/aws"
  version             = "0.44.1"
  namespace           = "impart"
  stage               = each.value.environment
  name                = "mysql"
  engine              = "aurora-mysql"
  engine_mode         = "serverless"
  engine_version      = "5.7.mysql_aurora.2.07.1"
  cluster_family      = "aurora-mysql5.7"
  publicly_accessible = false
  cluster_size        = 0
  admin_user          = "impart_db_admin"
  admin_password      = each.value.admin_password
  db_name             = "impart"
  db_port             = 3306
  vpc_id              = "vpc-0bfb2f75636067f54"
  source_region = "us-east-2"
  instance_availability_zone = "us-east-2a"
  security_groups = ["sg-0236e8abca966e1cd"]
  subnets = [ "subnet-077bf52c6c758f53a", "subnet-0e403c24d6ddfa62d"]
  allowed_cidr_blocks = [local.private_cidr_block]
  enable_http_endpoint = true
  retention_period = each.value.retention_period_days
  backtrack_window = 3600 * each.value.backtrack_window_hours
  scaling_configuration = [
    {
      auto_pause               = true
      max_capacity             = each.value.max_acu
      min_capacity             = each.value.min_acu
      seconds_until_auto_pause = 300
      timeout_action = "ForceApplyCapacityChange"
    }
  ]

  cluster_parameters = [
    {
      name = "innodb_file_per_table"
      value = "1"
      apply_method = "pending-reboot"
    },
    {
      name  = "character_set_client"
      value = "utf8mb4"
      apply_method = "pending-reboot"
    },
    {
      name  = "character_set_connection"
      value = "utf8mb4"
      apply_method = "pending-reboot"
    },
    {
      name  = "character_set_database"
      value = "utf8mb4"
      apply_method = "pending-reboot"
    },
    {
      name  = "character_set_results"
      value = "utf8mb4"
      apply_method = "pending-reboot"
    },
    {
      name  = "character_set_server"
      value = "utf8mb4"
      apply_method = "pending-reboot"
    },
    {
      name  = "collation_connection"
      value = "utf8mb4_unicode_ci"
      apply_method = "pending-reboot"
    },
    {
      name  = "collation_server"
      value = "utf8mb4_unicode_ci"
      apply_method = "pending-reboot"
    },
    {
      name         = "lower_case_table_names"
      value        = "1"
      apply_method = "pending-reboot"
    },
    {
      name         = "skip-character-set-client-handshake"
      value        = "1"
      apply_method = "pending-reboot"
    }
  ]
}