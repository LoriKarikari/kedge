package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var healthcheckFlags struct {
	port int
}

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check if kedge server is healthy",
	Long:  `Check if the kedge server is running and ready. Useful for container HEALTHCHECK.`,
	RunE:  runHealthcheck,
}

func init() {
	healthcheckCmd.Flags().IntVar(&healthcheckFlags.port, "port", 8080, "Server port to check")
	rootCmd.AddCommand(healthcheckCmd)
}

func runHealthcheck(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/health", healthcheckFlags.port), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
	}

	fmt.Println("healthy")
	return nil
}
