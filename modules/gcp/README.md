# Google Cloud connection (Beta)

Connect one project with one module. The module uses the Polylane provider for
request/registration and the Google providers for infrastructure ownership.
It enables project inventory, Pub/Sub topic and Storage bucket details, change
and selected log delivery, and read-only investigations. GKE access and
agent-driven Google mutations are outside this beta.

## Console setup

For a prefilled, copy-and-run experience, use the Polylane console. Enter the
project ID and number up front; the console produces a 38-line script with
inline Terraform, bootstrap commands and Infrastructure Manager apply. It uses
[the Google-only child](../google-connection/README.md) and completes activation
through the signed-in console. It requires no Polylane provider installation or
API key in Google Cloud. See the [complete setup guide](../../docs/guides/connect-google-cloud.md).

Use this composed module when Terraform should own both infrastructure and the
Polylane connection lifecycle.

## Requirements

This module requires the companion GCP backend and a provider build containing
the GCP resources. The currently published Polylane provider does not include
these resources. Use a locally built provider while reviewing this PR; release
publication and a full backend live proof are rollout gates.

Configure `POLYLANE_API_KEY` with `cloud_accounts:read`, `cloud_accounts:write`
and `cloud_accounts:delete`, and Google Application Default Credentials for the
installer. The installer needs project roles `roles/serviceusage.serviceUsageAdmin`,
`roles/iam.serviceAccountAdmin`, `roles/iam.workloadIdentityPoolAdmin`,
`roles/resourcemanager.projectIamAdmin`, `roles/iam.roleAdmin`,
`roles/pubsub.admin`, `roles/cloudasset.owner`, and `roles/logging.configWriter`.
Billing and organization policies must permit these resources. These installation
permissions are separate from the runtime reader grants.

From a checkout, reference the module with a local path:

```hcl
terraform {
  required_providers {
    polylane = { source = "registry.terraform.io/coreplanelabs/polylane" }
    google = { source = "hashicorp/google", version = "~> 7.0" }
    google-beta = { source = "hashicorp/google-beta", version = "~> 7.0" }
  }
}

provider "polylane" {}
provider "google" { project = "example-project" }
provider "google-beta" { project = "example-project" }

module "polylane_gcp" {
  source           = "./modules/gcp"
  workspace_id     = "ws_00000000000000000000000000000000"
  project_id       = "example-project"
  installer_member = "serviceAccount:installer@example-project.iam.gserviceaccount.com"
}
```

For remote consumption after publication, use this repository's Git source with
`//modules/gcp?ref=<reviewed-commit-or-release>`; pin an immutable revision.
There is no claim that this module is separately published in the Terraform Registry.

The three inputs are workspace ID, Google project ID and the IAM member running
Terraform. Outputs are `cloud_account_id` and `request_id`. Resource names and
trust parameters come from Polylane, so installations do not share identities.

## Lifecycle and recovery

Terraform creates the request, all Google resources and IAM grants, then calls
activation. Polylane independently validates Google metadata and federation;
`registered` means durable registration, not completed initial synchronization.
IAM propagation may require rerunning apply; the existing request remains in state.
Pending requests expire after seven days. Replace the request and connection to
start again; default request creation generates a fresh idempotency UUID.
An explicit raw-resource `idempotency_key` must be changed after revocation or expiry.

Destroy disconnects Polylane before Google teardown. A disconnect error halts
normal destroy. The Google-only child owns dedicated service accounts, WIF pool
and provider, IAM memberships and custom role, Pub/Sub topics/subscriptions,
CAI feed and Logging sink. Enabled Google APIs and generated Google service agents
remain project-level prerequisites after destroy. Shared Google resources are not
removed. Infrastructure Manager's deployment state/staging buckets are owned by IM.

A disconnected integration can be cleaned up from Google even if Polylane is
unreachable, but that is an explicit recovery operation: retain the Terraform
state, disconnect/revoke through Polylane when available, and use the Google-only
configuration/state owner to remove its resources. Do not blindly remove state
or apply a second copy of the module over existing resources.

During a long setup, push messages may reach the dead-letter subscription before
activation. Inspect the `${resource_prefix}-dead-letters` pull subscription,
persist envelopes before acknowledging, unwrap their original message and
republish to the original assets/logs topic. Initial reconciliation restores
inventory; logs require replay. Both primary and dead-letter retention are one day.
