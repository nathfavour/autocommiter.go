# autocommiter-maintenance

Specialized skill for cleanup and uninstallation.

## Instructions

Use this skill when the user needs to clean up data or uninstall the tool.

### Workflows

#### 1. Cleanup
- `autocommiter clean`: Wipes the `~/.autocommiter` data directory, including the SQLite index and cached model list.

#### 2. Uninstallation
- `autocommiter uninstall`: Removes the binary.
- `autocommiter uninstall --clean`: Removes the binary AND all configuration data.

### Key Commands
- `autocommiter clean`
- `autocommiter uninstall [--clean]`
