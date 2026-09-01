# Terraform Provider for Polylane

This repository contains a Terraform (and OpenTofu) provider for Polylane.

The provider is currently a scaffold: build, lint, test, docs, and release
tooling are in place, but no resources or data sources are implemented yet.

Once published, the provider is sourced as `coreplanelabs/polylane`:

```hcl
terraform {
  required_providers {
    polylane = {
      source  = "coreplanelabs/polylane"
      version = ">= 0.1.0"
    }
  }
}
```

Generated provider documentation lives in [docs/](docs/) and is published on the
Terraform Registry.

## Development

```sh
source bin/activate-hermit
task init
task do
```

Provider configuration can come from Terraform configuration or environment variables:

- `POLYLANE_API_KEY`
- `POLYLANE_ENDPOINT`

### Acceptance Tests

Acceptance tests use the Terraform provider test harness. Today they only
exercise provider configuration; once resources exist they will create real
Polylane objects, and `POLYLANE_API_KEY` will need to be in `.env` or exported
in your shell.

```sh
task test:acc
```

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

Releases are automated:

1. [release-please](https://github.com/googleapis/release-please) opens a release
   PR from Conventional Commits. Merging it tags `vX.Y.Z` and creates a GitHub
   release.
2. The `goreleaser` job in [release.yaml](.github/workflows/release.yaml) then
   builds the provider for all target platforms and attaches the artifacts the
   Terraform Registry ingests: the per-platform zips, `*_SHA256SUMS`,
   `*_SHA256SUMS.sig` (GPG detached signature), and `*_manifest.json`.

### One-time setup to publish to the Terraform Registry

1. Make this repository **public**.
2. Generate a GPG signing key (RSA/DSA, not ECC), then:
   - Add the **public** key at <https://registry.terraform.io> → *User Settings →
     Signing Keys*.
   - Add the **private** key and its passphrase as repository secrets
     `GPG_PRIVATE_KEY` (ASCII-armored) and `PASSPHRASE`.
3. On the registry, *Publish → Provider* and select this repo. The registry adds a
   release webhook and ingests each finalized release automatically.

Provider metadata for the registry lives in
[`terraform-registry-manifest.json`](terraform-registry-manifest.json)
(`protocol_versions: ["6.0"]`, terraform-plugin-framework).

Preview the rendered registry pages at
<https://registry.terraform.io/tools/doc-preview>.
