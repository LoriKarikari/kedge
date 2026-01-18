# kedge serve

## Usage

```
kedge serve
```

## Description

Starts the Kedge controller which:

1. Loads all registered repositories from the database
2. Clones repositories (or pulls latest if already cloned)
3. Deploys Docker Compose stacks
4. Watches for Git changes (polls at configured interval)
5. Monitors for drift and reconciles based on mode
6. Exposes HTTP health endpoints

The controller runs until terminated with `SIGINT` (Ctrl+C) or `SIGTERM`.

## Examples

```bash
# Start the controller
kedge serve

# Run in background
kedge serve &

# Run with Docker
docker run -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/lorikarikari/kedge serve
```

## Output

```
INFO starting kedge server port=8080
INFO starting repo repo=webapp url=https://github.com/acme/webapp
INFO initial reconcile complete repo=webapp
INFO git change detected repo=webapp commit=abc1234
INFO reconciliation complete repo=webapp changes=1
```

## Health Endpoints

While running, Kedge exposes HTTP endpoints:

| Endpoint | Description | Success |
|----------|-------------|---------|
| `GET /health` | Server is running | `200 OK` |
| `GET /ready` | At least one repo is synced | `200 OK` |
| `GET /metrics` | Prometheus metrics | `200 OK` |

For detailed metrics documentation, see [Telemetry](../telemetry.md).

Default port: `8080`. Configure in `~/.config/kedge/config.yaml`:

```yaml
server:
  port: 8080
```

## Graceful Shutdown

```bash
# Send SIGINT or SIGTERM
kill -SIGTERM $(pgrep kedge)
```

```
INFO shutting down...
INFO stopped repo repo=webapp
INFO shutdown complete
```

## Related Commands

- [kedge repo add](repo/add.md)
- [kedge healthcheck](healthcheck.md)
- [kedge status](status.md)
