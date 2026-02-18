# GitAll

Manage all your GitHub repositories across multiple accounts with a single command.

Clone, pull, fetch, and check status for every repo under your GitHub users and organisations — concurrently.

## Installation

### From source

```sh
go install github.com/boycook/gitall@latest
```

### Build locally

```sh
git clone https://github.com/BoyCook/GitAll.git
cd GitAll
make install
```

This builds the binary, installs it to `~/go/bin`, and adds it to your `PATH` if needed.

## Quick start

```sh
# Set up your config
gitall config init
# Edit ~/.gitall/config.yaml with your accounts, then:

# Clone all repos for all configured accounts
gitall clone

# Check status across all repos
gitall status

# Pull latest changes (skips dirty repos)
gitall pull
```

## Commands

### `gitall clone`

Clone all repositories for configured accounts from GitHub.

```sh
gitall clone                                  # all configured accounts
gitall clone --user BoyCook --dir ~/code      # single account, ad-hoc
gitall clone --no-forks --no-archived         # exclude forks and archived repos
gitall clone --filter "api-*"                 # only repos matching pattern
gitall clone --dry-run                        # show what would be cloned
gitall clone -j 8                             # 8 concurrent clones
```

**Flags:**
`--user`, `--dir`, `--protocol`, `--no-forks`, `--no-archived`, `--filter`, `--dry-run`, `-j`

### `gitall pull`

Pull latest changes for all local repositories. Repos with uncommitted changes or unpushed commits are safely skipped.

```sh
gitall pull                                   # all configured directories
gitall pull --dir ~/code/myorg                # specific directory
gitall pull --stash                           # auto-stash dirty repos
gitall pull --rebase                          # use git pull --rebase
gitall pull --owner BoyCook                   # only repos owned by this user
```

**Flags:**
`--user`, `--dir`, `--stash`, `--rebase`, `--owned-only`, `--owner`, `-j`

### `gitall fetch`

Fetch from all remotes without modifying your working tree. A safe way to check for updates.

```sh
gitall fetch                                  # all configured directories
gitall fetch --dir ~/code/myorg               # specific directory
```

**Flags:**
`--user`, `--dir`, `-j`

### `gitall status`

Show git status for all repositories. Only dirty repos are shown by default.

```sh
gitall status                                 # dirty repos only
gitall status --all                           # include clean repos
gitall status --dir ~/code/myorg              # specific directory
```

**Flags:**
`--user`, `--dir`, `--all`, `-j`

### `gitall list`

List all local repositories with branch, state, and remote URL. Local only — no network calls.

```sh
gitall list                                   # all configured directories
gitall list --dir ~/code/myorg                # specific directory
```

**Flags:**
`--dir`, `-j`

### `gitall config`

Manage the configuration file at `~/.gitall/config.yaml`.

```sh
gitall config init                            # create default config
gitall config list                            # display current config
gitall config add --username BoyCook --dir ~/code/boycook
gitall config remove BoyCook
gitall config discover --dir ~/code           # auto-generate from existing repos
gitall config discover --dir ~/code --dry-run # preview without writing
```

## Configuration

Config file: `~/.gitall/config.yaml`

```yaml
accounts:
  - username: BoyCook
    dir: ~/code/boycook
    protocol: ssh

  - username: MyOrg
    dir: ~/code/org
    protocol: https
    token: ghp_xxxxxxxxxxxx    # optional, for private repos
    active: false              # skip this account
```

**Fields:**

| Field | Required | Default | Description |
| --- | --- | --- | --- |
| `username` | yes | | GitHub user or organisation name |
| `dir` | yes | | Target directory for repos |
| `protocol` | no | `ssh` | `ssh` or `https` |
| `token` | no | | GitHub personal access token |
| `api_url` | no | `https://api.github.com` | GitHub Enterprise API URL |
| `active` | no | `true` | Set `false` to skip this account |

Tokens can also be set via the `GITHUB_TOKEN` environment variable (per-account `token` takes precedence).

## Global flags

All commands support:

| Flag | Description |
| --- | --- |
| `-v, --verbose` | Verbose output |
| `-q, --quiet` | Suppress non-essential output |
| `--json` | Machine-readable JSON output |
| `--version` | Print version |

## Examples

Clone all repos for a user (without config file):

```sh
gitall clone --user BoyCook --dir ~/code/boycook --protocol ssh
```

Check which repos have uncommitted work:

```sh
gitall status --dir ~/code
```

Safely update everything with auto-stash:

```sh
gitall pull --stash --rebase
```

Get JSON output for scripting:

```sh
gitall status --all --json | jq '.repos[] | select(.clean == false) | .name'
```

## Prerequisites

- Git must be installed and available on your `PATH`
- Go 1.21+ (for building from source)
