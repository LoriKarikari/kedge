package telemetry

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	meterName = "github.com/LoriKarikari/kedge"
)

type Provider struct {
	meterProvider *sdkmetric.MeterProvider
	registry      *prometheus.Registry
	Metrics       *Metrics
}

type Option func(*options)

type options struct {
	registry *prometheus.Registry
}

func WithRegistry(r *prometheus.Registry) Option {
	return func(o *options) {
		o.registry = r
	}
}

func New(opts ...Option) (*Provider, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	registry := cfg.registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)

	meter := provider.Meter(meterName)
	metrics, err := newMetrics(meter)
	if err != nil {
		return nil, err
	}

	return &Provider{
		meterProvider: provider,
		registry:      registry,
		Metrics:       metrics,
	}, nil
}

func (p *Provider) SetGlobal() {
	otel.SetMeterProvider(p.meterProvider)
}

func (p *Provider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.meterProvider.Shutdown(ctx)
}
