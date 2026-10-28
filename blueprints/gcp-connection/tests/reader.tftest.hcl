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
run "reader_configuration_reads" {
  command = plan
  module { source = "../../modules/google-connection" }
  assert {
    condition = alltrue([for role in ["roles/iam.securityReviewer", "roles/compute.viewer", "roles/run.viewer", "roles/cloudsql.viewer", "roles/cloudfunctions.viewer", "roles/spanner.viewer", "roles/redis.viewer", "roles/cloudbuild.builds.viewer", "roles/secretmanager.viewer"] :
    contains(keys(google_project_iam_member.reader), role)])
    error_message = "The reader can read the configuration of every service the agent inspects."
  }
  assert {
    condition = length(setintersection(keys(google_project_iam_member.reader), [
      "roles/viewer", "roles/editor", "roles/owner", "roles/storage.objectViewer", "roles/bigquery.dataViewer",
      "roles/secretmanager.secretAccessor", "roles/artifactregistry.reader", "roles/alloydb.viewer", "roles/cloudtasks.viewer", "roles/composer.viewer",
    ])) == 0
    error_message = "The reader receives no basic role and no role that reads data."
  }
  assert {
    condition     = alltrue([for permission in ["artifactregistry.dockerimages.list", "alloydb.instances.list"] : contains(google_project_iam_custom_role.metadata.permissions, permission)]) && !contains(google_project_iam_custom_role.metadata.permissions, "artifactregistry.repositories.downloadArtifacts")
    error_message = "Artifact Registry and AlloyDB reads stay metadata-only."
  }
}
run "resource_names" {
  command = plan
  module { source = "../../modules/google-connection" }
  assert {
    condition = alltrue([for name in [google_service_account.reader.display_name, google_service_account.push.display_name, google_iam_workload_identity_pool.connection.display_name] : !strcontains(lower(name), "beta")]) && alltrue([
    for topic in google_pubsub_topic.events : topic.labels.purpose == "polylane-connection"])
    error_message = "Customer-visible connection resources use their release names."
  }
}
