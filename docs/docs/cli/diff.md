# kedge diff

## Usage

```
kedge diff --repo <name>
```

## Description

Compares the desired state (from `docker-compose.yaml`) against the actual state (running containers) and displays any differences.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--repo` | Repository name (required) | |

## Examples

```bash
kedge diff --repo webapp
```

## Output (with drift)

```
Repository: webapp

Service: web
  Status: Image mismatch
  Expected: nginx:1.25
  Actual: nginx:1.24

Service: worker
  Status: Not running
  Expected: running
  Actual: exited
```

## Output (in sync)

```
Repository: webapp
Status: In sync
```

## Related Commands

- [kedge sync](sync.md)
- [kedge status](status.md)
