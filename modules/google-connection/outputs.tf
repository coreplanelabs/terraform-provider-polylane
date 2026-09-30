output "project_number" {
  description = "Project number used to activate the verified Polylane connection."
  value       = data.google_project.connection.number
}
output "reader_email" {
  description = "Read-only runtime service account."
  value       = google_service_account.reader.email
}
output "push_subject" {
  description = "Immutable delivery service-account ID. Activation independently verifies it."
  value       = google_service_account.push.unique_id
}
