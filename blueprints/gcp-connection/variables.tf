variable "project_id" {
  type        = string
  description = "Google project ID to connect."
}
variable "request_id" {
  type        = string
  description = "Polylane connection request ID."
  validation {
    condition     = can(regex("^[A-Za-z0-9_-]+$", var.request_id))
    error_message = "Use the request ID returned by Polylane."
  }
}
variable "subject" {
  type        = string
  description = "Immutable Polylane federation subject."
  validation {
    condition     = can(regex("^[A-Za-z0-9_:-]+$", var.subject))
    error_message = "Use the subject returned by Polylane."
  }
}
variable "issuer_url" {
  type        = string
  description = "Public HTTPS OIDC issuer returned by Polylane."
}
variable "push_endpoint" {
  type        = string
  description = "HTTPS push endpoint and OIDC audience returned by Polylane."
}
variable "resource_prefix" {
  type        = string
  description = "Google resource prefix returned by Polylane."
  validation {
    condition     = can(regex("^polylane-[a-z0-9]{16}$", var.resource_prefix))
    error_message = "Use the resource prefix returned by Polylane."
  }
}
variable "installer_member" {
  type        = string
  description = "IAM member running Terraform, for example serviceAccount:installer@project.iam.gserviceaccount.com. Receives actAs only on the delivery identity."
  validation {
    condition     = can(regex("^(serviceAccount|user):[^ ]+@[^ ]+$", var.installer_member))
    error_message = "Provide the user: or serviceAccount: IAM member executing this deployment."
  }
}
