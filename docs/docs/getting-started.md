# Getting Started

## Installation

### From source

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest
```

### Docker

```bash
docker pull ghcr.io/lorikarikari/kedge:latest
```

## Setup

1. Create a Git repository with your `docker-compose.yaml`
2. Run kedge pointing to your repo

### Binary

```bash
kedge serve \
  --repo https://github.com/you/your-compose-repo \
  --project myapp \
  --branch main
```

### Docker Compose

```bash
KEDGE_REPO=https://github.com/you/your-compose-repo \
KEDGE_PROJECT=myapp \
docker compose up -d
```

## Reconciliation modes

| Mode | Behavior |
|------|----------|
| `auto` | Automatically fix drift (default) |
| `notify` | Log drift but don't fix |
| `manual` | Wait for `kedge sync` command |

```bash
kedge serve --repo ... --mode notify
```

## Checking status

```bash
kedge status --project myapp
kedge diff --project myapp --compose ./docker-compose.yaml
```

## Manual sync

```bash
kedge sync --project myapp --compose ./docker-compose.yaml
```

## Rollback

```bash
kedge rollback --previous --project myapp --state .kedge/state.db
kedge rollback --to abc123 --project myapp --state .kedge/state.db
```
