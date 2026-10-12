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

variable "gke_access_enabled" {
  type        = bool
  default     = true
  description = "Grant the runtime identity GKE object reads and pod logs in this project. Cluster endpoints must permit access."
}

variable "gke_operations_enabled" {
  type        = bool
  default     = false
  description = "Grant the runtime identity Kubernetes Engine Developer and node update permissions for approved workload operations."
}
