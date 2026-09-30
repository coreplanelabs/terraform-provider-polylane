variable "workspace_id" {
  type        = string
  description = "Polylane workspace ID."
}
variable "project_id" {
  type        = string
  description = "Google project ID to connect."
}
variable "installer_member" {
  type        = string
  description = "IAM user: or serviceAccount: member executing Terraform."
}
