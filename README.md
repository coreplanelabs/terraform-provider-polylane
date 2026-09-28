# Terraform Provider for Polylane

This repository contains the official Terraform (and OpenTofu) provider for
Polylane. The initial provider intentionally covers a small workspace-level
surface:

- adopting an existing workspace;
- adopting existing workspace members and managing their roles and scopes;
- managing teams and team membership;
- connecting customer-managed AWS infrastructure without giving the Polylane
  provider AWS credentials or AWS resource ownership;
- managing autofix, digest, investigation, investigation-limit, model-training,
  observability, and pull-request-review settings.

Workspace creation and destructive deletion are intentionally not supported.
Integration credentials and custom model provider credentials are also outside
the initial scope so they never need to be stored in Terraform state.

Install the provider from the [Terraform Registry](https://registry.terraform.io/providers/coreplanelabs/polylane/latest/docs) as `coreplanelabs/polylane`:

```hcl
terraform {
  required_providers {
    polylane = {
      source  = "coreplanelabs/polylane"
      version = "~> 0.1.0"
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

### Customer-managed AWS connections

AWS onboarding is deliberately split around the customer's AWS resources. A
`polylane_aws_connection_request` first returns the Polylane-issued STS external
ID, trusted principal, and SNS subscription endpoint. The customer uses those
values in ordinary resources from the `hashicorp/aws` provider, then passes the
resulting identifiers to `polylane_aws_connection` for validation and
registration.

```hcl
resource "polylane_aws_connection_request" "this" {
  workspace_id = polylane_workspace.current.id
  account_id   = data.aws_caller_identity.current.account_id
  regions      = ["us-east-1"]
}

# aws_iam_role, aws_s3_bucket, aws_sns_topic,
# aws_sns_topic_subscription, and aws_cloudtrail are managed by the customer.

resource "polylane_aws_connection" "this" {
  workspace_id           = polylane_workspace.current.id
  request_id             = polylane_aws_connection_request.this.id
  region                 = polylane_aws_connection_request.this.region
  role_arn               = aws_iam_role.polylane.arn
  bucket_name            = aws_s3_bucket.cloudtrail.id
  topic_arn              = aws_sns_topic.cloudtrail.arn
  topic_subscription_arn = aws_sns_topic_subscription.polylane.arn
  cloudtrail_name        = aws_cloudtrail.polylane.name
}
```

The Polylane provider never reads AWS credentials and never creates, updates,
or deletes these AWS resources. On destroy it deregisters the Polylane
connection; Terraform's dependency graph then leaves AWS teardown to the AWS
provider and the customer's own lifecycle rules.

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

The AWS connection-request acceptance test creates, reads, and deletes only a
temporary Polylane-side handshake request; it never creates AWS resources. It
only runs when `POLYLANE_AWS_CONNECTION_REQUEST_ACCEPTANCE=1` and the API key
has `cloud_accounts:read`, `cloud_accounts:write`, and
`cloud_accounts:delete` scopes.

Pull requests run a secret-free provider protocol smoke test. Pushes to `main`,
scheduled runs, and manually dispatched runs load `POLYLANE_API_KEY`,
`POLYLANE_ENDPOINT`, and `POLYLANE_WORKSPACE_ID` from the dedicated 1Password
Environment named `Terraform Provider`. Configure these settings in the
`acceptance` GitHub environment, which permits only the `main` branch:

- Secret `OP_TERRAFORM_PROVIDER_SERVICE_ACCOUNT_TOKEN`: a dedicated 1Password
  service-account token with read-only access to that Environment and any vault
  items it references.
- Variable `OP_TERRAFORM_PROVIDER_ENVIRONMENT_ID`: the UUID copied from the
  Environment's **Manage environment** page.

The live CI job enables the AWS connection-request lifecycle automatically, so
the Environment's UAT API key needs `workspaces:read`, `workspace_members:read`,
`cloud_accounts:read`, `cloud_accounts:write`, and `cloud_accounts:delete`.
`POLYLANE_TEAM_ACCEPTANCE=1` may be added to the Environment when its API key
and owning member both have `teams:read`, `teams:write`, and `teams:delete`.

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

Release Please maintains a version PR from Conventional Commits. Merging it
creates a tag and draft release. GoReleaser builds and signs the provider,
verifies the uploaded artifacts, then publishes the complete release for the
Terraform Registry. The first version is `0.1.0`.

See [RELEASING.md](RELEASING.md) for credential setup, Registry enrollment,
release verification, and recovery. `task release:check` exercises the complete
archive build without signing or publishing.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contributions and
[SECURITY.md](SECURITY.md) for private vulnerability reports.
