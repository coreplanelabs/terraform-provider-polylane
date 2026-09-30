data "google_project" "connection" { project_id = var.project_id }

resource "google_service_account" "reader" {
  depends_on   = [google_project_service.required]
  project      = var.project_id
  account_id   = "${var.resource_prefix}-read"
  display_name = "Polylane beta runtime reader"
}

resource "google_service_account" "push" {
  depends_on   = [google_project_service.required]
  project      = var.project_id
  account_id   = "${var.resource_prefix}-push"
  display_name = "Polylane beta Pub/Sub delivery"
}

resource "google_project_iam_member" "reader" {
  project  = var.project_id
  for_each = toset(["roles/cloudasset.viewer", "roles/monitoring.viewer", "roles/logging.viewer", "roles/pubsub.viewer"])
  role     = each.value
  member   = google_service_account.reader.member
}

resource "google_iam_workload_identity_pool" "connection" {
  depends_on                = [google_project_service.required]
  project                   = var.project_id
  workload_identity_pool_id = var.resource_prefix
  display_name              = "Polylane beta connection"
}

resource "google_iam_workload_identity_pool_provider" "connection" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.connection.workload_identity_pool_id
  workload_identity_pool_provider_id = "polylane"
  attribute_mapping = {
    "google.subject"          = "assertion.sub"
    "attribute.connection_id" = "assertion.connection_id"
  }
  attribute_condition = "assertion.sub == '${var.subject}' && assertion.connection_id == '${var.request_id}'"
  oidc {
    issuer_uri = var.issuer_url
  }
}

resource "google_service_account_iam_member" "federation" {
  service_account_id = google_service_account.reader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${google_iam_workload_identity_pool.connection.name}/subject/${var.subject}"
}

resource "google_service_account_iam_member" "push_token" {
  service_account_id = google_service_account.push.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_project_service_identity.pubsub.email}"
}

resource "google_service_account_iam_member" "installer_push" {
  service_account_id = google_service_account.push.name
  role               = "roles/iam.serviceAccountUser"
  member             = var.installer_member
}

resource "google_pubsub_topic" "events" {
  depends_on = [google_project_service.required]
  project    = var.project_id
  for_each   = toset(["assets", "logs", "dead-letters"])
  name       = "${var.resource_prefix}-${each.key}"
  labels     = { purpose = "polylane-gcp-beta" }
}

resource "google_pubsub_topic_iam_member" "dead_letter_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.events["dead-letters"].name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_project_service_identity.pubsub.email}"
}

resource "google_pubsub_subscription" "events" {
  project                    = var.project_id
  for_each                   = toset(["assets", "logs"])
  name                       = "${var.resource_prefix}-${each.key}"
  topic                      = google_pubsub_topic.events[each.key].id
  ack_deadline_seconds       = 30
  message_retention_duration = "86400s"
  expiration_policy { ttl = "" }
  push_config {
    push_endpoint = var.push_endpoint
    oidc_token {
      service_account_email = google_service_account.push.email
      audience              = var.push_endpoint
    }
  }
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "60s"
  }
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.events["dead-letters"].id
    max_delivery_attempts = 5
  }
  depends_on = [google_service_account_iam_member.push_token, google_service_account_iam_member.installer_push, google_pubsub_topic_iam_member.dead_letter_publisher]
}

resource "google_pubsub_subscription_iam_member" "dead_letter_subscriber" {
  project      = var.project_id
  for_each     = google_pubsub_subscription.events
  subscription = each.value.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${google_project_service_identity.pubsub.email}"
}

resource "google_pubsub_subscription" "dead_letters" {
  project                    = var.project_id
  name                       = "${var.resource_prefix}-dead-letters"
  topic                      = google_pubsub_topic.events["dead-letters"].id
  message_retention_duration = "86400s"
  expiration_policy { ttl = "" }
}

resource "google_pubsub_topic_iam_member" "asset_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.events["assets"].name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_project_service_identity.cloudasset.email}"
}

resource "google_cloud_asset_project_feed" "connection" {
  project      = var.project_id
  feed_id      = var.resource_prefix
  content_type = "RESOURCE"
  asset_types  = ["pubsub.googleapis.com/Topic", "storage.googleapis.com/Bucket"]
  feed_output_config {
    pubsub_destination { topic = google_pubsub_topic.events["assets"].id }
  }
  depends_on = [google_pubsub_topic_iam_member.asset_publisher, google_pubsub_subscription.events]
}

resource "google_logging_project_sink" "connection" {
  project                = var.project_id
  name                   = var.resource_prefix
  destination            = "pubsub.googleapis.com/${google_pubsub_topic.events["logs"].id}"
  unique_writer_identity = true
  filter                 = "(log_id(\"cloudaudit.googleapis.com/activity\") AND (protoPayload.serviceName=\"pubsub.googleapis.com\" OR protoPayload.serviceName=\"storage.googleapis.com\")) OR (severity>=ERROR AND (resource.type=\"pubsub_topic\" OR resource.type=\"gcs_bucket\"))"
  depends_on             = [google_pubsub_subscription.events]
}

resource "google_pubsub_topic_iam_member" "log_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.events["logs"].name
  role    = "roles/pubsub.publisher"
  member  = google_logging_project_sink.connection.writer_identity
}

