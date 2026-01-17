# kedge history

## Usage

```
kedge history --repo <name>
```

## Description

Displays the deployment history including commit hashes, status, and timestamps. Use this to find a commit hash for rollback.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository name (required) | |

## Examples

```bash
kedge history --repo webapp
```

## Output

```
Repository: webapp

COMMIT    STATUS       DEPLOYED AT
abc1234   success      2024-01-15 10:30:00
def5678   success      2024-01-14 15:45:00
ghi9012   failed       2024-01-14 12:00:00
jkl3456   rolled_back  2024-01-13 09:00:00
```

## Status Values

| Status | Description |
|--------|-------------|
| `success` | Deployment completed successfully |
| `failed` | Deployment encountered an error |
| `rolled_back` | Deployment was rolled back from |
| `pending` | Deployment in progress |
| `skipped` | No changes were necessary |

## Related Commands

- [kedge rollback](rollback.md)
- [kedge status](status.md)
