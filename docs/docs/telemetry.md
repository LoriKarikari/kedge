# Telemetry

Kedge exposes Prometheus metrics for monitoring deployments, drift detection, and reconciliation performance.

## Metrics Endpoint

Metrics are exposed at `/metrics` on the HTTP server (default port `8080`):

```bash
curl http://localhost:8080/metrics
```

## Prometheus Configuration

Add Kedge as a scrape target in your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'kedge'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
```

## Metrics Reference

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `kedge_info` | Gauge | `version`, `commit` | Build information |
| `kedge_deployments_total` | Counter | `repo`, `status` | Total deployments |
| `kedge_drift_detected_total` | Counter | `repo`, `service` | Drift detections |
| `kedge_git_polls_total` | Counter | `repo`, `success` | Git poll operations |
| `kedge_reconciliation_duration_seconds` | Histogram | `repo`, `success` | Reconciliation duration |
| `kedge_git_poll_duration_seconds` | Histogram | `repo`, `success` | Git poll duration |
| `kedge_services_total` | UpDownCounter | `repo`, `state` | Current services by state |
| `kedge_last_deployment_timestamp_seconds` | Gauge | `repo` | Last deployment timestamp |

### Label Values

**status** (deployments):

- `success` - Deployment completed successfully
- `failed` - Deployment failed

**success** (git polls, reconciliation):

- `true` - Operation succeeded
- `false` - Operation failed

**state** (services):

- `running` - Service is running
- `stopped` - Service is stopped
- `created` - Service is created but not started

## Example Queries

**Deployment rate per repository:**

```promql
rate(kedge_deployments_total[5m])
```

**Failed deployments:**

```promql
sum(rate(kedge_deployments_total{status="failed"}[5m])) by (repo)
```

**Reconciliation p95 latency:**

```promql
histogram_quantile(0.95, rate(kedge_reconciliation_duration_seconds_bucket[5m]))
```

**Drift detection rate:**

```promql
rate(kedge_drift_detected_total[5m])
```

## Grafana Dashboard

A pre-built Grafana dashboard is available at [`contrib/grafana/kedge-dashboard.json`](https://github.com/LoriKarikari/kedge/blob/main/contrib/grafana/kedge-dashboard.json).

To import:

1. In Grafana, go to **Dashboards** > **Import**
2. Upload the JSON file or paste its contents
3. Select your Prometheus data source
4. Click **Import**

The dashboard includes:

- Kedge version info
- Total deployments counter
- Drift detection counter
- Git poll counter
- Deployment rate over time
- Reconciliation duration (p50/p95)
- Git poll rate and duration
- Last deployment timestamp
- Go runtime goroutines
