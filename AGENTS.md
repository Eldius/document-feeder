# Repository Guidelines

## Project Structure & Module Organization
- `cmd/cli` holds the Cobra-driven CLI entrypoint and subcommands (`feed`, `ask`, `models`, `testing`).
- `cmd/benchmarker` provides benchmarking helpers for the feeder; build outputs live in `dist/` when released via GoReleaser.
- `internal/adapter` contains integrations (feeds storage, Ollama client); `internal/model` holds core domain structs.
- `internal/config` loads defaults from `config.yaml`; override via env vars or flags when running the CLI.
- `internal/ui` contains terminal UI helpers for reading and processing articles.
- `data/feeds.db` is a sample Storm/Bolt database; regenerate locally rather than committing changes.

## Build, Test, and Development Commands
- `go run ./cmd/cli --help` to inspect commands; e.g., `go run ./cmd/cli feed list`, `feed refresh`, `feed search "golang debugging"`, or `ask "Explain supervised vs unsupervised"`.
- `make add` seeds the feed list with defaults; `make refresh` pulls new articles; `make search` runs an example query.
- `make test` runs `go test -cover ./...`; `make linter` runs `golangci-lint`; `make vulncheck` runs `govulncheck`; `make validate` runs all three.
- `make release --snapshot` builds cross-platform artifacts via GoReleaser (outputs to `dist/`).

## Coding Style & Naming Conventions
- Go 1.25.6; format with `gofmt` and imports via `goimports` (or `golangci-lint`’s default linters).
- Package names are lowercase and align with folder names; exported identifiers should be descriptive and avoid abbreviations.
- CLI commands follow verb-first naming (`feed add`, `feed refresh`, `models ls/ps`).

## Testing Guidelines
- Place tests alongside code as `_test.go` files; prefer table-driven cases for adapters and UI helpers.
- Target meaningful coverage for adapters and config loading; mock external services (Ollama, feeds) to keep tests deterministic.
- Run `make validate` before opening a PR to catch style, vuln, and test issues together.

## Commit & Pull Request Guidelines
- Follow the short, present-tense style seen in history (e.g., `Some usage improvements...`). Keep scope focused per commit.
- PRs should describe behavior changes, note config impacts (e.g., `config.yaml` defaults), and include reproduction steps for CLI commands. Attach logs or screenshots if UI output changes.
- Link related issues and mention any manual steps (migrating `data/feeds.db`, regenerating `dist/` artifacts) in the PR description.

## Configuration & Security Tips
- Keep secrets and host endpoints out of commits; use local overrides of `config.yaml` or environment variables when running the CLI.
- Avoid committing regenerated `data/feeds.db` or `dist/` artifacts unless intentionally publishing a release.
