package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/LoriKarikari/kedge/internal/telemetry"
)

type ReadinessChecker interface {
	IsReady() bool
}

type Server struct {
	server    *http.Server
	checker   ReadinessChecker
	telemetry *telemetry.Provider
	logger    *slog.Logger
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

type ReadyOutput struct {
	Body struct {
		Status string `json:"status"`
		Ready  bool   `json:"ready"`
	}
}

func New(port int, checker ReadinessChecker, tp *telemetry.Provider, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Kedge API", "1.0.0"))

	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		checker:   checker,
		telemetry: tp,
		logger:    logger,
	}

	huma.Register(api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Liveness check",
	}, s.handleHealth)

	huma.Register(api, huma.Operation{
		OperationID: "ready",
		Method:      http.MethodGet,
		Path:        "/ready",
		Summary:     "Readiness check",
	}, s.handleReady)

	if tp != nil {
		mux.Handle("/metrics", tp.Handler())
	}

	return s
}

func (s *Server) handleHealth(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	return &HealthOutput{
		Body: struct {
			Status string `json:"status"`
		}{Status: "ok"},
	}, nil
}

func (s *Server) handleReady(ctx context.Context, input *struct{}) (*ReadyOutput, error) {
	ready := s.checker != nil && s.checker.IsReady()
	output := &ReadyOutput{}
	output.Body.Ready = ready
	if ready {
		output.Body.Status = "ok"
	} else {
		output.Body.Status = "not ready"
	}
	return output, nil
}

func (s *Server) Start(ctx context.Context) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", s.server.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server error", slog.Any("error", err))
		}
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
