# kedge status

## Usage

```
kedge status --repo <name>
```

## Description

Displays the current deployment status including the active commit, deployment time, and the state of each service.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository name (required) | |

## Examples

```bash
kedge status --repo webapp
```

## Output

```
Repository: webapp
Status: Running
Commit: abc1234
Deployed: 2024-01-15 10:30:00

Services:
  web: running (nginx:1.25)
  api: running (myapp/api:v1.2.3)
  worker: running (myapp/worker:v1.2.3)
```

## Related Commands

- [kedge diff](diff.md)
- [kedge history](history.md)
