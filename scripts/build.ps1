param(
  [string]$OutDir = "dist"
)

$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

function Build([string]$Goos, [string]$Goarch, [string]$Ext) {
  $out = Join-Path $OutDir ("daily-code-churn-{0}-{1}{2}" -f $Goos, $Goarch, $Ext)
  Write-Host "Building $out"
  $env:GOOS = $Goos
  $env:GOARCH = $Goarch
  $env:CGO_ENABLED = "0"
  go build -trimpath -o $out ./go
}

Build linux amd64 ""
Build linux arm64 ""
Build darwin amd64 ""
Build darwin arm64 ""
Build windows amd64 ".exe"
Build windows arm64 ".exe"
