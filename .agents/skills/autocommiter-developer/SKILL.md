# autocommiter-developer

Specialized skill for contributors and developers working on the Autocommiter project itself.

## Instructions

Use this skill when modifying the Autocommiter codebase or setting up a developer environment.

### Workflows

#### 1. Dev Mode Detection
- Autocommiter automatically enables `build_from_source` if it detects both `go` and `git` in the environment.
- This ensures that `autocommiter update` compiles the latest source instead of downloading a binary.

#### 2. Building and Testing
- Build: `go build -o autocommiter ./cmd/autocommiter`
- Test: `go test ./internal/...`

#### 3. Internal Architecture
- **Inference**: Logic in `internal/api` and `internal/summarizer`.
- **Auth**: Logic in `internal/auth` and `internal/processor/account.go`.
- **Database**: SQLite index managed in `internal/index`.
- **Networking**: Custom resolver in `internal/netutil` for Termux/cross-platform resilience.

### Development Resources
- `ARCHITECTURE.md`: Technical overview.
- `TODO.md`: Roadmap and pending tasks.
