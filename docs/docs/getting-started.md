# Getting Started

This guide walks you through installing Kedge and deploying your first application.

---

## Installation

### From Source

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest
```

### Build from Source

```bash
git clone https://github.com/LoriKarikari/kedge.git
cd kedge && make build
```

### Docker

```bash
docker pull ghcr.io/lorikarikari/kedge:latest
```

---

## Prepare Your Repository

Kedge watches Git repositories that contain a Docker Compose file and a Kedge configuration file.

### 1. Create `kedge.yaml`

Add this file to the root of your repository:

```yaml
docker:
  project_name: myapp
  compose_file: docker-compose.yaml
```

This tells Kedge:

- **project_name**: The Docker Compose project name (used for container naming)
- **compose_file**: Path to your compose file (relative to repo root)

### 2. Ensure you have a `docker-compose.yaml`

```yaml
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
  api:
    image: myapp/api:latest
    environment:
      - DATABASE_URL=postgres://db:5432/app
```

### 3. Push to Git

```bash
git add kedge.yaml
git commit -m "Add kedge configuration"
git push
```

---

## Register Your Repository

Tell Kedge about your repository:

```bash
kedge repo add https://github.com/you/your-app
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Custom name for the repository | Derived from URL |
| `--branch` | Branch to watch | `main` |

### Examples

```bash
# Use defaults
kedge repo add https://github.com/acme/webapp

# Custom name and branch
kedge repo add https://github.com/acme/webapp --name production --branch release

# Private repository (uses git credentials)
kedge repo add https://github.com/acme/private-app
```

### Verify Registration

```bash
kedge repo list
```

---

## Start Kedge

```bash
kedge serve
```

Kedge will:

1. Clone all registered repositories
2. Deploy the Docker Compose stacks
3. Watch for Git changes
4. Monitor for drift and auto-heal

### What Happens Next

- **Git push** → Kedge pulls changes and redeploys
- **Container stops** → Kedge restarts it (in `auto` mode)
- **Image changes** → Kedge recreates the container

---

## Verify Deployment

### Check Status

```bash
kedge status --repo myapp
```

### View Running Containers

```bash
docker ps --filter "label=io.kedge.managed=true"
```

### Check Logs

Kedge logs all operations:

```
INFO starting repo repo=myapp url=https://github.com/you/your-app
INFO git change detected commit=abc123 message="Update nginx config"
INFO reconciliation complete changes=1
```

---

## Next Steps

- [Core Concepts](concepts.md) - Understand repositories, reconciliation, and drift
- [Configuration](configuration.md) - All configuration options
- [CLI Reference](cli/index.md) - Complete command documentation
