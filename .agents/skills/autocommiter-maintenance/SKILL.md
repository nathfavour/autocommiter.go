# autocommiter-maintenance

Specialized skill for lifecycle management, updates, and cleanup.

## Instructions

Use this skill when the user needs to update Autocommiter, clean up data, or uninstall the tool.

### Workflows

#### 1. Self-Update
- `autocommiter update`: Manually trigger the update process.
- **Modes**: Supports both binary download and "Build from Source". It automatically cleans up conflicting binaries in the `$PATH`.

#### 2. Cleanup
- `autocommiter clean`: Wipes the `~/.autocommiter` data directory, including the SQLite index and cached model list.

#### 3. Uninstallation
- `autocommiter uninstall`: Removes the binary.
- `autocommiter uninstall --clean`: Removes the binary AND all configuration data.

### Key Commands
- `autocommiter update`
- `autocommiter clean`
- `autocommiter uninstall [--clean]`
