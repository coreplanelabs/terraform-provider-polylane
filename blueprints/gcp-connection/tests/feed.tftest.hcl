mock_provider "google" {
  mock_data "google_project" {
    defaults = { number = "123456789012" }
  }
  mock_resource "google_service_account" {
    defaults = {
      member = "serviceAccount:reader@polylane-feed-test.iam.gserviceaccount.com"
      name   = "projects/polylane-feed-test/serviceAccounts/reader@polylane-feed-test.iam.gserviceaccount.com"
      email  = "reader@polylane-feed-test.iam.gserviceaccount.com"
    }
  }
  mock_resource "google_logging_project_sink" {
    defaults = {
      writer_identity = "serviceAccount:service-123456789012@gcp-sa-logging.iam.gserviceaccount.com"
    }
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

run "broad_resource_feed" {
  command = plan

  module {
    source = "../../modules/google-connection"
  }

  assert {
    condition = google_cloud_asset_project_feed.connection.asset_types != null && alltrue([
      for asset_type in [
        "pubsub.googleapis.com/Topic",
        "storage.googleapis.com/Bucket",
        "compute.googleapis.com/Instance",
        "run.googleapis.com/Service",
        "container.googleapis.com/Cluster",
        "sqladmin.googleapis.com/Instance",
        "config.googleapis.com/Preview",
        ] : anytrue([
          for selector in coalesce(google_cloud_asset_project_feed.connection.asset_types, []) : can(regex(selector, asset_type))
      ])
    ])
    error_message = "The RESOURCE feed must explicitly select every supported resource family, including generic inventory assets."
  }

  assert {
    condition     = google_cloud_asset_project_feed.connection.content_type == "RESOURCE"
    error_message = "The broad feed must carry resource observations."
  }

  assert {
    condition     = length(google_cloud_asset_project_feed.connection.condition) == 0
    error_message = "The feed must deliver creation, updates and deletion observations."
  }
}
