# Google Cloud dashboard blueprint

Infrastructure Manager runs this root configuration from a pinned Git commit.
The Polylane dashboard passes the connection request's seven inputs after the
project administrator approves setup. Providers run as the project's dedicated
installer service account. No Polylane API key or administrator OAuth token is
passed to Terraform.

The existing [Google connection module](../../modules/google-connection) owns
the resources and teardown behavior. API bootstrap, the installer account and
its project roles are prepared before deployment and remain customer-owned.
The dashboard verifies the resulting federation independently before registering
the cloud account. The dashboard pins the Infrastructure Manager executor to Terraform 1.5.7.
This root also validates with newer Terraform 1.x tooling and pins the Google
providers to 7.0.0.

## GKE access

GKE object reads and pod logs use the same federated runtime service account.
`gke_access_enabled` defaults to `true`; set it to `false` to omit Kubernetes grants.
Enable external access through each cluster's DNS endpoint, or connect the Polylane
Kubernetes agent for clusters whose network or service perimeter blocks access.

`gke_operations_enabled` defaults to `false`. Enabling it grants Kubernetes Engine
Developer access and node updates to the runtime identity. Polylane also requires
workspace write access, safety review and human approval before mutations.
Reapply this module to add GKE permissions to an existing connection.
