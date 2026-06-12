# autocommiter-core

Specialized skill for managing the core AI-powered commit generation process of Autocommiter.

## Instructions

Use this skill when the user wants to generate commit messages, stage changes, or perform the full autocommit workflow.

### Workflows

#### 1. Generate and Apply Commit
- Ensure changes are staged (or let `autocommiter` handle it).
- Run `autocommiter generate` to generate a message and commit.
- Use `--no-push` if the user doesn't want to push immediately.
- Use `--force` to skip the confirmation prompt.
- **Batch Processing**: The `-r/--repo` flag supports comma-separated paths (e.g., `-r repo1,repo2`) to process multiple repositories at once.

#### 2. Generate Message Only
- Run `autocommiter generate-message` to see what the AI suggests without committing.
- This uses a fixed system prompt optimized for **Conventional Commits**.

#### 3. Prepare Repository
- Run `autocommiter prepare` to stage all changes and ensure `.gitignore` safety.
- It will automatically add critical patterns (like `.env`) if `update_gitignore` is enabled.

### Key Commands
- `autocommiter generate [-r <repo(s)>] [-n] [-f] [-u <user>]`
- `autocommiter generate-message [-r <repo>]`
- `autocommiter prepare [-r <repo>]`
