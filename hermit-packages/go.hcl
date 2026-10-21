description = "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software."
binaries = ["bin/*"]
strip = 1
test = "go version"
repository = "https://github.com/golang/go"
source = "https://go.dev/dl/go${version}.${os}-${arch}.tar.gz"

env = {
  "GOROOT": "${root}",
  "GOBIN": "${HERMIT_ENV}/.hermit/go/bin",
  "PATH": "${GOBIN}:${PATH}",
  "GOTOOLCHAIN": "local",
}

version "1.27.2" "1.27.3" {}

sha256sums = {
  "https://go.dev/dl/go1.27.2.darwin-amd64.tar.gz": "587b59182488b23aa6e5fc25110405a3e0e5b38ed2f5b2f46ed13c32aee356fe",
  "https://go.dev/dl/go1.27.2.darwin-arm64.tar.gz": "76812b213b1b2302c978d28fa52fa92d541704b9e7d9d5db8002c50e4018c4c5",
  "https://go.dev/dl/go1.27.2.linux-amd64.tar.gz": "ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5",
  "https://go.dev/dl/go1.27.2.linux-arm64.tar.gz": "94f3e30b8e374bc285e7dadc11e0865726b9bc6e85b841ccceaabc0214c6b7c8",
  "https://go.dev/dl/go1.27.3.darwin-amd64.tar.gz": "b6f8e931e4ab497fcde6da8ae4499a58f9d258e55c8a80824a516eae5950a750",
  "https://go.dev/dl/go1.27.3.darwin-arm64.tar.gz": "dd3908cf20959d798c1e06ebe5b72213f0505ce01373b40ce3110975a6cd99c2",
  "https://go.dev/dl/go1.27.3.linux-amd64.tar.gz": "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40",
  "https://go.dev/dl/go1.27.3.linux-arm64.tar.gz": "a06e99309dbcbdca1a8ccd2282c3ee146e55930886e26b4f1193269587165f49",
}
