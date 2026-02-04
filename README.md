# st

A Git CLI tool for managing stacked branches, inspired by [Graphite](https://graphite.dev). Built with Go, Cobra, and Bubbletea.

Stacked branches let you break large changes into small, reviewable PRs that build on each other — without waiting for each one to merge before starting the next.

## Install

Requires [Go](https://go.dev/dl/):

```
git clone https://github.com/rodrigolobo25/st.git
cd st
./install.sh
```

The script installs the `st` binary to `/usr/local/bin`. Override with `INSTALL_DIR=~/.local/bin ./install.sh`.

If Go isn't installed and you have Homebrew, the script will offer to install it for you.

## Quick start

```bash
# Initialize in a git repo
st init

# Create a stacked branch
st create feat-auth
# ... make changes ...
st modify -acm "add auth layer"

# Stack another branch on top
st create feat-auth-ui
# ... make changes ...
st modify -acm "add auth UI"

# See the full tree
st log
# main
# └── feat-auth      1 commit
#     └── feat-auth-ui  1 commit  ← you are here

# Navigate the stack
st down        # move toward trunk
st up          # move toward leaves
st top         # jump to leaf
st bottom      # jump to stack root
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `st init` | | Set up st in a git repo (auto-detects `main`/`master`) |
| `st create <name>` | | Create a new branch stacked on the current one |
| `st log` | `st ls` | Show the stack tree with commit counts and status |
| `st up [n]` | | Move n branches away from trunk (default 1) |
| `st down [n]` | | Move n branches toward trunk (default 1) |
| `st top` | | Jump to the leaf of the current stack |
| `st bottom` | | Jump to the first branch above trunk |
| `st modify` | `st m` | Amend HEAD or create a new commit |
| `st restack` | | Rebase all branches in the stack onto their parents |
| `st continue` | | Resume restacking after resolving conflicts |
| `st delete [name]` | | Remove a branch and reparent its children |
| `st switch` | `st sw` | Interactive TUI branch picker |
| `st sync` | | Fetch, fast-forward trunk, clean merged branches, restack |
| `st branch` | `st b` | Show info about the current branch |

## Workflow

```bash
# Start a stack
st init
st create feat-1
# ... work, commit ...
st modify -acm "first feature"

st create feat-2
# ... work, commit ...
st modify -acm "second feature"

# Go back and update feat-1
st down
# ... make changes ...
st modify -acm "update first feature"

# Rebase feat-2 onto the updated feat-1
st restack

# Interactive branch switcher
st switch

# Sync with remote (fetch, ff trunk, clean merged, restack)
st sync

# Delete a branch (children get reparented)
st delete feat-1
```

## `st modify` flags

| Flag | Description |
|------|-------------|
| `-a` | Stage all changes |
| `-m "msg"` | Set commit message |
| `-c` | Create a new commit instead of amending |

Combine them: `st modify -acm "message"` stages everything and creates a new commit.

## How it works

All metadata is stored in `.git/config` using `git config --local` — no extra files, no external services:

```ini
[st]
    trunk = main

[stack "feat-auth"]
    parent = main

[stack "feat-auth-ui"]
    parent = feat-auth
```

Branches whose parent is trunk are stack roots. A "stack" is the tree rooted at each root branch.

`st restack` walks the tree bottom-up and runs `git rebase --onto` for each branch that has diverged from its parent. If a conflict occurs, it saves state so you can resolve and run `st continue`.
