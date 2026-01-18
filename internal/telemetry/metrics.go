package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/LoriKarikari/kedge/internal/version"
)

type Metrics struct {
	deploymentsTotal   metric.Int64Counter
	driftDetectedTotal metric.Int64Counter
	gitPollsTotal      metric.Int64Counter

	reconciliationDuration metric.Float64Histogram
	gitPollDuration        metric.Float64Histogram

	servicesTotal           metric.Int64UpDownCounter
	lastDeploymentTimestamp metric.Int64Gauge
	info                    metric.Int64Gauge
}

func newMetrics(meter metric.Meter) (*Metrics, error) {
	m := &Metrics{}
	var err error

	m.deploymentsTotal, err = meter.Int64Counter(
		"kedge_deployments_total",
		metric.WithDescription("Total number of deployments"),
		metric.WithUnit("{deployment}"),
	)
	if err != nil {
		return nil, err
	}

	m.driftDetectedTotal, err = meter.Int64Counter(
		"kedge_drift_detected_total",
		metric.WithDescription("Total number of drift detections"),
		metric.WithUnit("{drift}"),
	)
	if err != nil {
		return nil, err
	}

	m.gitPollsTotal, err = meter.Int64Counter(
		"kedge_git_polls_total",
		metric.WithDescription("Total number of git poll operations"),
		metric.WithUnit("{poll}"),
	)
	if err != nil {
		return nil, err
	}

	m.reconciliationDuration, err = meter.Float64Histogram(
		"kedge_reconciliation_duration_seconds",
		metric.WithDescription("Duration of reconciliation operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.gitPollDuration, err = meter.Float64Histogram(
		"kedge_git_poll_duration_seconds",
		metric.WithDescription("Duration of git poll operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.servicesTotal, err = meter.Int64UpDownCounter(
		"kedge_services_total",
		metric.WithDescription("Current number of services by state"),
		metric.WithUnit("{service}"),
	)
	if err != nil {
		return nil, err
	}

	m.lastDeploymentTimestamp, err = meter.Int64Gauge(
		"kedge_last_deployment_timestamp",
		metric.WithDescription("Unix timestamp of last deployment"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.info, err = meter.Int64Gauge(
		"kedge_info",
		metric.WithDescription("Kedge build information"),
	)
	if err != nil {
		return nil, err
	}

	m.info.Record(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("version", version.Version()),
			attribute.String("commit", version.Commit()),
		),
	)

	return m, nil
}

func (m *Metrics) RecordDeployment(ctx context.Context, repo, status string) {
	m.deploymentsTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.String("status", status),
		),
	)
	m.lastDeploymentTimestamp.Record(ctx, time.Now().Unix(),
		metric.WithAttributes(attribute.String("repo", repo)),
	)
}

func (m *Metrics) RecordDrift(ctx context.Context, repo, service string) {
	m.driftDetectedTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.String("service", service),
		),
	)
}

func (m *Metrics) RecordReconciliation(ctx context.Context, repo string, duration time.Duration, success bool) {
	m.reconciliationDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.Bool("success", success),
		),
	)
}

func (m *Metrics) RecordGitPoll(ctx context.Context, repo string, duration time.Duration, success bool) {
	m.gitPollsTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.Bool("success", success),
		),
	)
	m.gitPollDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.Bool("success", success),
		),
	)
}

func (m *Metrics) SetServices(ctx context.Context, repo, state string, delta int64) {
	m.servicesTotal.Add(ctx, delta,
		metric.WithAttributes(
			attribute.String("repo", repo),
			attribute.String("state", state),
		),
	)
}
