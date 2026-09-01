#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo_dir=$(cd "${script_dir}/.." && pwd)
spec_dir="${repo_dir}/specs"
original_spec="${spec_dir}/polylane.openapi.original.json"
working_spec="${spec_dir}/polylane.openapi.json"
spec_url="${POLYLANE_OPENAPI_URL:-https://api.polylane.com/v1/doc}"
downloaded_spec=$(mktemp)
trap 'rm -f "${downloaded_spec}"' EXIT

mkdir -p "${spec_dir}"

curl --silent --show-error --fail-with-body "${spec_url}" > "${downloaded_spec}"
jq --sort-keys . "${downloaded_spec}" > "${original_spec}.tmp"
mv "${original_spec}.tmp" "${original_spec}"

# Provider endpoints already include /v1, so generated paths must be relative
# to that base. Preserve the complete upstream document while normalizing only
# its path keys.
jq --sort-keys '
  .paths = (
    .paths
    | to_entries
    | map(.key |= if startswith("/v1/") then ltrimstr("/v1") else . end)
    | from_entries
  )
' "${original_spec}" > "${working_spec}.tmp"
mv "${working_spec}.tmp" "${working_spec}"

cd "${repo_dir}"
go generate ./...
gofmt -w internal/client/generated
go mod tidy
go test ./...
