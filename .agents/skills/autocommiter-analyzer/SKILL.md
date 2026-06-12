# autocommiter-analyzer

Specialized skill for analyzing repository state and summarizing changes.

## Instructions

Use this skill when the user wants to understand what changes are staged or when they need to inspect the repository's configuration health.

### Workflows

#### 1. Summarize Staged Changes
- Use `autocommiter summarize` to get a JSON summary of all staged changes.
- This is useful for providing context to other AI agents or for generating reports.

#### 2. Repository Health Check
- Use `autocommiter analyze` to check if the current Git identity (name/email) matches the authenticated GitHub account.
- Use `autocommiter analyze --apply` to automatically fix identity mismatches.

#### 3. Repository Discovery
- Use `autocommiter list-repos` to find all git repositories within a directory tree.
- Useful for batch operations across multiple projects.

### Key Commands
- `autocommiter summarize [-r <repo>]`
- `autocommiter analyze [-r <repo>] [--apply]`
- `autocommiter list-repos [-r <path>] [--json]`
