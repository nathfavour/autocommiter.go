# autocommiter-vibe

Specialized skill for Anyisland integration and Vibe manifests.

## Instructions

Use this skill when the project is being used within the Anyisland ecosystem or when managing Vibe toolsets.

### Workflows

#### 1. Manifest Generation
- Use `autocommiter vibe-manifest` to output the tool's capabilities in a format compatible with `vibeauracle`.
- This includes schema definitions for `generate_commit_message` and `summarize_changes`.

#### 2. Tool Execution
- Use `autocommiter execute [tool] [args]` to run autocommiter functions in "vibe mode" (JSON-in, JSON-out).
- Supported tools: `generate_commit_message`, `summarize_changes`.

### Key Commands
- `autocommiter vibe-manifest`
- `autocommiter execute <tool> <json_params>`
