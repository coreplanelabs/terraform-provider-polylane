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

## Console and script setup

In Polylane, choose **Continue with Google**, authorize a project administrator,
select a project and review the setup. **Set up connection** enables bootstrap
APIs, prepares the execution identity and submits this pinned blueprint to
Infrastructure Manager. Choose **Verify and connect** once deployment is ready.
Temporary Google setup access is removed after deployment or cancellation.

Alternatively, choose **Use a setup script or Terraform**, enter the project ID
and numeric project number, then choose **Prepare setup** and **Copy setup script**.
Paste the complete block into Google Cloud Shell or a terminal already signed in
with gcloud and press Enter. Return to Polylane and choose **Verify and connect**
after deployment completes. A script resumed from dashboard setup uses that same
request, resource names and immutable blueprint revision.

The script embeds one commit-pinned module block, enables the bootstrap
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
Viewer and Pub/Sub Viewer plus a custom role for activation, bucket metadata
and trace reads, including `serviceusage.services.use` for project API reads.
For configuration reads during investigations, it also receives Security
Reviewer, Service Usage Viewer and the metadata viewer role of each supported
service: Compute, Cloud Run, Cloud Functions, Cloud SQL, Spanner, Bigtable,
BigQuery metadata, Memorystore for Redis and Memcached, Filestore, Cloud DNS
reader, Serverless VPC Access, Cloud Build builds, Cloud Deploy, Cloud
Scheduler, Workflows, Eventarc, Dataflow, Dataproc, API Gateway, Secret Manager
and Cloud KMS. The full list is `local.reader_roles` in `main.tf`. These
predefined roles read configuration and metadata. They do not read secret
payloads, BigQuery table rows or Storage objects. Artifact Registry and AlloyDB
have no metadata-only predefined role, so the custom role grants only their
repository, image, cluster and instance `get` and `list` permissions. Reapply
the module to give an existing reader these grants.
The module enables Cloud Trace and grants `cloudtrace.traces.get` and
`cloudtrace.traces.list` to this reader. Trace collection uses the application's
existing instrumentation. Reapply the module to add trace access to an existing
connection.
The custom reader role includes `iam.roles.get` at the connected project for
IAM permission analysis. Cloud Asset Viewer supplies `cloudasset.assets.listIamPolicy`.
The reader can resolve predefined and project custom roles; it receives no
folder or organization grant. Installer IAM administration permissions do not
supply runtime reader access. Reapply the module for an existing reader to receive
role-definition access.
It has no Storage object read/write grant. The delivery account
has no project roles. Only the project's Pub/Sub service agent can mint its
push token through the module's service-account binding.

Use a stable public HTTPS issuer. The WIF provider constrains both immutable
subject and request ID and retains Google's provider-specific default audience.
Push subscriptions require the exact endpoint as OIDC audience and use a
dedicated delivery account. The asset feed uses the `.*` asset-type regular
expression, covering all resource types supported by Cloud Asset Inventory.
The Logging sink forwards
Pub/Sub and Storage Admin Activity mutations as an additional source of resource
changes. Log and metric investigations query Google directly.

Deletion through IM removes this module's dedicated resources, not the Polylane
registration. Disconnect in Polylane first. Enabled APIs and Google-managed
service identities remain enabled because other project workloads may use them.
The composed parent module orders disconnect before teardown automatically.

The bootstrap execution account and its project-level installer grants are not
owned by this child. Remove them separately when the deployment no longer needs
to be updated or destroyed.

## GKE access

GKE object reads and pod logs use the same federated runtime service account.
`gke_access_enabled` defaults to `true`; set it to `false` to omit Kubernetes grants.
Enable external access through each cluster's DNS endpoint, or connect the Polylane
Kubernetes agent for clusters whose network or service perimeter blocks access.

`gke_operations_enabled` defaults to `false`. Enabling it grants Kubernetes Engine
Developer access and node updates to the runtime identity. Polylane also requires
workspace write access, safety review and human approval before mutations.
Reapply this module to add GKE permissions to an existing connection.
