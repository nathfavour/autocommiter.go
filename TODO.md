# TODO: Intelligent GitHub Account Management

## Core Objective
Enable the tool to automatically detect, index, and switch between multiple GitHub accounts based on the current repository context, ensuring seamless commits and pushes without manual configuration or `gh auth switch` intervention.

## Architectural Strategy: "Hidden Intelligence"
- **Sentinel Optimization (Fast-Exit):** If only one account is logged in, skip all logic (1ms `stat` check).
- **Parallel Discovery:** Run account verification in a background routine while the AI generates the commit message to hide latency.
- **Affinity Mapping:** Priority for account selection: 
    1. SQLite Cache (Successful previous push)
    2. Local Git Config (`user.email`)
    3. Local Git History (Last committer email)
    4. Remote URL Owner (Fallback)
- **Reactive Safety Net:** If `push` fails due to auth, iterate through other logged-in accounts and retry.

## Tasks
- [ ] **Infrastructure**
    - [ ] Create `~/.autocommiter/.single_account` sentinel logic.
    - [ ] Setup SQLite (`modernc.org/sqlite`) in `~/.autocommiter/index.db`.
    - [ ] Define cache schema: `repo_path_hash`, `account_handle`, `email`, `name`.
- [ ] **GitHub CLI Integration (`internal/auth`)**
    - [ ] `ListAccounts()`: Parse `gh auth status` or `hosts.yml`.
    - [ ] `GetAccountIdentity(handle)`: Get primary verified email and display name via `gh api`.
    - [ ] `SwitchAccount(handle)`: Execute `gh auth switch`.
- [ ] **Git Intelligence (`internal/git`)**
    - [ ] `GetLocalIdentity()`: Read `user.email` and `user.name`.
    - [ ] `GetHistoryIdentity()`: Get email from `git log -1`.
    - [ ] `SyncLocalConfig(email, name)`: Set local git config.
- [ ] **Workflow Integration (`internal/processor`)**
    - [ ] Implement `AccountManager` to run discovery in parallel with AI generation.
    - [ ] Implement `SyncAccount` hook before commit/push.
    - [ ] Implement `ReactivePush` to handle auth failures and retry with different accounts.
- [ ] **Verification**
    - [ ] Benchmark to ensure <5ms impact for single-account users.
