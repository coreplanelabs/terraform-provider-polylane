# Publishing and maintaining the provider

The provider address is `coreplanelabs/polylane`. The first release is `0.1.0`;
pre-1.0 breaking changes bump the minor version. Release PRs remain the review
point for each version; merging ordinary fixes does not publish immediately.

## One-time GitHub setup

1. Merge the provider readiness PR after CI passes. Review and apply the
   infrastructure PR that makes the repository public and enables secret
   scanning, push protection, release-tag protection, and the environments
   below. Review the full repository history, issues, PRs, Actions logs, and
   artifacts before making them public. Automated secret scanning does not
   detect all confidential information.
2. Create a dedicated GitHub App installed **only** on this repository, with
   Contents, Pull requests, and Issues read/write permissions (Issues is used
   for release labels). Do not give it administration or organization access.
   In the `release` GitHub environment, set variable `RELEASE_APP_CLIENT_ID` and
   secret `RELEASE_APP_PRIVATE_KEY`. Store the private key in the team's secret
   manager. App-authored release PRs trigger CI; the default `GITHUB_TOKEN`
   cannot reliably trigger those downstream workflows.
3. In that same environment, add `GPG_PRIVATE_KEY` (ASCII-armored private key)
   and `PASSPHRASE`. Use a dedicated passphrase-protected RSA signing key, not
   a personal key or the default ECC type. Keep a backup in the team's secret
   manager. Never save private key material inside this checkout or paste it
   into issues, PRs, or chat.
4. The infrastructure configuration restricts both `release` and `acceptance`
   environments to the `main` branch. Put credentials in these environments,
   not in repository-wide secrets available to arbitrary same-repo branches.
5. In `acceptance`, add secret
   `OP_TERRAFORM_PROVIDER_SERVICE_ACCOUNT_TOKEN` and variable
   `OP_TERRAFORM_PROVIDER_ENVIRONMENT_ID`. The service account must only read
   the dedicated `Terraform Provider` 1Password Environment and its referenced
   items. Use a disposable test workspace. See the README for API scopes.
6. After the visibility change, confirm fork workflow approval settings and
   Depot's runner policy are appropriate for public contributions. Keep
   organization production secrets restricted to private/selected repositories.
   Do not grant this repository access to the organization-wide 1Password bot.

For an RSA signing key, `gpg --full-generate-key` provides an interactive flow.
Choose RSA with signing capability, 4096 bits, the team's release identity,
expiry, and a strong passphrase. Record its fingerprint. Export only the public
part to share with the Registry:

```sh
gpg --armor --export "$GPG_FINGERPRINT" > /tmp/polylane-release-public.asc
```

Transfer the private export directly to the secret manager or GitHub environment
secret through a pipe or protected temporary file, never terminal output.

## First release and Registry enrollment

1. Confirm the replacement release PR proposes `0.1.0`, including
   `.release-please-manifest.json`, `internal/buildinfo/buildinfo.go`, and
   `CHANGELOG.md`. The readiness commit also carries `Release-As: 0.1.0` so the
   earlier unmerged `1.0.0` proposal is superseded.
2. Merge that release PR once all checks pass. Release Please creates an
   immutable version tag and a **draft** release. GoReleaser checks out that
   tag, tests the source, builds the platform archives, and signs SHA256SUMS.
   The workflow downloads the uploaded files, verifies their checksums,
   manifest, archive layout, and GPG signature, then publishes the release.
3. Sign in at [Terraform Registry](https://registry.terraform.io). For team
   ownership, use an HCP Terraform organization and claim the GitHub namespace
   `coreplanelabs` under **Registry → Public namespaces**. Use an organization
   owner account to authorize the GitHub connection. If the namespace is
   already claimed, use its owning HCP organization rather than creating a
   second claim. HCP Europe does not currently support public namespaces.
4. Add the ASCII-armored **public** RSA key to the namespace's **Settings → New
   GPG Key**. For a namespace still managed directly in the Registry, use
   **User Settings → Signing Keys** instead. Never upload the private key.
5. In the namespace's **Providers → Publish → New provider** flow (or the
   Registry's **Publish → Provider** flow), select
   `coreplanelabs/terraform-provider-polylane`. The public repository and at
   least one complete, signed release must already exist. Authorize the
   required repository webhook access and complete the publishing form.
6. Confirm version `0.1.0`, documentation, and target platforms appear. The
   provider uses protocol `6.0`. Preview generated docs with the
   [Registry doc preview](https://registry.terraform.io/tools/doc-preview).
7. In a fresh directory without developer overrides, use the following config
   and run `terraform init` and `terraform providers schema -json`. This checks
   installation and protocol startup without needing an API key or creating
   resources. Inspect the reported signing fingerprint and lockfile.

```hcl
terraform {
  required_providers {
    polylane = {
      source  = "coreplanelabs/polylane"
      version = "0.1.0"
    }
  }
}
```

## Later releases and recovery

Conventional Commits update a release PR automatically. Merge it when ready;
the same signed publication flow runs. Registry enrollment creates a GitHub
release webhook, so later published releases are ingested automatically. No
Registry token is required in the build workflow.

If building or uploading fails, the release stays in draft. Fix the cause and
resume the **release-please** workflow on **main**, supplying the existing tag:

```sh
gh workflow run release.yaml --ref main -f tag=v0.1.0
```

Recovery requires an existing draft, a tag reachable from `main`, and a matching
version manifest. It rejects published releases. If source changes are needed,
release a new version; do not move a tag or replace published artifacts. Draft
uploads can be replaced while recovering an interrupted upload.

If Registry ingestion fails after publication, inspect the release webhook and
provider publishing errors. Use the Registry's Resync action after fixing the
cause. Keep old public signing keys registered when rotating keys so installed
versions remain verifiable.

## Local release validation

`task release:check` builds all platform archives in snapshot mode, without
signing or publishing, then checks the Registry artifact contract. It runs in
PR CI. `task security` scans reachable Go vulnerabilities, Git history and the
worktree for secrets, and dependencies/IaC with Trivy. CodeQL activates once the
repository is public; actionlint and zizmor check workflows on PRs.

References: [provider publishing requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing),
[HCP public namespaces](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/organizations/public-namespace),
[namespace signing keys](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/organizations/public-namespace/manage).
