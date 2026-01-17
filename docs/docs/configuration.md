# Configuration

Kedge uses two configuration files: a per-repository `kedge.yaml` and an optional global config.

---

## Repository Configuration

Each repository must have a `kedge.yaml` at the root.

### Minimal Example

```yaml
docker:
  project_name: myapp
  compose_file: docker-compose.yaml
```

### Full Example

```yaml
docker:
  project_name: myapp
  compose_file: docker-compose.yaml

reconciliation:
  mode: auto

logging:
  level: info
  format: text
```

### Reference

#### `docker`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project_name` | string | Yes | Docker Compose project name |
| `compose_file` | string | Yes | Path to compose file (relative to repo root) |

#### `reconciliation`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `auto` | Reconciliation mode: `auto`, `notify`, or `manual` |

#### `logging`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `format` | string | `text` | Log format: `text` or `json` |

---

## Global Configuration

Optional global settings at `~/.config/kedge/config.yaml`.

### Example

```yaml
state:
  path: ~/.local/share/kedge/state.db

server:
  port: 8080

git:
  poll_interval: 60s

logging:
  level: info
  format: text
```

### Reference

#### `state`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `path` | string | `~/.local/share/kedge/state.db` | SQLite database path |

#### `server`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | integer | `8080` | HTTP server port for health endpoints |

#### `git`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `poll_interval` | duration | `60s` | How often to check for Git changes |

#### `logging`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `format` | string | `text` | Log format: `text` or `json` |

---

## Environment Variables

Environment variable expansion is supported in `kedge.yaml`:

```yaml
docker:
  project_name: ${APP_NAME:-myapp}
  compose_file: ${COMPOSE_FILE:-docker-compose.yaml}
```

### Syntax

| Syntax | Description |
|--------|-------------|
| `${VAR}` | Value of `VAR`, empty if unset |
| `${VAR:-default}` | Value of `VAR`, or `default` if unset |

---

## Configuration Precedence

1. Repository `kedge.yaml` (highest priority)
2. Global `~/.config/kedge/config.yaml`
3. Built-in defaults (lowest priority)

Repository settings override global settings where applicable.
