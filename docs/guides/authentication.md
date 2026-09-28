---
page_title: "Authentication"
---

# Authentication

The Polylane provider authenticates with a workspace API key. Create the key in the [Polylane console](https://console.polylane.com) under your workspace's **Settings → API keys**. Use a dedicated key for Terraform and select the scopes required by your configuration.

## Supply credentials

Set `POLYLANE_API_KEY` in the environment where Terraform runs. For CI, store it in your secret manager and inject it into the Terraform process. Keep the provider configuration simple:

```terraform
provider "polylane" {}
```

For an interactive Bash or Zsh session, this prompts for the key without putting its value in shell history:

```shell
read -rs POLYLANE_API_KEY
export POLYLANE_API_KEY
```

Paste the key, press Enter, and run your Terraform commands in that shell. Run `unset POLYLANE_API_KEY` when finished.

The optional `api_key` provider argument overrides `POLYLANE_API_KEY`. Avoid putting a literal key in a `.tf` file or committing credentials to version control. Terraform's sensitive-value marking redacts normal output; it does not make state or saved plans safe to share.

## Choose scopes

Use these scopes for the operations your configuration needs. The key's owning member must also have permission to perform those operations.

| Operation | API key scopes |
| --- | --- |
| Import and read a workspace | `workspaces:read` |
| Import and read existing workspace members | `workspace_members:read` |
| Create, read, update, and delete teams | `teams:read`, `teams:write`, `teams:delete` |
| Create, read, and delete AWS connection requests; manage AWS connections | `cloud_accounts:read`, `cloud_accounts:write`, `cloud_accounts:delete` |

This table covers common starting points. When managing member permissions or workspace settings, select the additional scopes for those operations in the API key creation screen. A read-only key is useful for evaluating imports and plans before granting write access.

## Select an API environment

The default endpoint is `https://api.polylane.com/v1`. Leave it unset for the standard Polylane service. To target another environment, set `POLYLANE_ENDPOINT` to that environment's API base URL, including `/v1`, or configure `endpoint` in the provider block.

The API key, workspace ID, and endpoint must refer to the same environment. Use the API hostname, not the console URL.

## Troubleshoot access

- **Missing API key:** confirm `POLYLANE_API_KEY` is available to the process running Terraform, including remote or CI runs.
- **Unauthorized or forbidden:** check key validity, selected scopes, and the owning member's permissions.
- **Workspace not found:** check the workspace ID, the key's workspace, and the API endpoint together.

Start with `terraform plan` and review changes before applying them. Use a separate workspace for automated tests that create and delete resources.
