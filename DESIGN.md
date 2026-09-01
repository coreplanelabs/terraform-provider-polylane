# Provider design

The first provider release is intentionally narrower than the Polylane API.
Terraform works best when a remote object has stable identity, declarative
readback, and lifecycle operations whose effects are unsurprising. This design
uses those constraints to decide what belongs in the provider.

## Supported surface

`polylane_workspace` adopts a workspace that already exists. A workspace API
key is bound to an existing workspace and cannot safely represent the user
session needed by workspace creation. The resource therefore supports import,
read, and safe property updates. Create returns an import instruction. Destroy
removes the resource from Terraform state without calling the destructive
workspace DELETE endpoint.

Each `polylane_workspace_*_settings` resource represents one existing singleton
under a workspace ID. Create and update both PATCH the singleton, read refreshes
all managed fields, and destroy stops management without attempting a reset.
The API has no delete operation for these settings. The provider exposes intent
such as `enabled`, rather than API implementation details such as a
`disabledAt` timestamp.

`polylane_workspace_member` follows the same adoption model as the workspace.
The public API can read, update, and delete a membership, but it cannot create
one. Terraform therefore manages an imported member's role and scopes while
destroy only stops management. This avoids a resource Terraform could destroy
but never recreate.

`polylane_team` has a complete API lifecycle and is a normal CRUD resource.
`polylane_team_member` models the stable workspace/team/user relationship and
also has a complete create, read, and delete lifecycle. Composite imports use
slash-separated IDs such as `workspace_id/team_id/user_id`.

The workspace resource excludes custom LLM provider configuration and all
credential values. Terraform Plugin Framework supports write-only arguments in
Terraform 1.11 and later, but adding a secret is still a product and lifecycle
decision, not a reason to expose every write-only API field. See the official
[write-only argument guidance](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments).

## Terraform conventions

- Import uses the stable Polylane workspace ID and the Plugin Framework import
  lifecycle. Terraform calls Read after import to populate the remaining state.
  See the [Plugin Framework import contract](https://developer.hashicorp.com/terraform/plugin/framework/resources/import).
- The documented import-block workflow keeps adoption reviewable in normal
  plan/apply output. See Terraform's [import block reference](https://developer.hashicorp.com/terraform/language/block/import).
- A 404 during refresh removes the object from state, following the Plugin
  Framework [Read guidance](https://developer.hashicorp.com/terraform/plugin/framework/resources/read).
- Changing `workspace_id` replaces a settings resource because it changes the
  singleton being managed. See the Plugin Framework
  [plan-modification guidance](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification).

## Generated API client

The checked-in client follows the SDK generation workflow used by `zpax-go`:
the upstream document is retained, a deterministic normalized document is used
for generation, the generator version is pinned in `go:generate`, and CI checks
for generated drift. `task sync:openapi` downloads the production document,
strips the `/v1` prefix from paths because the provider endpoint already
contains it, regenerates, and tests.

Generation is deliberately constrained by `include-operation-ids` in
`api-client-config.yaml`. The full spec remains available for review and future
work, but only provider-owned operations become Go methods. A thin handwritten
adapter keeps Terraform-facing types stable and centralizes Polylane envelope
errors. Model-training settings remain one explicit handwritten exception
because the route is hidden from the public document.

## Deferred surface

Workspace invitations remain deferred because invitation policy and lifecycle
need a separate design. Integrations that are only a container for sensitive API
keys also remain deferred. A future integration resource should have useful
non-secret lifecycle and readback on its own; otherwise it does not belong in
Terraform.
