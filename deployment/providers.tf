terraform {
  backend "s3" {
    key                  = "tfstate/backend-api"
    dynamodb_table       = "tflock"
    workspace_key_prefix = "tfenv"
    bucket               = "impart-wealth-us-east-2"
    region               = "us-east-2"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.region
}