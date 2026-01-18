package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
)

const (
	testRepo  = "test-repo"
	errNewFmt = "New() error = %v"
)

func setupProvider(t *testing.T) *Provider {
	t.Helper()
	registry := prometheus.NewRegistry()
	provider, err := New(WithRegistry(registry))
	if err != nil {
		t.Fatalf(errNewFmt, err)
	}
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return provider
}

func getMetricsBody(t *testing.T, provider *Provider) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	provider.Handler().ServeHTTP(rec, req)
	return rec.Body.String()
}

func getGaugeValue(t *testing.T, provider *Provider, metricName, labelName, labelValue string) (float64, bool) {
	t.Helper()
	metricFamilies, err := provider.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, mf := range metricFamilies {
		if mf.GetName() != metricName {
			continue
		}
		for _, m := range mf.GetMetric() {
			for _, label := range m.GetLabel() {
				if label.GetName() == labelName && label.GetValue() == labelValue {
					return m.GetGauge().GetValue(), true
				}
			}
		}
	}
	return 0, false
}

func TestNew(t *testing.T) {
	provider := setupProvider(t)

	if provider.Metrics == nil {
		t.Error("expected Metrics to be initialized")
	}
	if provider.meterProvider == nil {
		t.Error("expected meterProvider to be initialized")
	}
}

func TestNewWithoutRegistry(t *testing.T) {
	provider, err := New()
	if err != nil {
		t.Fatalf(errNewFmt, err)
	}
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	if provider.registry == nil {
		t.Error("expected default registry to be created when not provided")
	}

	handler := provider.Handler()
	if handler == nil {
		t.Error("expected Handler to return non-nil with default registry")
	}
}

func TestSetGlobal(t *testing.T) {
	previous := otel.GetMeterProvider()
	t.Cleanup(func() { otel.SetMeterProvider(previous) })

	registry := prometheus.NewRegistry()
	provider, err := New(WithRegistry(registry))
	if err != nil {
		t.Fatalf(errNewFmt, err)
	}
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	provider.SetGlobal()
}

func TestHandler(t *testing.T) {
	provider := setupProvider(t)

	handler := provider.Handler()
	if handler == nil {
		t.Fatal("expected Handler to return non-nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "kedge_info") {
		t.Error("expected response to contain kedge_info metric")
	}
}

func TestMetricsRecordDeployment(t *testing.T) {
	provider := setupProvider(t)

	provider.Metrics.RecordDeployment(context.Background(), testRepo, "success")

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_deployments_total") {
		t.Error("expected kedge_deployments_total metric")
	}
	if !strings.Contains(body, `repo="`+testRepo+`"`) {
		t.Error("expected metric to have repo label")
	}
}

func TestMetricsRecordDeploymentUpdatesTimestamp(t *testing.T) {
	provider := setupProvider(t)

	before := time.Now().Unix()
	provider.Metrics.RecordDeployment(context.Background(), testRepo, "success")
	after := time.Now().Unix()

	timestampValue, found := getGaugeValue(t, provider, "kedge_last_deployment_timestamp_seconds", "repo", testRepo)
	if !found {
		t.Fatal("expected kedge_last_deployment_timestamp_seconds metric with repo label")
	}

	if int64(timestampValue) < before || int64(timestampValue) > after {
		t.Errorf("timestamp %v not in expected range [%d, %d]", timestampValue, before, after)
	}
}

func TestMetricsRecordDrift(t *testing.T) {
	provider := setupProvider(t)

	provider.Metrics.RecordDrift(context.Background(), testRepo, "web")

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_drift_detected_total") {
		t.Error("expected kedge_drift_detected_total metric")
	}
	if !strings.Contains(body, `service="web"`) {
		t.Error("expected metric to have service label")
	}
}

func TestMetricsRecordReconciliation(t *testing.T) {
	provider := setupProvider(t)

	provider.Metrics.RecordReconciliation(context.Background(), testRepo, 500*time.Millisecond, true)

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_reconciliation_duration_seconds") {
		t.Error("expected kedge_reconciliation_duration_seconds metric")
	}
}

func TestMetricsRecordGitPoll(t *testing.T) {
	provider := setupProvider(t)

	provider.Metrics.RecordGitPoll(context.Background(), testRepo, 100*time.Millisecond, true)

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_git_polls_total") {
		t.Error("expected kedge_git_polls_total metric")
	}
	if !strings.Contains(body, "kedge_git_poll_duration_seconds") {
		t.Error("expected kedge_git_poll_duration_seconds metric")
	}
}

func TestMetricsSetServices(t *testing.T) {
	provider := setupProvider(t)

	provider.Metrics.SetServices(context.Background(), testRepo, "running", 3)

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_services_total") {
		t.Error("expected kedge_services_total metric")
	}
	if !strings.Contains(body, `state="running"`) {
		t.Error("expected metric to have state label")
	}
	if !strings.Contains(body, `repo="`+testRepo+`"`) {
		t.Error("expected metric to have repo label")
	}
}

func TestMetricsSetServicesUpDown(t *testing.T) {
	provider := setupProvider(t)
	ctx := context.Background()

	provider.Metrics.SetServices(ctx, testRepo, "running", 5)
	provider.Metrics.SetServices(ctx, testRepo, "running", -2)

	body := getMetricsBody(t, provider)
	if !strings.Contains(body, "kedge_services_total") {
		t.Error("expected kedge_services_total metric")
	}
}

func TestIsolatedRegistries(t *testing.T) {
	registry1 := prometheus.NewRegistry()
	provider1, err := New(WithRegistry(registry1))
	if err != nil {
		t.Fatalf(errNewFmt, err)
	}
	t.Cleanup(func() { _ = provider1.Shutdown(context.Background()) })

	registry2 := prometheus.NewRegistry()
	provider2, err := New(WithRegistry(registry2))
	if err != nil {
		t.Fatalf(errNewFmt, err)
	}
	t.Cleanup(func() { _ = provider2.Shutdown(context.Background()) })

	provider1.Metrics.RecordDeployment(context.Background(), "repo1", "success")
	provider2.Metrics.RecordDeployment(context.Background(), "repo2", "success")

	body1 := getMetricsBody(t, provider1)
	body2 := getMetricsBody(t, provider2)

	if !strings.Contains(body1, `repo="repo1"`) {
		t.Error("provider1 should have repo1 metric")
	}
	if strings.Contains(body1, `repo="repo2"`) {
		t.Error("provider1 should NOT have repo2 metric (isolation failed)")
	}

	if !strings.Contains(body2, `repo="repo2"`) {
		t.Error("provider2 should have repo2 metric")
	}
	if strings.Contains(body2, `repo="repo1"`) {
		t.Error("provider2 should NOT have repo1 metric (isolation failed)")
	}
}
