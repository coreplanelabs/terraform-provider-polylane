resource "google_project_service" "required" {
  for_each = toset([
    "cloudasset.googleapis.com", "cloudresourcemanager.googleapis.com", "cloudtrace.googleapis.com",
    "iam.googleapis.com", "iamcredentials.googleapis.com", "logging.googleapis.com",
    "monitoring.googleapis.com", "pubsub.googleapis.com", "sts.googleapis.com",
    "storage.googleapis.com", "container.googleapis.com"
  ])
  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

resource "google_project_service_identity" "pubsub" {
  provider   = google-beta
  project    = var.project_id
  service    = "pubsub.googleapis.com"
  depends_on = [google_project_service.required]
}

resource "google_project_service_identity" "cloudasset" {
  provider   = google-beta
  project    = var.project_id
  service    = "cloudasset.googleapis.com"
  depends_on = [google_project_service.required]
}

resource "google_project_iam_custom_role" "metadata" {
  project = var.project_id
  role_id = replace("${var.resource_prefix}_metadata", "-", "_")
  title   = "Polylane connection metadata"
  permissions = [
    "iam.serviceAccounts.get", "iam.workloadIdentityPoolProviders.get",
    "storage.buckets.get", "cloudasset.feeds.get", "logging.sinks.get",
    "resourcemanager.projects.get", "serviceusage.services.use",
    "cloudtrace.traces.get", "cloudtrace.traces.list"
  ]
  depends_on = [google_project_service.required]
}

resource "google_project_iam_member" "metadata" {
  project = var.project_id
  role    = google_project_iam_custom_role.metadata.name
  member  = google_service_account.reader.member
}

resource "google_project_iam_member" "gke_reader" {
  count   = var.gke_access_enabled ? 1 : 0
  project = var.project_id
  role    = "roles/container.viewer"
  member  = google_service_account.reader.member
}

resource "google_project_iam_custom_role" "gke_access" {
  count   = var.gke_access_enabled ? 1 : 0
  project = var.project_id
  role_id = replace("${var.resource_prefix}_gke", "-", "_")
  title   = "Polylane Kubernetes access"
  permissions = concat(
    ["container.pods.getLogs"],
    var.gke_operations_enabled ? ["container.nodes.update"] : []
  )
  depends_on = [google_project_service.required]
}

resource "google_project_iam_member" "gke_access" {
  count   = var.gke_access_enabled ? 1 : 0
  project = var.project_id
  role    = google_project_iam_custom_role.gke_access[0].name
  member  = google_service_account.reader.member
}

resource "google_project_iam_member" "gke_operator" {
  count   = var.gke_access_enabled && var.gke_operations_enabled ? 1 : 0
  project = var.project_id
  role    = "roles/container.developer"
  member  = google_service_account.reader.member
}
