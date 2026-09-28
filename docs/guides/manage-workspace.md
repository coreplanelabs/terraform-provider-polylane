---
page_title: "Manage an existing workspace"
---

# Manage an existing workspace

Use import to bring an existing Polylane workspace under Terraform management, then add only the properties you want Terraform to control. The provider does not create workspaces.

## Import without changing settings

First configure the provider and [authenticate](authentication.md) with a key that has `workspaces:read`. Add a workspace resource and an import block:

```terraform
resource "polylane_workspace" "current" {}

import {
  to = polylane_workspace.current
  id = "ws_00000000000000000000000000000000"
}
```

Replace the example ID with your workspace ID. Run:

```shell
terraform init
terraform plan
terraform apply
```

Check that the plan shows the expected workspace import before approving the apply. The empty resource block leaves existing workspace properties unchanged.

## Manage selected properties

After importing, add the properties you want to manage. Ensure the API key and its owning member have the required write permissions before applying updates.

```terraform
resource "polylane_workspace" "current" {
  description = "Production operations for the platform team."
}
```

Run `terraform plan` again to see the proposed change. See the [workspace reference](../resources/workspace.md) for all supported properties.

## Add teams and settings

Reference the imported workspace ID from other resources:

```terraform
resource "polylane_team" "platform" {
  workspace_id = polylane_workspace.current.id
  name         = "Platform"
  description  = "Owns shared infrastructure and developer tooling."
}
```

Teams support creation, updates, import, and deletion. Existing workspace members are adopted by import instead; see [workspace members](../resources/workspace_member.md).

Workspace settings use separate resources so you can manage only the features you need. Each settings reference includes its configuration and import syntax.

## Understand removal

Removing the workspace, an adopted workspace member, or a workspace settings resource from configuration removes it from Terraform state on apply. It does not delete the workspace or member, or reset settings in Polylane.

Teams and team memberships are different: removing them from configuration deletes the corresponding object or membership in Polylane. Always review the plan before applying a removal.
