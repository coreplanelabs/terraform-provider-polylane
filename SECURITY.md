# Security policy

Report vulnerabilities privately to <dev@coreplane.ai>. Include the affected
provider version, a description, and reproduction steps with credentials and
customer information removed. Please do not report vulnerabilities in public
issues. Security fixes target the latest provider release.

Use a workspace-scoped API key with only the permissions required by your
configuration. Prefer the `POLYLANE_API_KEY` environment variable over embedding
credentials in configuration. The provider marks `api_key` sensitive, but
Terraform's sensitive flag is not encryption: protect state, saved plans, CI
logs, and debug output as confidential data.

The endpoint defaults to HTTPS. Override it only with a trusted API endpoint;
HTTP is intended for local development. Authenticated requests do not follow
redirects. Customer AWS resources and credentials remain under the customer's
AWS provider configuration.
