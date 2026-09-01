terraform {
  required_version = ">= 1.6.0"

  required_providers {
    polylane = {
      source  = "registry.terraform.io/coreplanelabs/polylane"
      version = ">= 0.1.0"
    }
  }
}

provider "polylane" {
  api_key = var.polylane_api_key
}

variable "polylane_api_key" {
  type      = string
  sensitive = true
}
