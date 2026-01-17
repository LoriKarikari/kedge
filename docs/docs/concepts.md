# Core Concepts

Understanding how Kedge works.

---

## Repositories

A **repository** in Kedge is a Git repository containing:

- A `docker-compose.yaml` (or similar) defining your services
- A `kedge.yaml` configuration file

Kedge clones repositories locally, watches them for changes, and deploys the Docker Compose stack.

### Repository Lifecycle

```
Register → Clone → Deploy → Watch → Reconcile
```

1. **Register**: `kedge repo add <url>` saves the repository to the database
2. **Clone**: On `kedge serve`, repositories are cloned to `.kedge/repos/<name>/`
3. **Deploy**: The Docker Compose stack is deployed
4. **Watch**: Kedge polls for Git changes (default: every 60 seconds)
5. **Reconcile**: When changes are detected, Kedge updates the deployment

### Managing Repositories

```bash
# Add a repository
kedge repo add https://github.com/acme/app --branch main

# List all repositories
kedge repo list

# Remove a repository
kedge repo remove myapp
```

---

## Reconciliation

**Reconciliation** is the process of making the actual state match the desired state.

- **Desired state**: What's defined in your `docker-compose.yaml`
- **Actual state**: What's currently running in Docker

### Reconciliation Triggers

Kedge reconciles when:

1. **Git changes**: A new commit is detected on the watched branch
2. **Drift detected**: Running containers don't match the compose file
3. **Manual sync**: You run `kedge sync --repo <name>`

### Reconciliation Modes

Configure the mode in `kedge.yaml`:

```yaml
reconciliation:
  mode: auto
```

| Mode | Behavior | Use Case |
|------|----------|----------|
| `auto` | Automatically apply all changes | Production with confidence |
| `notify` | Log changes but don't apply | Review before applying |
| `manual` | Wait for `kedge sync` | Full manual control |

### Example: Auto Mode

```
Git push detected → Pull changes → Compare state → Apply changes → Done
```

### Example: Manual Mode

```
Git push detected → Pull changes → Compare state → Log diff → Wait
                                                            ↓
                              kedge sync --repo myapp → Apply changes → Done
```

---

## Drift Detection

**Drift** is when the actual state diverges from the desired state. This can happen when:

- A container crashes or is manually stopped
- Someone manually changes a container (different image, env vars)
- A container is deleted
- Docker daemon restarts

### Types of Drift

| Drift Type | Description | Example |
|------------|-------------|---------|
| **Missing** | Service defined but container doesn't exist | Container was deleted |
| **Stopped** | Container exists but isn't running | Container crashed |
| **Wrong Image** | Container running different image | Manual `docker pull` + restart |
| **Extra** | Container exists but not in compose | Orphaned from old config |

### Viewing Drift

```bash
kedge diff --repo myapp
```

Example output:

```
Service: web
  Status: Image mismatch
  Expected: nginx:1.25
  Actual: nginx:1.24

Service: worker
  Status: Not running
  Expected: running
  Actual: exited
```

### Drift Resolution

In `auto` mode, Kedge automatically fixes drift:

- **Missing/Stopped**: Recreate and start the container
- **Wrong Image**: Pull correct image and recreate container
- **Extra**: Remove orphaned containers (configurable)

---

## Deployment History

Kedge tracks every deployment in a SQLite database.

### What's Tracked

| Field | Description |
|-------|-------------|
| Commit hash | The Git commit that triggered the deployment |
| Timestamp | When the deployment occurred |
| Status | `success`, `failed`, `rolled_back` |
| Compose content | Snapshot of the compose file |

### Viewing History

```bash
kedge history --repo myapp
```

```
COMMIT    STATUS    DEPLOYED AT
abc1234   success   2024-01-15 10:30:00
def5678   success   2024-01-14 15:45:00
ghi9012   failed    2024-01-14 12:00:00
```

### Rollback

Restore a previous deployment:

```bash
kedge rollback --repo myapp abc1234
```

This:

1. Retrieves the compose file from that deployment
2. Applies it to Docker
3. Records a new deployment with status `rolled_back`

---

## Labels

Kedge adds labels to managed containers for identification:

| Label | Description |
|-------|-------------|
| `io.kedge.managed` | Always `true` for Kedge containers |
| `io.kedge.project` | The project name |
| `io.kedge.service` | The service name |
| `io.kedge.commit` | The Git commit hash |

Query Kedge-managed containers:

```bash
docker ps --filter "label=io.kedge.managed=true"
```
