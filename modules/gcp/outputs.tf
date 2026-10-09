output "cloud_account_id" {
  description = "Registered Polylane cloud-account ID. Initial synchronization is asynchronous."
  value       = polylane_gcp_connection.this.id
}
output "request_id" {
  description = "Polylane request ID used for troubleshooting and lifecycle ownership."
  value       = polylane_gcp_connection_request.this.id
}
