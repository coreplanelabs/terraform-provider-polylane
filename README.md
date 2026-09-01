# Terraform Provider for Polylane

This repository contains the official Terraform (and OpenTofu) provider for
Polylane. The initial provider intentionally covers a small workspace-level
surface:

- adopting an existing workspace;
- adopting existing workspace members and managing their roles and scopes;
- managing teams and team membership;
- managing autofix, digest, investigation, investigation-limit, model-training,
  observability, and pull-request-review settings.

Workspace creation and destructive deletion are intentionally not supported.
Integration credentials and custom model provider credentials are also outside
the initial scope so they never need to be stored in Terraform state.

Once published, the provider is sourced as `coreplanelabs/polylane`:

```hcl
terraform {
  required_providers {
    polylane = {
      source  = "coreplanelabs/polylane"
      version = ">= 0.1.0"
    }
  }
}
```

Configure authentication with `POLYLANE_API_KEY`, then adopt a workspace with an
import block:

```hcl
provider "polylane" {}

resource "polylane_workspace" "current" {}

import {
  to = polylane_workspace.current
  id = "ws_00000000000000000000000000000000"
}
```

Removing an adopted workspace, workspace member, or workspace settings resource
from Terraform configuration stops Terraform management. It does not delete or
reset the adopted object in Polylane. Teams and team memberships are normal
lifecycle resources and are deleted when removed from configuration.

Generated provider documentation lives in [docs/](docs/) and is published on the
Terraform Registry.

## Development

```sh
source bin/activate-hermit
task init
task do
```

### API client generation

The provider checks in the production OpenAPI document and a generated Go
client for only the operations the provider supports. The operation allowlist
in `api-client-config.yaml` prevents unrelated API endpoints from silently
expanding the provider surface.

```sh
task generate       # regenerate from the checked-in normalized spec
task generate:check # fail if generated code has drifted
task sync:openapi   # download production /v1/doc, normalize, generate, and test
```

The model-training settings route is deliberately the sole handwritten client
exception because that route is not published in the OpenAPI document.

Provider configuration can come from Terraform configuration or environment variables:

- `POLYLANE_API_KEY`
- `POLYLANE_ENDPOINT`

### Acceptance Tests

Acceptance tests use the Terraform provider test harness and the read-only
workspace import path. Set `POLYLANE_API_KEY`, `POLYLANE_ENDPOINT`, and
`POLYLANE_WORKSPACE_ID` in an ignored `.env` file or export them in your shell.

```sh
task test:acc
```

The team lifecycle acceptance test creates, updates, imports, and deletes a
temporary team. It only runs when `POLYLANE_TEAM_ACCEPTANCE=1` and the API key
has `teams:read`, `teams:write`, and `teams:delete` scopes.

## Documentation

Registry docs under [docs/](docs/) are generated from the provider schema and the
files in [examples/](examples/) with
[`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs):

```sh
task docs           # regenerate docs/ and validate them
task docs:validate  # validate only
```

`task do` and CI regenerate and validate docs, so edit the schema/examples rather
than the generated Markdown.

## Releasing & Publishing

Releases are automated:

1. [release-please](https://github.com/googleapis/release-please) opens a release
   PR from Conventional Commits. Merging it tags `vX.Y.Z` and creates a GitHub
   release.
2. The `goreleaser` job in [release.yaml](.github/workflows/release.yaml) then
   builds the provider for all target platforms and attaches the artifacts the
   Terraform Registry ingests: the per-platform zips, `*_SHA256SUMS`,
   `*_SHA256SUMS.sig` (GPG detached signature), and `*_manifest.json`.

### One-time setup to publish to the Terraform Registry

1. Make this repository **public**.
2. Generate a GPG signing key (RSA/DSA, not ECC), then:
   - Add the **public** key at <https://registry.terraform.io> → *User Settings →
     Signing Keys*.
   - Add the **private** key and its passphrase as repository secrets
     `GPG_PRIVATE_KEY` (ASCII-armored) and `PASSPHRASE`.
3. On the registry, *Publish → Provider* and select this repo. The registry adds a
   release webhook and ingests each finalized release automatically.

Provider metadata for the registry lives in
[`terraform-registry-manifest.json`](terraform-registry-manifest.json)
(`protocol_versions: ["6.0"]`, terraform-plugin-framework).

Preview the rendered registry pages at
<https://registry.terraform.io/tools/doc-preview>.
