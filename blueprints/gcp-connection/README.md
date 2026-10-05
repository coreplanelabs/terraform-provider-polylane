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
the cloud account. This root uses Terraform 1.5.7 and Google providers 7.0.0.
