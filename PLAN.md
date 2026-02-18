# GitAll — Go Rewrite Plan

## 1. Existing Features & Use Cases

### 1.1 CLI Interface

| Usage | Description |
|-------|-------------|
| `gitall {action}` | Run action against all accounts in config file |
| `gitall {action} {user}` | Run action for a single user, current directory, SSH protocol |
| `gitall {action} {user} {dir}` | Run action for a single user in a specific directory, SSH protocol |
| `gitall {action} {user} {dir} {protocol}` | Run action with all params specified |

**Actions**: `clone`, `pull`, `status`, `config`

### 1.2 Clone

- Fetches the full list of public repos for a GitHub user/org via the GitHub API (`GET /users/{user}/repos?per_page=100`)
- Clones each repo into the target directory using the configured protocol
- Skips repos that already exist locally (directory check)
- Clones into a **lowercase** directory name (`repo.toLowerCase()`)
- Logs each clone operation to the log file

### 1.3 Pull

- Scans the target directory for subdirectories containing a `.git` folder
- For each discovered git repo, checks the `remote.origin.url`
- Only pulls if the remote URL matches the configured user's SSH or HTTPS base URL (skips repos belonging to other users/orgs)
- Logs each update to the log file
- Prints total count of updated repos

### 1.4 Status

- Scans the target directory for subdirectories containing a `.git` folder
- For each repo, runs `git status --porcelain` and `git log origin/master..HEAD`
- Only prints output for repos that have uncommitted changes or unpushed commits
- Displays a separator line between repos for readability

### 1.5 Config

- Displays the contents of `~/.gitall/config.json`
- Reports if no config file exists

### 1.6 Configuration File (`~/.gitall/config.json`)

Supports multiple accounts in a JSON array:

```json
[
  {
    "username": "BoyCook",
    "dir": "/Users/boycook/code/boycook",
    "protocol": "ssh",
    "active": true
  },
  {
    "username": "SomeOrg",
    "dir": "/Users/boycook/code/org",
    "protocol": "https",
    "active": false
  }
]
```

**Fields**:
- `username` — GitHub user or organisation name (case-sensitive)
- `dir` — target directory for cloned repos
- `protocol` — `ssh`, `https`, or `svn`
- `active` — optional boolean (defaults to `true`); inactive accounts are skipped with a message

### 1.7 Protocol Support

| Protocol | Clone URL format |
|----------|-----------------|
| `ssh` | `git@github.com:{user}/{repo}.git` |
| `https` | `https://github.com/{user}/{repo}` |
| `svn` | `https://github.com/{user}/{repo}` |

### 1.8 Validation

- Target directory must exist
- Protocol must be one of `ssh`, `https`, `svn`
- Action must be one of `clone`, `pull`, `status`, `config`

### 1.9 Logging

- All operations logged to `~/.gitall/gitall.log`
- Console output for progress messages and status results

### 1.10 Repo Discovery

- Scans target directory for immediate child directories
- Identifies git repos by checking for a `.git` subdirectory
- Used by both `pull` and `status` actions

### 1.11 Pull Ownership Filter

- Before pulling, checks `remote.origin.url` against the configured user's SSH and HTTPS base URLs
- Skips repos that don't belong to the configured user (e.g. forks cloned manually from other users)
- Prints "Ignoring non users repository" message for skipped repos

---

## 2. Known Bugs

1. **Assignment instead of comparison in CLI argument parsing** (`bin/gitall:40,46,52`) — `args.length = 2` should be `args.length === 2`. This means only the 1-arg path (config file) ever works correctly; all other argument counts fall through to the first branch.

2. **`repo.toLowerCase()` called on object** (`lib/gitall.js:116`) — `repo` is a GitHub API response object, not a string. Should be `repo.name.toLowerCase()`.

3. **No GitHub API error handling** — if the API call fails (network error, 404, rate limit), the error is silently swallowed and no repos are cloned.

4. **No pagination** — hardcoded `per_page=100` means users/orgs with >100 repos will have repos silently missed.

5. **Status assumes `origin/master`** — fails for repos using `main` or other default branch names.

6. **Pull doesn't check local state before pulling** — if a repo has uncommitted changes, `git pull` can fail or create merge conflicts. Similarly, if there are unpushed local commits, pull can cause unwanted merges or conflicts. There is no pre-flight check, no stash, and no warning — it just runs `git pull` and hopes for the best.

7. **No feedback on pull failures** — when a pull fails (due to conflicts, dirty working tree, etc.), the error output goes to the log file but the user gets no clear indication of which repos failed or why.

8. **Synchronous file I/O** — `readdirSync`, `appendFileSync`, `readFileSync` block the event loop.

---

## 3. Features & Improvements for Go Rewrite

### 3.1 Bug Fixes (all existing bugs from Section 2)

- Fix all known bugs listed above
- Status should detect the default branch dynamically (`git symbolic-ref refs/remotes/origin/HEAD`) rather than assuming `master`
- Pull must check for dirty working tree and unpushed commits before pulling
- Repos with dirty state should be skipped with a clear warning, not silently fail
- Full GitHub API pagination to handle accounts with >100 repos
- Proper error handling and reporting for API failures, git failures, etc.

### 3.2 CLI Improvements

- Use proper subcommand-style CLI (`gitall clone`, `gitall pull`, etc.) with flags instead of positional args
- Add `--user`, `--dir`, `--protocol` flags on each subcommand for ad-hoc usage
- Add `--concurrency` / `-j` flag to control parallelism (default: sensible limit, e.g. 4)
- Add `--dry-run` flag to show what would happen without executing
- Add `--verbose` / `-v` flag for detailed output
- Add `--quiet` / `-q` flag to suppress non-essential output
- Add `--version` flag
- Add `--help` for each subcommand with usage examples
- Coloured terminal output (repo names, success/failure/skipped indicators)

### 3.3 Clone Improvements

- Parallel cloning with configurable concurrency
- Progress output showing `[3/47] Cloning repo-name...`
- Support cloning private repos (requires auth token)
- Option to include/exclude forks (`--no-forks`, `--forks-only`)
- Option to filter repos by name pattern (`--filter "prefix-*"`)
- Option to include/exclude archived repos (`--no-archived`)
- Summary at end: `Cloned: 12, Skipped (exists): 35, Failed: 0`

### 3.4 Pull Improvements

- Pre-flight check: skip repos with uncommitted changes (with clear warning)
- Pre-flight check: skip repos with unpushed commits (with clear warning)
- Parallel pulling with configurable concurrency
- Progress output showing current repo and outcome
- `--stash` flag: automatically stash changes before pull and pop after
- `--rebase` flag: use `git pull --rebase` instead of merge
- Remove ownership filter by default (pull all repos in dir); add `--owned-only` flag to restore old behaviour
- Summary at end: `Updated: 10, Skipped (dirty): 3, Skipped (unpushed): 2, Failed: 1, Up-to-date: 31`

### 3.5 Status Improvements

- Detect default branch dynamically instead of assuming `master`
- Show more detail: branch name, ahead/behind counts, staged/unstaged/untracked counts
- Compact single-line-per-repo format by default, `--verbose` for full detail
- Colour coding: green (clean), yellow (dirty), red (diverged)
- Summary counts at end

### 3.6 Config Improvements

- `gitall config init` — create a default config file interactively
- `gitall config add` — add a new account to the config
- `gitall config remove` — remove an account from the config
- `gitall config list` — display current config in a readable table
- Validate config on load (check required fields, valid protocols, directories exist)
- Support `~` and `$HOME` expansion in directory paths
- YAML or TOML as config format (more human-friendly than JSON)

### 3.7 Authentication

- Support GitHub personal access token for private repos
- Token stored in config or read from `GITHUB_TOKEN` env var
- Support GitHub Enterprise (configurable API base URL per account)

### 3.8 Concurrency & Performance

- All repo operations (clone, pull, status) run in parallel using goroutines
- Configurable concurrency limit to avoid overwhelming the system
- Progress bar or live-updating output showing overall progress

### 3.9 Output & Reporting

- Clear, structured console output with colour
- Summary report after each action (counts of success/skip/fail)
- `--json` flag for machine-readable output
- Log file with timestamps and structured entries

### 3.10 New Actions

- `gitall fetch` — run `git fetch` on all repos (lighter than pull, useful for checking remote state)
- `gitall list` — list all local repos with their remote URL, branch, and clean/dirty state (no git operations, just local inspection)

---

## 4. Go Architecture

### 4.1 Project Layout

```
gitall/
├── main.go                     # Entry point — initialises cobra root command
├── go.mod
├── go.sum
├── cmd/                        # CLI layer — one file per subcommand
│   ├── root.go                 # Root command, global flags (--verbose, --quiet, --json)
│   ├── clone.go                # `gitall clone` subcommand
│   ├── pull.go                 # `gitall pull` subcommand
│   ├── fetch.go                # `gitall fetch` subcommand
│   ├── status.go               # `gitall status` subcommand
│   ├── list.go                 # `gitall list` subcommand
│   └── config.go               # `gitall config [init|add|remove|list]` subcommands
├── internal/
│   ├── config/                 # Config loading, validation, path expansion
│   │   └── config.go
│   ├── github/                 # GitHub API client — list repos, pagination, auth
│   │   └── client.go
│   ├── git/                    # Git operations — clone, pull, fetch, status
│   │   └── git.go
│   ├── runner/                 # Concurrent execution engine — worker pool, progress
│   │   └── runner.go
│   └── output/                 # Output formatting — colour, JSON, summaries
│       └── output.go
└── testdata/                   # Test fixtures (fake git repos, config files)
```

### 4.2 Key Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | CLI framework (subcommands, flags, help) |
| `github.com/fatih/color` | Coloured terminal output |
| `gopkg.in/yaml.v3` | YAML config file parsing |

No git library — shell out to `git` directly. Keeps it simple, avoids CGO, and users already have git installed. All git interaction goes through `internal/git/` so it's easy to test via interface.

### 4.3 Package Responsibilities

**`cmd/`** — Thin CLI layer. Each file defines a cobra command, parses flags, loads config, calls into `internal/` packages. No business logic here.

**`internal/config/`** — Loads and validates `~/.gitall/config.yaml`. Handles path expansion (`~`, `$HOME`). Merges CLI flags with config file values (flags take precedence).

**`internal/github/`** — GitHub API client. Lists repos for a user/org with full pagination. Supports auth token. Supports filtering (forks, archived, name pattern). Returns a clean `[]Repo` slice.

**`internal/git/`** — Runs git commands against local repos. Each function (Clone, Pull, Fetch, Status) takes a path and options, returns structured results. Handles all pre-flight checks (dirty state, unpushed commits). Detects default branch dynamically.

**`internal/runner/`** — Generic concurrent task runner. Takes a slice of work items and a function, runs them across a worker pool with configurable concurrency. Reports progress and collects results. Used by clone, pull, fetch, and status commands.

**`internal/output/`** — Handles all output formatting. Supports three modes: normal (coloured text), quiet (minimal), and JSON. Prints progress lines, summaries, and error reports.

### 4.4 Data Flow

```
CLI flags + config.yaml
        │
        ▼
   cmd/clone.go         ← parses flags, loads config, builds account list
        │
        ▼
   github.ListRepos()   ← fetches repos from API (with pagination + filters)
        │
        ▼
   runner.Run()          ← runs git.Clone() for each repo concurrently
        │
        ▼
   output.Summary()      ← prints results summary
```

### 4.5 Key Types

```go
// Config types
type Config struct {
    Accounts []Account `yaml:"accounts"`
}

type Account struct {
    Username string `yaml:"username"`
    Dir      string `yaml:"dir"`
    Protocol string `yaml:"protocol"` // ssh, https
    Token    string `yaml:"token"`    // optional, overrides GITHUB_TOKEN
    APIURL   string `yaml:"api_url"`  // optional, for GitHub Enterprise
    Active   *bool  `yaml:"active"`   // optional, defaults to true
}

// GitHub types
type Repo struct {
    Name     string
    CloneURL string // built from protocol
    Fork     bool
    Archived bool
}

// Git operation results
type RepoResult struct {
    Name    string
    Path    string
    Status  ResultStatus // Success, Skipped, Failed
    Message string       // human-readable detail
}

type ResultStatus int
const (
    Success ResultStatus = iota
    Skipped
    Failed
    UpToDate
)

// Runner
type Task[T any] struct {
    Item    T
    Execute func(T) RepoResult
}
```

### 4.6 Config File Format (`~/.gitall/config.yaml`)

```yaml
accounts:
  - username: BoyCook
    dir: ~/code/boycook
    protocol: ssh

  - username: SomeOrg
    dir: ~/code/org
    protocol: https
    token: ghp_xxxxxxxxxxxx  # optional per-account token
    active: false
```

Global token fallback: `GITHUB_TOKEN` env var.

---

## 5. Implementation Plan

### Phase 1 — Scaffold & Core

1. Initialise Go module, install cobra, set up project layout
2. Implement `cmd/root.go` with global flags (`--verbose`, `--quiet`, `--json`, `--version`)
3. Implement `internal/config/` — load YAML, validate, expand paths, merge with CLI flags
4. Implement `cmd/config.go` — `config list` and `config init` subcommands
5. Add tests for config loading and validation

### Phase 2 — GitHub API & Clone

6. Implement `internal/github/` — list repos with pagination, auth token, fork/archived filtering
7. Implement `internal/git/` — `Clone()` function (shell out to `git clone`)
8. Implement `internal/runner/` — concurrent worker pool with progress reporting
9. Implement `internal/output/` — coloured text output, summary formatting
10. Implement `cmd/clone.go` — wire it all together with `--dry-run`, `--no-forks`, `--no-archived`, `--filter`, `-j`
11. Add tests for GitHub client (mock HTTP), git clone, runner

### Phase 3 — Status & List

12. Implement `internal/git/` — `Status()` function (porcelain status, ahead/behind, branch detection)
13. Implement `cmd/status.go` — concurrent status with coloured compact output
14. Implement `cmd/list.go` — local repo inspection (remote URL, branch, clean/dirty)
15. Add tests for status parsing and list output

### Phase 4 — Pull & Fetch

16. Implement `internal/git/` — `Pull()` with pre-flight checks (dirty, unpushed), `--stash`, `--rebase`
17. Implement `internal/git/` — `Fetch()` function
18. Implement `cmd/pull.go` — concurrent pull with summaries, `--owned-only`, `--stash`, `--rebase`
19. Implement `cmd/fetch.go` — concurrent fetch with summaries
20. Add tests for pull pre-flight checks, fetch

### Phase 5 — Config Management & Polish

21. Implement `cmd/config.go` — `config add` and `config remove` subcommands
22. Add `--json` output mode across all commands
23. Add GitHub Enterprise support (custom API URL per account)
24. End-to-end testing with real git repos in `testdata/`
25. README and install instructions
