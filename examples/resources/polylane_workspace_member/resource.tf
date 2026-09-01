resource "polylane_workspace_member" "alice" {
  workspace_id = "ws_00000000000000000000000000000000"
  user_id      = "usr_00000000000000000000000000000000"

  position = "member"
  scopes = [
    "teams:read",
    "threads:read",
  ]
}

import {
  to = polylane_workspace_member.alice
  id = "ws_00000000000000000000000000000000/usr_00000000000000000000000000000000"
}
