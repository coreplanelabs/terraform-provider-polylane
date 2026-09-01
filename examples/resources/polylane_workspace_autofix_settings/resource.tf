resource "polylane_workspace_autofix_settings" "current" {
  workspace_id = "ws_00000000000000000000000000000000"
  enabled      = true
  trigger_mode = "approval"
}
