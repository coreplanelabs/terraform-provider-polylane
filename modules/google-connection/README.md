# Google-only connection infrastructure (Beta)

This child module is the Infrastructure Manager blueprint and the infrastructure
part of `../gcp`. It uses only `hashicorp/google` and `hashicorp/google-beta`.
It contains no Polylane provider credentials. The console supplies public
request identity parameters and separately activates through the authenticated
Polylane API after the deployment finishes.

Inputs: `project_id`, `request_id`, `subject`, `issuer_url`, `push_endpoint`,
`resource_prefix`, `installer_member`. The installer member is the execution
service account prefixed with `serviceAccount:` (or the user running Terraform).
Outputs: `project_number`, `reader_email`, `push_subject`.

## Prefilled Infrastructure Manager setup

In Polylane, enter the project ID and numeric project number up front, choose
Prepare setup, then Copy setup script. Paste the complete block into Google
Cloud Shell or a terminal already signed in with gcloud and press Enter. Return
to Polylane and choose Verify and connect after deployment completes.

The 38-line script embeds one commit-pinned module block, enables the bootstrap
APIs, creates a dedicated execution service account, grants its installer roles
and runs `gcloud infra-manager deployments apply --local-source=...`. All inputs
are literal values from the saved connection request. Commands chain with `&&`
without shell variables, `set`, traps or a subshell wrapper. Terraform is kept in
`.polylane/<resource-prefix>/main.tf` for inspection and retry. No environment
configuration, file download or manual Terraform upload is required.

The [setup guide](../../docs/guides/connect-google-cloud.md) contains the complete
review example and explains the Terraform-managed alternative.

Infrastructure Manager requires `config.googleapis.com`,
`cloudbuild.googleapis.com`, `serviceusage.googleapis.com`, `iam.googleapis.com`,
`cloudresourcemanager.googleapis.com` and `storage.googleapis.com` before apply.
The bootstrap execution service account receives `roles/config.agent` and the
installation roles listed in [the parent module](../gcp/README.md). Creating that
privileged execution identity is part of the generated bootstrap script; this
child module does not grant its installer project administration permissions.
The child enables the APIs used by the integration itself.

The runtime reader receives Cloud Asset Viewer, Logging Viewer, Monitoring
Viewer and Pub/Sub Viewer plus a metadata-only custom role for activation and
bucket metadata. It has no Storage object read/write grant. The delivery account
has no project roles. Only the project's Pub/Sub service agent can mint its
push token through the module's service-account binding.

Use a stable public HTTPS issuer. The WIF provider constrains both immutable
subject and request ID and retains Google's provider-specific default audience.
Push subscriptions require the exact endpoint as OIDC audience and use a
dedicated delivery account. Feed coverage is Pub/Sub topics and Storage buckets.
The Logging sink includes their Admin Activity records and error logs with
`pubsub_topic` or `gcs_bucket` monitored-resource types.

Deletion through IM removes this module's dedicated resources, not the Polylane
registration. Disconnect in Polylane first. Enabled APIs and Google-managed
service identities remain enabled because other project workloads may use them.
The composed parent module orders disconnect before teardown automatically.

The bootstrap execution account and its project-level installer grants are not
owned by this child. Remove them separately when the deployment no longer needs
to be updated or destroyed.
