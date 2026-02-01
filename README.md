# daily-code-churn

Generate a daily code churn SVG (added/removed lines per day) from git history.

## Usage (as a GitHub Action)

Add this to a workflow in another repo:

```yml
name: churn
on:
  workflow_dispatch:

jobs:
  churn:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - uses: oierxjn/daily-code-churn@latest
        with:
          days: "30"
          out: "daily-churn.svg"
```

### Inputs

- `days`: How many days to include (string, optional)
- `branch`: Branch/ref to analyze (string, optional)
- `out`: Output SVG path (string, optional)
- `width`: SVG width (string, optional)
- `height`: SVG height (string, optional)
- `commit_branch`: Target branch to push the generated output (string, optional)

### Environment variables

The Go binary reads these when inputs are not provided:

- `CHURN_DAYS`
- `CHURN_BRANCH`
- `CHURN_OUT`
- `CHURN_WIDTH`
- `CHURN_HEIGHT`

Precedence: `with:` inputs > `CHURN_*` env > built-in defaults.

## Output commit behavior

This Action commits the generated output file back to the repository by default.
Make sure your workflow grants `contents: write` permission and set `commit_branch`.

## Workflows in this repo

- `build-and-commit-dist`: Manually build and push updated `dist/` binaries to `main`
- `code-churn`: Manually run this Action in the repo to generate and commit `daily-churn.svg`

## Repository layout

- `index.js`: JavaScript Action entrypoint
- `go/`: Go module
- `dist/`: Prebuilt binaries (committed for Action runtime)
- `scripts/`: Build scripts
- `action.yml`: Action metadata

## Build binaries

Build all target binaries into `dist/`:

```bash
bash scripts/build.sh
```

On Windows:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\scripts\build.ps1
```

The JavaScript Action expects binaries in `dist/` with names like:

- `dist/daily-code-churn-linux-amd64`
- `dist/daily-code-churn-darwin-arm64`
- `dist/daily-code-churn-windows-amd64.exe`

## Example: run locally

```bash
CHURN_DAYS=14 go run ./go
```
