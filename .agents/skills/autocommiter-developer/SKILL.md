# autocommiter-developer

Specialized skill for contributors and developers working on the Autocommiter project itself.

## Instructions

Use this skill when modifying the Autocommiter codebase or setting up a developer environment.

### Workflows

#### 1. Dev Mode Detection
- Autocommiter automatically enables `build_from_source` if it detects both `go` and `git` in the environment.
- This ensures that `autocommiter update` compiles the latest source instead of downloading a binary.

#### 2. Building and Testing
- Build: `go build -o autocommiter ./cmd/autocommiter` (Remember: per `AGENTS.md`, always build into `bin/` for final binaries: `go build -o bin/autocommiter ./cmd/autocommiter`).
- Test: `go test ./internal/...`

#### 3. Internal Architecture
- **Inference (`internal/api`, `internal/summarizer`)**: Compresses git diffs (up to 12,000 chars) and calls the GitHub Models API.
- **Auth & Account Mgmt (`internal/auth`, `internal/processor/account.go`)**: Manages identity switching via `gh` CLI and reactive push recovery.
- **Database (`internal/index`)**: SQLite index at `~/.autocommiter/index.db`. Uses SHA256 hashes of repo paths. Features "Gravity" weights for account discovery.
- **Networking (`internal/netutil`)**: Custom HTTP client with timeout and resilience for various environments (including Termux).
- **Anyisland (`internal/anyisland`)**: UDP-based auto-registration and Unix-socket-based management status.

### Development Resources
- `ARCHITECTURE.md`: Technical overview.
- `TODO.md`: Roadmap and pending tasks.
- `CONTRIBUTING.md`: Guidelines for PRs and issues.
