# Repository Guidelines

## Project Structure & Module Organization
- `cmd/cli` contains the Cobra-based CLI entrypoint and subcommands (`feed`, `ask`, `models`, `testing`).
- `cmd/benchmarker` holds benchmarking helpers; release artifacts are produced in `dist/` via GoReleaser.
- `internal/adapter` hosts integrations (feeds storage, Ollama client); `internal/model` defines core domain structs.
- `internal/config` loads defaults from `config.yaml`; `internal/ui` provides terminal UI helpers.
- `data/feeds.db` is a sample Storm/Bolt database for local use only.

## Build, Test, and Development Commands
- `go run ./cmd/cli --help`: show CLI commands and flags.
- `go run ./cmd/cli feed list`: list configured feeds.
- `go run ./cmd/cli feed refresh`: fetch new articles from configured feeds.
- `go run ./cmd/cli feed search "golang debugging"`: search articles.
- `go run ./cmd/cli ask "Explain supervised vs unsupervised"`: query the knowledge base.
- `go run ./cmd/cli testing`: run the CLI testing workflow.
- `go run ./cmd/cli models ls` / `go run ./cmd/cli models ps`: inspect available/loaded models.
- `make add`: seed default feeds into the local database.
- `make list` / `make ask`: convenience wrappers for common CLI commands.
- `make refresh`: pull new articles into the local database.
- `make search`: run an example query.
- `make test`: run `go test -cover ./...`.
- `make linter`: run `golangci-lint`.
- `make vulncheck`: run `govulncheck`.
- `make validate`: run tests, lint, and vulnerability checks together.
- `make release --snapshot`: build cross-platform artifacts into `dist/` with GoReleaser.

## Coding Style & Naming Conventions
- Go 1.26.0; format with `gofmt` and imports via `goimports` (or `golangci-lint` defaults).
- Package names are lowercase and match folder names; exported identifiers are descriptive and avoid abbreviations.
- CLI commands follow verb-first naming (e.g., `feed add`, `feed refresh`, `models ls`).

## Testing Guidelines
- Use Go’s standard `testing` package; place tests alongside code as `_test.go`.
- Prefer table-driven tests for adapters and UI helpers; mock external services (Ollama, feeds).
- Run `make test` or `make validate` before proposing changes.

## Commit & Pull Request Guidelines
- Commit messages are short, present-tense, and focused (e.g., “Improve feed refresh output”).
- PRs should describe behavior changes, note any `config.yaml` default impacts, and include CLI reproduction steps.
- Attach logs or screenshots if terminal UI output changes.

## CI & Automation
- No CI pipeline is configured yet; run `make validate` locally before opening a PR.

## Security & Configuration Tips
- Do not commit secrets or host endpoints; use env vars or local overrides to `config.yaml`.
- Do not commit regenerated `data/feeds.db` or `dist/` artifacts unless publishing a release.
