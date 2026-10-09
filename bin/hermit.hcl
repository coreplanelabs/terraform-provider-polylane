// Use the patched Go release until it is available in the upstream catalog.
sources = ["env:///hermit-packages", "https://github.com/cashapp/hermit-packages.git"]

manage-git = false

github-token-auth {
}
