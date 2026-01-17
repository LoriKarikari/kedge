# CLI Reference

Complete reference for the `kedge` command-line interface.

## Usage

```
kedge <command> [flags]
```

## Commands

### Repository Management

| Command | Description |
|---------|-------------|
| [kedge repo add](repo/add.md) | Register a repository |
| [kedge repo list](repo/list.md) | List registered repositories |
| [kedge repo remove](repo/remove.md) | Remove a repository |

### Controller

| Command | Description |
|---------|-------------|
| [kedge serve](serve.md) | Start the GitOps controller |

### Operations

| Command | Description |
|---------|-------------|
| [kedge status](status.md) | Show deployment status |
| [kedge diff](diff.md) | Show drift between desired and actual state |
| [kedge sync](sync.md) | Trigger immediate reconciliation |

### Deployment History

| Command | Description |
|---------|-------------|
| [kedge history](history.md) | Show deployment history |
| [kedge rollback](rollback.md) | Rollback to a previous deployment |

### Diagnostics

| Command | Description |
|---------|-------------|
| [kedge healthcheck](healthcheck.md) | Check server health |
| [kedge version](version.md) | Print version information |

## Global Flags

| Option | Description |
|--------|-------------|
| `-h`, `--help` | Display help for the command |

## Getting Help

```bash
kedge --help
kedge repo --help
kedge repo add --help
```
