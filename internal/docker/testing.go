package docker

import "testing"

const (
	TestComposeFile    = "docker-compose.yaml"
	SkipIntegrationMsg = "skipping integration test"
)

func NewTestClient(t *testing.T, projectName string) *Client {
	t.Helper()
	client, err := NewClient(projectName, nil)
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	_ = client.Remove(t.Context())
	t.Cleanup(func() {
		_ = client.Remove(t.Context())
		_ = client.Close()
	})
	return client
}
