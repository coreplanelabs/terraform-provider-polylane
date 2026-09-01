resource "polylane_workspace_investigation_limits_settings" "current" {
  workspace_id                      = "ws_00000000000000000000000000000000"
  minimum_auto_investigate_severity = "high"

  # Omit per_24h for no workspace-specific limit.
}
