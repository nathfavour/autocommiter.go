# autocommiter-auth

Specialized skill for managing GitHub accounts, identity syncing, and repairing commit authorship.

## Instructions

Use this skill when the user needs to switch between multiple GitHub accounts or fix authorship issues.

### Workflows

#### 1. Account Discovery Heuristics
Autocommiter uses a multi-stage discovery process to find the right account for a repo:
1. **Default User**: Checks the SQLite index for a manual override.
2. **Gravity**: Checks if a parent directory name matches a logged-in account.
3. **Affinity**: Checks history identity, local Git config, and remote owner.

#### 2. User/Account Repair
- Use `autocommiter fix --user <username>` to repair the last commit.
- **Under the hood**: It switches the GH account, syncs local Git config (name/email), amends the commit author, and force-pushes with lease safety.

#### 3. Privacy Settings
- If `prefer_noreply_email` is enabled, Autocommiter will use the GitHub `<id>+<user>@users.noreply.github.com` format.

#### 4. Fork Syncing
- `autocommiter sync [USER]`: Manually sync a fork.
- `autocommiter set-fork-user [USER]`: Set default target for fork syncs.

### Key Commands
- `autocommiter fix --user <username> [-r <repo>]`
- `autocommiter sync [USER] [-r <repo>]`
- `autocommiter set-fork-user [USER]`
