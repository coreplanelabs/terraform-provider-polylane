resource "polylane_workspace" "current" {
  # Workspaces are adopted by import. Add name, slug, description, or
  # auto_join only when Terraform should manage that property.
}

import {
  to = polylane_workspace.current
  id = "ws_00000000000000000000000000000000"
}
