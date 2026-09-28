# Contributing

Use the pinned tools with `source bin/activate-hermit`, then run `task init`
and `task do`. Include regression tests for behavior changes. Provider schema
or example changes must include regenerated Registry docs (`task docs`).

Pull requests use Conventional Commit titles such as `fix(provider): handle a
missing team`. CI builds and tests contributions without live credentials.
Live acceptance tests run separately against a dedicated test workspace.
Never include API keys, Terraform state, saved plans, or customer data in issues,
fixtures, or pull requests. See [SECURITY.md](SECURITY.md) for private reports.
