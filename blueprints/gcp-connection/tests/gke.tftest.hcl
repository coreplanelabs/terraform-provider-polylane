mock_provider "google" {
  mock_data "google_project" { defaults = { number = "123456789012" } }
  mock_resource "google_service_account" {
    defaults = {
      member = "serviceAccount:reader@polylane-feed-test.iam.gserviceaccount.com"
      name   = "projects/polylane-feed-test/serviceAccounts/reader@polylane-feed-test.iam.gserviceaccount.com"
      email  = "reader@polylane-feed-test.iam.gserviceaccount.com"
    }
  }
  mock_resource "google_logging_project_sink" {
    defaults = { writer_identity = "serviceAccount:service-123456789012@gcp-sa-logging.iam.gserviceaccount.com" }
  }
}
mock_provider "google-beta" {}
variables {
  project_id       = "polylane-feed-test"
  request_id       = "gcpconn_offline"
  subject          = "ws_offline:gcpconn_offline"
  issuer_url       = "https://api.example.com/gcp/oidc"
  push_endpoint    = "https://api.example.com/gcp/events"
  resource_prefix  = "polylane-0123456789abcdef"
  installer_member = "serviceAccount:installer@polylane-feed-test.iam.gserviceaccount.com"
}
run "gke_reads" {
  command = plan
  module { source = "../../modules/google-connection" }
  assert {
    condition     = google_project_iam_member.gke_reader[0].role == "roles/container.viewer" && contains(google_project_iam_custom_role.gke_access[0].permissions, "container.pods.getLogs")
    error_message = "GKE reads include object visibility and live pod logs."
  }
  assert {
    condition     = length(google_project_iam_member.gke_operator) == 0 && !contains(google_project_iam_custom_role.gke_access[0].permissions, "container.nodes.update")
    error_message = "Default access must remain read-only."
  }
}
run "gke_operations" {
  command = plan
  module { source = "../../modules/google-connection" }
  variables { gke_operations_enabled = true }
  assert {
    condition     = google_project_iam_member.gke_operator[0].role == "roles/container.developer" && contains(google_project_iam_custom_role.gke_access[0].permissions, "container.nodes.update")
    error_message = "Opting into operations grants workload and node operations."
  }
}
run "gke_disabled" {
  command = plan
  module { source = "../../modules/google-connection" }
  variables { gke_access_enabled = false }
  assert {
    condition     = length(google_project_iam_member.gke_reader) == 0 && length(google_project_iam_custom_role.gke_access) == 0 && length(google_project_iam_member.gke_operator) == 0
    error_message = "Disabled GKE access creates no Kubernetes grants."
  }
}
