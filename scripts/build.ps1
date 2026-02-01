param(
  [string]$OutDir = "dist"
)

$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
$GoDir = Join-Path $RootDir "go"
$DistDir = Join-Path $RootDir $OutDir

New-Item -ItemType Directory -Path $DistDir -Force | Out-Null

function Build([string]$Goos, [string]$Goarch, [string]$Ext) {
  $out = Join-Path $DistDir ("daily-code-churn-{0}-{1}{2}" -f $Goos, $Goarch, $Ext)
  Write-Host "Building $out"
  $env:GOOS = $Goos
  $env:GOARCH = $Goarch
  $env:CGO_ENABLED = "0"
  Push-Location $GoDir
  try {
    go build -trimpath -o $out .
  } finally {
    Pop-Location
  }
}

Build linux amd64 ""
Build linux arm64 ""
Build darwin amd64 ""
Build darwin arm64 ""
Build windows amd64 ".exe"
Build windows arm64 ".exe"
