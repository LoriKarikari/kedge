package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/LoriKarikari/kedge/internal/manager"
	"github.com/LoriKarikari/kedge/internal/server"
	"github.com/LoriKarikari/kedge/internal/state"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GitOps controller",
	Long:  `Start watching all registered repositories and automatically deploy changes.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	statePath := cfg.State.Path

	if err := os.MkdirAll(filepath.Dir(statePath), 0o750); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	store, err := state.New(ctx, statePath)
	if err != nil {
		return fmt.Errorf("open state: %w", err)
	}
	defer store.Close()

	mgr := manager.New(store, logger)
	defer mgr.Close()

	srv := server.New(cfg.Server.Port, mgr, logger)
	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", slog.Any("error", err))
		}
	}()

	logger.Info("server started", slog.Int("port", cfg.Server.Port))
	logger.Info("starting kedge manager")

	return mgr.Start(ctx, manager.Config{
		StatePath: statePath,
	})
}
