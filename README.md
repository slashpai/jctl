# jctl

A fast CLI for managing Jira Cloud issues from your terminal.

## Why jctl?

Atlassian's [ACLI](https://developer.atlassian.com/cloud/acli) ships hundreds of commands for every Atlassian product. jctl takes the opposite approach — it covers only the Jira workflows you use daily, in a single lightweight binary.

## Install

```bash
go install github.com/slashpai/jctl@latest
```

Or build from source:

```bash
git clone https://github.com/slashpai/jctl.git
cd jctl
go install .
```

This installs `jctl` to your `$GOPATH/bin`. Make sure it's in your `PATH`.

Alternatively, build locally and move it to a directory in your `PATH`:

```bash
make build
sudo mv jctl /usr/local/bin/
```

## Setup

1. Generate a Jira API token at [https://id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Run the configure command:

```bash
jctl configure
```

You can also set credentials via environment variables:

```bash
export JCTL_BASE_URL=https://yourorg.atlassian.net
export JCTL_EMAIL=you@example.com
export JCTL_TOKEN=your-api-token
```

## Shell completion

jctl supports tab completion for commands, subcommands, and flags in bash, zsh, and fish. Ensure `jctl` is on your `$PATH` (not invoked as `./jctl`).

**Current session** — pick your shell:

```bash
# Bash
source <(jctl completion -c bash)

# Zsh
source <(jctl completion -c zsh)

# Fish
jctl completion -c fish | source
```

**Permanent setup** — add the same command to your shell init file:

| Shell | Init file                      |
| ----- | ------------------------------ |
| Bash  | `~/.bashrc`                    |
| Zsh   | `~/.zshrc`                     |
| Fish  | `~/.config/fish/config.fish`   |

Run `jctl completion` without flags for setup instructions tailored to your shell.

## Usage

### Create an issue

```bash
jctl issue create -p PROJ -s "Fix login bug" -t Bug --priority High
jctl issue create -p PROJ -s "New feature" -d "Detailed description" -l backend -l urgent
jctl issue create -p PROJ -s "Team task" --assignee none    # leave unassigned
```

### View an issue

```bash
jctl issue view PROJ-123
```

### Update an issue

```bash
jctl issue update PROJ-123 -s "Updated summary"
jctl issue update PROJ-123 --priority High --assignee user@example.com
jctl issue update PROJ-123 --assignee me
jctl issue update PROJ-123 -c "Adding a comment"
```

### List issues

```bash
jctl issue list -p PROJ
jctl issue list -p PROJ --status "In Progress" -a me    # 'me' refers to the current authenticated user
jctl issue list --jql "project = PROJ AND priority = High" -n 50
```

### Transition an issue

```bash
jctl issue transition PROJ-123 --list              # see available transitions
jctl issue transition PROJ-123 -s "In Progress"    # move to status
```

## Command Reference

| Command                 | Description                           |
| ----------------------- | ------------------------------------- |
| `jctl configure`        | Set Jira URL, email, and API token    |
| `jctl completion`       | Generate shell tab completion scripts |
| `jctl issue create`     | Create a new issue                    |
| `jctl issue view`       | View issue details                    |
| `jctl issue update`     | Update issue fields or add a comment  |
| `jctl issue list`       | Search/list issues (JQL or filters)   |
| `jctl issue transition` | Move an issue to a new status         |

Run `jctl --help` or `jctl issue --help` for full flag details.
