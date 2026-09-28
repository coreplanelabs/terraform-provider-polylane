terraform {
  required_version = ">= 1.6.0"

  required_providers {
    polylane = {
      source  = "registry.terraform.io/coreplanelabs/polylane"
      version = "~> 0.1.0"
    }
  }
}

provider "polylane" {
  # Prefer POLYLANE_API_KEY so the key is not written in configuration.
}
