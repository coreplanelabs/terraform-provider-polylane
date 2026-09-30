resource "polylane_gcp_connection" "this" {
  workspace_id   = "ws_00000000000000000000000000000000"
  request_id     = "gcp_00000000000000000000000000000000"
  project_number = "123456789012"
  # Use modules/gcp to order activation after all Google resources and IAM grants.
}
