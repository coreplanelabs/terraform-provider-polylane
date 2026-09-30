# Google-only connection infrastructure

This child module is the Infrastructure Manager blueprint and the infrastructure
part of `../gcp`. It uses only `hashicorp/google` and `hashicorp/google-beta`.
It contains no Polylane provider credentials. The console supplies public
request identity parameters and separately activates through the authenticated
Polylane API after the deployment finishes.

Inputs: `project_id`, `request_id`, `subject`, `issuer_url`, `push_endpoint`,
`resource_prefix`, `installer_member`. The installer member is the execution
service account prefixed with `serviceAccount:` (or the user running Terraform).
Outputs: `project_number`, `reader_email`, `push_subject`.

Infrastructure Manager needs enabled `config.googleapis.com` and
`cloudbuild.googleapis.com`, a region and an existing execution service account
with `roles/config.agent` plus the installation roles listed in
[the parent module](../gcp/README.md). Creating that privileged execution identity
is a customer bootstrap step; this child does not grant its installer project
administration permissions. Run IM against a root configuration that references
this child with a commit-pinned Git source and provides its public inputs.

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
