# kedge sync

## Usage

```
kedge sync --repo <name>
```

## Description

Forces an immediate reconciliation for a repository:

1. Pulls the latest changes from Git
2. Compares desired state vs actual state
3. Applies any necessary changes

Useful when using `manual` reconciliation mode or to force a redeploy.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository name (required) | |

## Examples

```bash
# Sync a repository
kedge sync --repo webapp

# Check diff first, then sync
kedge diff --repo webapp
kedge sync --repo webapp
```

## Output

```
Syncing webapp...
Pulled latest changes (commit: def5678)
Reconciliation complete: 2 services updated
```

## Related Commands

- [kedge diff](diff.md)
- [kedge status](status.md)
