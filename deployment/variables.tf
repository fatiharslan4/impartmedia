variable "region" {
  default = "us-east-2"
}

variable "image_repo" {
  default = "518740895671.dkr.ecr.us-east-2.amazonaws.com/impartwealth/backend"
}

variable "container_port" {
  default = 8080
}
variable "deployments" {
  type = map(object({
    enabled               = bool,
    api_key               = string,
    image_tag             = string,
    desired_count         = number,
    environment_variables = list(object({ key = string, value = string })),
  }))
  default = {
    "dev" = {
      enabled = true,
      desired_count = 4,
      api_key = "AAs8wLBVd41EEO7Qws25ocutQAjuzwz5MM1nNNLa",
      image_tag = "285801f",
      environment_variables = [
        {
          key = "ENV",
          value = "dev"
        },
        {
          key = "DEBUG",
          value = "true"
        },
        {
          key = "PORT",
          value = "8080"
        },
        {
          key = "API_KEY",
          value = "AAs8wLBVd41EEO7Qws25ocutQAjuzwz5MM1nNNLa"
        },
        {
          key = "REGION",
          value = "us-east-2"
        },
        {
          key = "DYNAMO_ENDPOINT",
          value = "dynamodb.us-east-2.amazonaws.com"
        },
        {
          key = "IOS_NOTIFICATION_ARN",
          value = ""
        },
        {
          key = "PROFILE_SCHEMA_PATH",
          value = "./schemas/json/Profile.json"
        }
      ]
    },
    "iosdev" = {
      enabled = true,
      desired_count = 4,
      api_key = "yCwm0JHpIR49GLTG8pqnd6lmTo10Cw2b5gr9qGNM",
      image_tag = "285801f",
      environment_variables = [
        {
          key = "ENV",
          value = "iosdev"
        },
        {
          key = "DEBUG",
          value = "false"
        },
        {
          key = "PORT",
          value = "8080"
        },
        {
          key = "API_KEY",
          value = "yCwm0JHpIR49GLTG8pqnd6lmTo10Cw2b5gr9qGNM"
        },
        {
          key = "REGION",
          value = "us-east-2"
        },
        {
          key = "DYNAMO_ENDPOINT",
          value = "dynamodb.us-east-2.amazonaws.com"
        },
        {
          key = "IOS_NOTIFICATION_ARN",
          value = ""
        },
        {
          key = "PROFILE_SCHEMA_PATH",
          value = "./schemas/json/Profile.json"
        }
      ]
    },
    "preprod" = {
      enabled = true,
      desired_count = 4,
      api_key = "K39z2qMKV959GdI7sWpczbnhyiw4Zno6RCVXh233",
      image_tag = "285801f",
      environment_variables = [
        {
          key = "ENV",
          value = "preprod"
        },
        {
          key = "DEBUG",
          value = "false"
        },
        {
          key = "PORT",
          value = "8080"
        },
        {
          key = "API_KEY",
          value = "K39z2qMKV959GdI7sWpczbnhyiw4Zno6RCVXh233"
        },
        {
          key = "REGION",
          value = "us-east-2"
        },
        {
          key = "DYNAMO_ENDPOINT",
          value = "dynamodb.us-east-2.amazonaws.com"
        },
        {
          key = "IOS_NOTIFICATION_ARN",
          value = ""
        },
        {
          key = "PROFILE_SCHEMA_PATH",
          value = "./schemas/json/Profile.json"
        }
      ]
    },
    "prod" = {
      enabled = true,
      desired_count = 4,
      api_key = "I1TuBFDPdh5vRYdqqIRDn7OqITyyPIQO3SQnemuS",
      image_tag = "285801f",
      environment_variables = [
        {
          key = "ENV",
          value = "prod"
        },
        {
          key = "DEBUG",
          value = "false"
        },
        {
          key = "PORT",
          value = "8080"
        },
        {
          key = "API_KEY",
          value = "I1TuBFDPdh5vRYdqqIRDn7OqITyyPIQO3SQnemuS"
        },
        {
          key = "REGION",
          value = "us-east-2"
        },
        {
          key = "DYNAMO_ENDPOINT",
          value = "dynamodb.us-east-2.amazonaws.com"
        },
        {
          key = "IOS_NOTIFICATION_ARN",
          value = ""
        },
        {
          key = "PROFILE_SCHEMA_PATH",
          value = "./schemas/json/Profile.json"
        }
      ]
    },
  }

}
