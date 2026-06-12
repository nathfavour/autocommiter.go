# autocommiter-config

Specialized skill for managing Autocommiter's configuration, AI models, and preferences.

## Instructions

Use this skill when the user wants to change settings, update API keys, or switch between AI models.

### Workflows

#### 1. Configuration Levels
- **Global**: Stored in `~/.autocommiter/config.json`.
- **Project-Level**: Create a `.autocommiter.json` in the repo root to override global settings for that specific project.
- **Key Fields**: `selected_model`, `enable_gitmoji`, `update_gitignore`, `prefer_noreply_email`, `gitignore_patterns`.

#### 2. Setup Authentication
- Use `autocommiter set-api-key [KEY]` to manually set a GitHub Models API key.
- Remind the user that `gh auth login` is also supported and preferred for zero-config.

#### 3. Model Management
- `autocommiter list-models`: List chat-completion models from GitHub.
- `autocommiter select-model`: Interactive selection.
- `autocommiter refresh-models`: Fetch the latest models from the Inference API and cache them.

#### 4. Preference Toggling
- `autocommiter toggle-gitmoji`: Enable/disable ✨ emojis.
- `autocommiter toggle-skip-confirmation`: Skip "Proceed with commit?" prompts.
- `autocommiter toggle-auto-update`: Background update checks.
- `autocommiter toggle-fork-sync`: Sync fork after push.

### Key Commands
- `autocommiter get-config`
- `autocommiter set-api-key [KEY]`
- `autocommiter select-model`
- `autocommiter toggle-gitmoji`
