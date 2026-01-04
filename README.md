<p align="center">
  <h1 align="center">kedge</h1>
  <h2 align="center">GitOps controller for Docker Compose</h2>
</p>

<p align="center">
  <a href="https://lorikarikari.github.io/kedge/"><img src="https://img.shields.io/badge/Documentation-394e79?logo=readthedocs&logoColor=00B9FF" alt="Documentation"></a>
<a href="https://github.com/LoriKarikari/kedge/releases"><img src="https://img.shields.io/github/v/release/LoriKarikari/kedge" alt="Release"></a>
  <a href="https://github.com/LoriKarikari/kedge/blob/main/LICENSE"><img src="https://img.shields.io/github/license/LoriKarikari/kedge" alt="License"></a>
  <a href="https://github.com/LoriKarikari/kedge/actions"><img src="https://github.com/LoriKarikari/kedge/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/LoriKarikari/kedge"><img src="https://goreportcard.com/badge/github.com/LoriKarikari/kedge" alt="Go Report Card"></a>
</p>

Kedge watches Git repositories and automatically deploys Docker Compose applications. It detects drift between desired and actual state, provides rollback capabilities, and supports multiple reconciliation modes.

### Features

- **Multi-Repo Support**: Register and manage multiple repositories
- **Git Integration**: Polls repositories for changes and deploys automatically
- **Drift Detection**: Compares running containers against compose files
- **Reconciliation Modes**: Auto, notify, or manual control
- **Deployment History**: Tracks deployments with commit hashes and status
- **Rollback**: Restore previous deployments by commit reference

## Quick Start

```bash
# Add a repository
kedge repo add https://github.com/user/app.git

# Start watching all repos
kedge serve

# Check status
kedge status --repo app
```

Each repository should contain a `kedge.yaml`:

```yaml
docker:
  project_name: myapp
  compose_file: docker-compose.yaml
```

## Installation

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest
```

Or build from source:

```bash
git clone https://github.com/LoriKarikari/kedge.git
cd kedge && make build
```

## Commands

### Repository Management

```bash
kedge repo add <url> [--name NAME] [--branch BRANCH]
kedge repo list
kedge repo remove <name>
```

### Operations

```bash
kedge serve                      # Start controller (watches all repos)
kedge sync --repo <name>         # Trigger reconciliation
kedge diff --repo <name>         # Show drift
kedge status --repo <name>       # Show deployment status
kedge history --repo <name>      # List deployments
kedge rollback --repo <name> <commit>
```

### Health Check

```bash
kedge healthcheck [--port N]
```

## Configuration

Repository `kedge.yaml`:

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

Environment variable expansion is supported:

```yaml
docker:
  project_name: ${APP_NAME:-myapp}
```

## AI Assistance Disclaimer

AI tools (such as Claude, CodeRabbit, Greptile) were used during development, but all code is reviewed and tested by the maintainers to ensure quality and correctness.
