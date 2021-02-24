
variable "region" { default = "us-east-2"}

variable "app_dns_entry" {
  default = "app.impartwealth.com"
}

variable "target_groups" {
  type = map(object({api_key = string, priority = number}))
  default  = {
    "dev" = {api_key = "AAs8wLBVd41EEO7Qws25ocutQAjuzwz5MM1nNNLa", priority = 40},
    "iosdev" = {api_key = "yCwm0JHpIR49GLTG8pqnd6lmTo10Cw2b5gr9qGNM", priority = 30},
    "preprod" ={api_key = "K39z2qMKV959GdI7sWpczbnhyiw4Zno6RCVXh233", priority = 20},
    "prod" = {api_key = "I1TuBFDPdh5vRYdqqIRDn7OqITyyPIQO3SQnemuS", priority = 10},
  }
}

