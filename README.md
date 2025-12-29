# kedge

Kedge is a GitOps controller for Docker Compose. It watches a Git repository and automatically deploys changes, detects drift between desired and actual state, and provides rollback capabilities.

## 1. Features

- **Git Integration**: Polls a Git repository for changes and deploys automatically
- **Drift Detection**: Compares running containers against the compose file and reconciles differences
- **Multiple Reconciliation Modes**: Auto, notify, or manual control over deployments
- **Deployment History**: Tracks all deployments with commit hashes and status
- **Rollback Support**: Restore previous deployments by commit reference
- **Health Endpoints**: Liveness and readiness probes for container health checks
- **YAML Configuration**: Optional configuration file with environment variable expansion

## 2. Installation

### From Source

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest
```

### Build from Repository

```bash
git clone https://github.com/LoriKarikari/kedge.git
cd kedge
make build
```

## 3. Quick Start

```bash
# Start watching a repository
kedge serve --repo https://github.com/user/app.git --project myapp

# Check current status
kedge status

# View deployment history
kedge history

# Force sync
kedge sync --force
```

## 4. Configuration

### 4.1 Configuration File

Create a `kedge.yaml` in your working directory:

```yaml
git:
  url: https://github.com/user/app.git
  branch: main
  poll_interval: 1m
  work_dir: .kedge/repo

docker:
  project_name: myapp
  compose_file: docker-compose.yaml

reconciliation:
  mode: auto
  interval: 1m

state:
  path: .kedge/state.db

logging:
  level: info
  format: text

server:
  port: 8080
```

### 4.2 Environment Variables

Configuration values support environment variable expansion:

```yaml
git:
  url: ${KEDGE_REPO}
  branch: ${KEDGE_BRANCH:-main}
```

## 5. CLI Commands

### 5.1 serve

Start the GitOps controller.

```bash
kedge serve --repo <url> [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--repo` | Git repository URL | required |
| `--branch` | Git branch to watch | main |
| `--project` | Docker compose project name | kedge |
| `--compose` | Path to compose file | docker-compose.yaml |
| `--mode` | Reconciliation mode | auto |
| `--poll` | Git poll interval | 1m |

### 5.2 sync

Trigger immediate reconciliation.

```bash
kedge sync [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--force` | Force sync even if no drift | false |
| `--project` | Docker compose project name | kedge |

### 5.3 diff

Show drift between desired and actual state.

```bash
kedge diff [flags]
```

### 5.4 status

Display current deployment status.

```bash
kedge status [flags]
```

### 5.5 history

List deployment history.

```bash
kedge history [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--limit` | Number of entries to show | 10 |

### 5.6 rollback

Rollback to a previous deployment.

```bash
kedge rollback <commit> [flags]
```

## 6. Reconciliation Modes

| Mode | Behavior |
|------|----------|
| `auto` | Automatically apply changes when drift is detected |
| `notify` | Detect drift but do not apply changes |
| `manual` | Only reconcile when explicitly triggered via `kedge sync` |

## 7. AI Assistance Disclaimer

AI tools (Claude) were used during development.