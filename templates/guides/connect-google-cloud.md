---
page_title: "Connect Google Cloud (Beta)"
subcategory: ""
description: |-
  Connect one Google Cloud project through the console, a setup script or Terraform.
---

# Connect Google Cloud (Beta)

Connect one project for inventory across Cloud Asset Inventory resource types,
including Compute Engine, Cloud Run, GKE, Cloud SQL and networking. Pub/Sub topics
and Storage buckets also have resource logs and metrics. Project Admin Activity
supports read-only investigation. Kubernetes API access and automated
Google changes are outside this beta.

## Set up from the console

1. In Polylane, choose **Continue with Google** and authorize a project
   administrator account.
2. Select one project and review the resources, permissions and Google costs.
3. Choose **Set up connection**. Polylane enables bootstrap APIs, prepares the
   execution service account and submits a pinned Terraform blueprint to
   Infrastructure Manager in your project.
4. When the deployment is ready, choose **Verify and connect**. Polylane checks
   Google identities, federation, delivery configuration and project access before
   starting inventory synchronization.

The Google setup token is temporary and is removed after deployment or
cancellation. Ongoing investigation uses a separate read-only identity with
short-lived federation credentials. Keep the panel open to advance setup, or
reopen it to resume. Google continues a deployment already submitted. Choose
**Authorize again to resume** if setup access expires.

Infrastructure Manager owns Terraform state in your Google project. Cancelling
setup revokes the pending Polylane request and clears saved setup access; delete
any Google deployment and its installer account separately. Requests expire
after seven days. After cancellation or expiry, clean up earlier resources and
start a new request.

## Copy and run from the console

Choose **Use a setup script or Terraform**, enter your Google project ID and
numeric project number in Polylane and choose
**Prepare setup**. Choose **Copy setup script**, paste the complete block into
Google Cloud Shell or a terminal signed in with `gcloud`, and press Enter.
Your project, workspace and connection identity are already filled in. When the
script finishes, return to Polylane and choose **Verify and connect**. The saved
project number is reused when you reopen setup. Switching from dashboard setup
preserves the existing connection request and pinned blueprint, so it does not
create a second deployment identity.

The script includes its Terraform configuration. Commands
use literal values and `&&` failure chaining. It does not set shell options,
assign shell variables, install traps, change directories or require environment
variables. It keeps the generated configuration in a connection-specific
`.polylane/` directory for inspection and retry.

The following is a review example with synthetic identifiers. Run the script
prepared by your own Polylane workspace; these example identities cannot connect
a project.

```sh
mkdir -p .polylane/polylane-0000000000000000 &&
cat > .polylane/polylane-0000000000000000/main.tf <<'POLYLANE_TERRAFORM' &&
terraform {
  required_version = "= 1.5.7"
  required_providers {
    google = { source = "hashicorp/google", version = "= 7.0.0" }
    google-beta = { source = "hashicorp/google-beta", version = "= 7.0.0" }
  }
}
provider "google" { project = "example-project" }
provider "google-beta" { project = "example-project" }
module "polylane_gcp" {
  source = "git::https://github.com/coreplanelabs/terraform-provider-polylane.git//modules/google-connection?ref=a44c0a5b158bb6d155bbab22dcbe03e23aee535c"
  project_id = "example-project"
  request_id = "gcpconn_000000000000000000000000"
  subject = "ws_00000000000000000000000000000000:gcpconn_000000000000000000000000"
  issuer_url = "https://api.polylane.com/v1/gcp_federation/ws_00000000000000000000000000000000/gcpconn_000000000000000000000000"
  push_endpoint = "https://api.polylane.com/v1/gcp_events/ws_00000000000000000000000000000000/gcpconn_000000000000000000000000"
  resource_prefix = "polylane-0000000000000000"
  installer_member = "serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com"
}
output "project_number" { value = module.polylane_gcp.project_number }
POLYLANE_TERRAFORM
gcloud services enable config.googleapis.com cloudbuild.googleapis.com serviceusage.googleapis.com iam.googleapis.com cloudresourcemanager.googleapis.com storage.googleapis.com --project=example-project --quiet &&
if ! gcloud iam service-accounts describe polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --project=example-project >/dev/null 2>&1; then
  gcloud iam service-accounts create polylane-0000000000000000-im --display-name="Polylane connection setup" --project=example-project --quiet
fi &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/config.agent --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/iam.serviceAccountAdmin --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/iam.workloadIdentityPoolAdmin --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/iam.roleAdmin --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/pubsub.admin --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/cloudasset.owner --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/logging.configWriter --condition=None --quiet --format='value(version)' &&
gcloud projects add-iam-policy-binding example-project --member=serviceAccount:polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --role=roles/serviceusage.serviceUsageAdmin --condition=None --quiet --format='value(version)' &&
gcloud infra-manager deployments apply projects/example-project/locations/us-central1/deployments/polylane-0000000000000000 --service-account=projects/example-project/serviceAccounts/polylane-0000000000000000-im@example-project.iam.gserviceaccount.com --local-source=.polylane/polylane-0000000000000000 --tf-version-constraint=1.5.7 --project=example-project --quiet &&
printf '%s\n' 'Google Cloud setup is complete. Return to Polylane and choose Verify and connect.'
```

The first Google commands enable the bootstrap APIs, create the Infrastructure
Manager execution service account if needed, and grant its installation roles.
These run under the signed-in project administrator. Infrastructure Manager then
uses that service account to run Terraform. The module enables the integration
APIs and provisions the reader, federation, authenticated Pub/Sub delivery,
Cloud Asset feed and Logging sink. The installer's project administration roles
are separate from the runtime reader's read-only grants.

`--local-source` uploads the generated Terraform directory; no manually uploaded
archive or customer source repository is needed. The Terraform root uses one
commit-pinned `modules/google-connection` block and only Google providers.
Polylane authenticates the final verification through the console, so this
Infrastructure Manager deployment needs no Polylane API key or Polylane provider
binary. The Google console link shows deployment progress.

A stopped script can be rerun with the same identity. Pending requests last seven
days. If a request expires or is cancelled, remove its old Google deployment and
prepare a fresh connection. Disconnect in Polylane before destroying the
Infrastructure Manager deployment. The bootstrap execution account and its
project grants are outside the module and require separate cleanup when no
longer needed. Enabled APIs remain available to other project workloads.

## Manage the whole connection in Terraform

For Terraform users, `modules/gcp` composes the Polylane request and activation
resources with the same Google-only child module. Configure the Polylane and
Google providers, then add one module block:

```hcl
module "polylane_gcp" {
  source           = "git::https://github.com/coreplanelabs/terraform-provider-polylane.git//modules/gcp?ref=a44c0a5b158bb6d155bbab22dcbe03e23aee535c"
  workspace_id     = "ws_00000000000000000000000000000000"
  project_id       = "example-project"
  installer_member = "user:administrator@example.com"
}
```

These are example inputs; use your workspace, project and installer identity.
The module obtains the project number from Google and connects it to the request.
It waits for every Google resource and IAM grant before activation, and orders
Polylane disconnect before infrastructure teardown on destroy.

This composed path requires a Polylane provider build containing
`polylane_gcp_connection_request` and `polylane_gcp_connection`, plus the companion
backend. Until those resources are released, reviewers must use a local provider
build. The console's Google-only Infrastructure Manager path does not depend on
publishing that provider build. See the [module requirements and lifecycle](https://github.com/coreplanelabs/terraform-provider-polylane/tree/codex/gcp-beta-connection/modules/gcp).

Registration completes before initial inventory synchronization. A transient
server error can be retried within the provider's bounded activation budget;
identity or permission errors require correcting setup and rerunning apply.
Keep the existing request and Terraform state when retrying.
