# Document Feed Embedder - Gemini Context

This project is a Go-based Proof of Concept (POC) for building a local knowledge base using Retrieval-Augmented Generation (RAG). It ingests RSS/OPDS feeds, generates embeddings using Ollama, stores them in a local vector database, and provides a CLI/TUI to query the data.

## Project Overview

- **Purpose:** Demonstrate a local RAG pipeline using RSS feeds as the data source.
- **Core Stack:**
  - **Language:** Go 1.26.0
  - **LLM/Embeddings:** [Ollama](https://ollama.com/) (local API)
  - **Vector DB:** [chromem-go](https://github.com/philippgille/chromem-go) (Persistent SQLite-based vector storage)
  - **Metadata DB:** [Storm](https://github.com/asdine/storm) (BoltDB wrapper for Go)
  - **CLI/TUI:** [Cobra](https://github.com/spf13/cobra), [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss)
  - **Telemetry:** OpenTelemetry (Traces, Metrics, Logs)
  - **Build Tool:** Makefile, GoReleaser

## Architecture

- `cmd/cli/`: The main entry point for the CLI application. Subcommands are defined in `cmd/cli/cmd/`.
- `cmd/benchmarker/`: A tool to compare different LLM models' performance and output quality.
- `internal/adapter/`: Orchestrates the flow between feeds, persistence, and Ollama. `FeedAdapter` is the central component.
- `internal/client/ollama/`: Custom HTTP client for interacting with Ollama's API (`/api/embed`, `/api/generate`, `/api/chat`, etc.).
- `internal/persistence/`:
  - `chromem/`: Handles vector storage, document splitting (using `langchaingo`), and embedding generation calls.
  - `storm/`: Handles metadata persistence (feed info, article details, answer cache).
- `internal/ui/`: Contains TUI screens and components using the Bubble Tea framework.
- `internal/feed/`: RSS and OPDS parsing logic using `gofeed`.

## Key Files & Configuration

- `config.yaml`: Central configuration for Ollama endpoints, models, chunk sizes, and telemetry.
- `go.mod`: Defines dependencies and the Go 1.26.0 toolchain.
- `Makefile`: Provides shortcuts for common development tasks.
- `internal/persistence/chromem/db.go`: Core logic for managing the vector database.
- `internal/adapter/feed.go`: Implements the RAG flow (`AskAQuestion`, `RefreshFeed`).

## Development Workflow

### Building and Running

- **Build all binaries:** `make release` (requires GoReleaser) or `go build ./cmd/cli` and `go build ./cmd/benchmarker`.
- **Add default feeds:** `make add`
- **Refresh feeds:** `make refresh`
- **Ask a question (RAG):** `make ask` or `go run ./cmd/cli ask "your question"`
- **Run benchmark:** `make benchmark`

### Testing and Validation

- **Run all tests:** `make test`
- **Run linter:** `make linter` (requires `golangci-lint`)
- **Check vulnerabilities:** `make vulncheck`
- **Full validation:** `make validate` (runs test, linter, and vulncheck)

### Development Conventions

- **Cobra Commands:** Follow the `verb-first` naming convention (e.g., `feed add`, `models ls`).
- **Persistence:** Use `storm` for structured data and `chromem` for vector data.
- **Error Handling:** Use `fmt.Errorf` with `%w` for error wrapping.
- **Logging:** Use the internal logging package (based on `go-kit/log`) which integrates with telemetry.
- **Telemetry:** Ensure new features include appropriate spans and metrics using the provided telemetry wrappers.

## Future Considerations

- **CI/CD:** GitHub Actions workflow is present (`.github/workflows/ci.yml`) but needs to be aligned with the latest build/test commands.
- **Caching:** A simple answer cache is implemented in `FeedAdapter` to avoid redundant LLM calls for identical questions.
- **Embedding Models:** Supports any Ollama-compatible embedding model; default is `nomic-embed-text`.
