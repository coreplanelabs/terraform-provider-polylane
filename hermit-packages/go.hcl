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

version "1.26.9" {}

sha256sums = {
  "https://go.dev/dl/go1.26.9.darwin-amd64.tar.gz": "00c29e3d4c8562f547410daa18d2959b641bd691ab00c4a525ea817a0253d4bd",
  "https://go.dev/dl/go1.26.9.darwin-arm64.tar.gz": "f9bb7c0a02506c5d9bf0d1eb1f7ee6c7684f844ae49558308a0427b830e022cc",
  "https://go.dev/dl/go1.26.9.linux-amd64.tar.gz": "42d158b4d8f7b61ac0a830567c940a86098fb7aac52e467a5ebec03ef5cc2f8d",
  "https://go.dev/dl/go1.26.9.linux-arm64.tar.gz": "4a97373d49fcacdcf3694fea368a500b00ee3e963974f3e7514132717632f052",
}
