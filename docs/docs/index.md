# Kedge

**GitOps for Docker Compose.** Push to Git, your containers update automatically.

```bash
kedge repo add https://github.com/you/app && kedge serve
```

---

## How It Works

1. **You push** to your Git repository
2. **Kedge detects** the change and pulls
3. **Docker Compose** stack updates automatically
4. **Drift happens?** Kedge fixes it

---

## Install

```bash
go install github.com/LoriKarikari/kedge/cmd/kedge@latest
```

Or with Docker:

```bash
docker pull ghcr.io/lorikarikari/kedge:latest
```

---

## Quick Start

**1. Add `kedge.yaml` to your repo:**

```yaml
docker:
  project_name: myapp
  compose_file: docker-compose.yaml
```

**2. Register and run:**

```bash
kedge repo add https://github.com/you/your-app
kedge serve
```

Done. Push to Git and watch it deploy.

---

## Next Steps

- [Getting Started](getting-started.md) — Full setup guide
- [Core Concepts](concepts.md) — Understand reconciliation and drift
- [Configuration](configuration.md) — All config options
- [CLI Reference](cli/index.md) — Every command documented
