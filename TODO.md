# TODO: Intelligent GitHub Account Management

## Core Objective
Enable the tool to automatically detect, index, and switch between multiple GitHub accounts based on the current repository context, ensuring seamless commits and pushes without manual configuration or `gh auth switch` intervention.

## Implementation Details
- **Database:** SQLite (`modernc.org/sqlite`) stored in `~/.autocommiter/index.db`.
- **Repo Identification:** Parse `git remote get-url origin` to extract the owner.
- **Account Discovery:** Use `gh api user` and potentially parse `gh auth status` or `hosts.yml` to identify available accounts.
- **Email Retrieval:** Use `gh api user/emails` to find the primary verified email for the active account.
- **Git Sync:** Automatically set `git config --local user.email` and `user.name` to match the active GitHub account.

## Tasks
- [x] **Deep Analysis**
- [ ] **Infrastructure**
    - [ ] Add SQLite dependency (`modernc.org/sqlite`).
    - [ ] Create `internal/index` package for database management.
    - [ ] Define SQLite schema for repositories and accounts.
- [ ] **GitHub CLI Integration**
    - [ ] Implement `ListAccounts()` in `internal/auth`.
    - [ ] Implement `SwitchAccount(user string)` in `internal/auth`.
    - [ ] Implement `GetAccountDetails(user string)` in `internal/auth` (name, email).
- [ ] **Git Intelligence**
    - [ ] Implement `GetRepoOwner(path string)` in `internal/git`.
    - [ ] Implement `SyncGitConfig(path string, name string, email string)` in `internal/git`.
- [ ] **Workflow Integration**
    - [ ] Create `processor.SyncAccount(repoRoot)` function.
    - [ ] Hook `SyncAccount` into `ProcessSingleRepo` before staging/committing.
- [ ] **Testing**
    - [ ] Verify switching between multiple logged-in accounts.
    - [ ] Verify local git config updates.
