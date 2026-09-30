terraform {
  required_version = ">= 1.5, < 2.0"
  required_providers {
    polylane = {
      source = "registry.terraform.io/coreplanelabs/polylane"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 7.0"
    }
  }
}

resource "polylane_gcp_connection_request" "this" {
  workspace_id = var.workspace_id
  project_id   = var.project_id
}

module "google_connection" {
  source           = "../google-connection"
  project_id       = var.project_id
  request_id       = polylane_gcp_connection_request.this.id
  subject          = polylane_gcp_connection_request.this.subject
  issuer_url       = polylane_gcp_connection_request.this.issuer_url
  push_endpoint    = polylane_gcp_connection_request.this.push_endpoint
  resource_prefix  = polylane_gcp_connection_request.this.resource_prefix
  installer_member = var.installer_member
}

resource "polylane_gcp_connection" "this" {
  workspace_id   = var.workspace_id
  request_id     = polylane_gcp_connection_request.this.id
  project_number = module.google_connection.project_number
  depends_on     = [module.google_connection]
}
