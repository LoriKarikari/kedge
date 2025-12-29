# Kedge

GitOps for Docker Compose.

## What it does

Kedge watches a Git repository and automatically deploys your Docker Compose application when changes are detected. It also monitors for drift and auto-heals when containers stop or diverge from the desired state.

## Features

- **Git sync** - Watches a branch and deploys on push
- **Drift detection** - Finds stopped or wrong-image containers
- **Auto-reconciliation** - Restarts drifted services automatically
- **Deployment history** - SQLite-backed history with rollback support
- **Multiple modes** - Auto, notify, or manual reconciliation

## Quick start

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest

kedge serve --repo https://github.com/you/your-compose-repo --project myapp
```

## Commands

| Command | Description |
|---------|-------------|
| `serve` | Start the GitOps controller |
| `status` | Show current deployment status |
| `diff` | Show drift between desired and actual state |
| `sync` | Trigger immediate reconciliation |
| `rollback` | Rollback to a previous deployment |
| `history` | Show deployment history |
| `version` | Print version information |
