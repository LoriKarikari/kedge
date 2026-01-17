# kedge rollback

## Usage

```
kedge rollback --repo <name> <commit>
```

## Description

Restores a previous deployment by:

1. Retrieving the compose file snapshot from the specified commit
2. Applying it to Docker
3. Recording a new deployment entry with the rollback

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository name (required) | |

## Arguments

| Argument | Description |
|----------|-------------|
| `commit` | Commit hash to rollback to (from `kedge history`) |

## Examples

```bash
# View history to find commit
kedge history --repo webapp

# Rollback to specific commit
kedge rollback --repo webapp abc1234

# Verify the rollback
kedge status --repo webapp
```

## Output

```
Rolling back webapp to abc1234...
Retrieved compose file from deployment
Applying configuration...
Rollback complete

Current deployment:
  Commit: abc1234
  Status: success
  Services: 3 running
```

## Related Commands

- [kedge history](history.md)
- [kedge status](status.md)
