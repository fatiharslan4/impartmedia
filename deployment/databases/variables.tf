variable "region" { default = "us-east-2" }

variable "cidr_block" {
  default = "10.0.0.0/16"
}

variable "deployments" {
  type = map(object({
    environment           = string,
    admin_password        = string,
    deletion_protection   = bool,
    retention_period_days = number,
    backtrack_window_hours = number,
    min_acu = number,
    max_acu = number,
    apply_method = string,
  }))
  default = {
    "dev" = {
      environment           = "dev",
      admin_password        = "1pjj82aRrkyFMYnmUZgRfBdLrhb1pjj7gqIJe",
      deletion_protection   = false,
      retention_period_days = 1,
      min_acu = 1,
      max_acu = 1,
      apply_method = "immediate"
      backtrack_window_hours = 0
    },
    "iosdev" = {
      environment           = "iosdev",
      admin_password        = "1pmNfyLFbbGwjrIV0tYJiUWE9Ql1pmNfy1noX6wC",
      deletion_protection   = false,
      retention_period_days = 1,
      min_acu = 1,
      max_acu = 1,
      apply_method = "immediate"
      backtrack_window_hours = 0
    },
    "preprod" = {
      environment           = "preprod",
      admin_password        = "1pmNjs2QQ5lAqqeYZC7Si2GziNz1pmNjpZCoxEHB",
      deletion_protection   = false,
      retention_period_days = 1,
      min_acu = 1,
      max_acu = 1,
      apply_method = "immediate"
      backtrack_window_hours = 0
    },
    "prod" = {
      environment           = "prod",
      admin_password        = "1pmNkvdxyXoBEF12zsR7R6iB0qW1pmNkpYoK4ock",
      deletion_protection   = false,
      retention_period_days = 1,
      min_acu = 1,
      max_acu = 1,
      apply_method = "immediate"
      backtrack_window_hours = 0
    },
  }
}