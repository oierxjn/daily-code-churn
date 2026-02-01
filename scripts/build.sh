#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_dir="$root_dir/go"
dist_dir="$root_dir/dist"

mkdir -p "$dist_dir"

build() {
  local goos="$1"
  local goarch="$2"
  local ext="$3"
  local out="$dist_dir/daily-code-churn-${goos}-${goarch}${ext}"
  echo "Building ${out}"
  (cd "$go_dir" && GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -o "$out" .)
}

build linux amd64 ""
build linux arm64 ""
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 ".exe"
build windows arm64 ".exe"
