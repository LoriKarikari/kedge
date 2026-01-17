# kedge healthcheck

## Usage

```
kedge healthcheck [flags]
```

## Description

Connects to the Kedge server's health endpoint and reports the status. Returns exit code `0` if healthy, `1` if unhealthy.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--port` | Server port to check | `8080` |

## Examples

```bash
# Check default port
kedge healthcheck

# Check custom port
kedge healthcheck --port 9090

# Use in scripts
if kedge healthcheck; then
  echo "Kedge is running"
else
  echo "Kedge is not running"
fi
```

## Output (healthy)

```
Kedge server is healthy
```

## Output (unhealthy)

```
Error: could not connect to Kedge server at localhost:8080
```

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Server is healthy |
| `1` | Server is unreachable or unhealthy |

## Related Commands

- [kedge serve](serve.md)
- [kedge version](version.md)
