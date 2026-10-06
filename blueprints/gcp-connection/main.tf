terraform {
  required_version = ">= 1.5.7, < 2.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "= 7.0.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "= 7.0.0"
    }
  }
}

provider "google" { project = var.project_id }
provider "google-beta" { project = var.project_id }

module "connection" {
  source           = "../../modules/google-connection"
  project_id       = var.project_id
  request_id       = var.request_id
  subject          = var.subject
  issuer_url       = var.issuer_url
  push_endpoint    = var.push_endpoint
  resource_prefix  = var.resource_prefix
  installer_member = var.installer_member
}

output "project_number" { value = module.connection.project_number }
