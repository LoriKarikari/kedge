package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/LoriKarikari/kedge/internal/manager"
	"github.com/LoriKarikari/kedge/internal/state"
)

const (
	webhookPath     = "/webhook"
	contentTypeJSON = "application/json"
	wantStatusFmt   = "status: got %d, want %d"
	wantFieldFmt    = "got %q, want %q"
)

type mockService struct {
	repos  map[string]*state.Repo
	synced []string
	store  *state.Store
}

func (m *mockService) IsReady() bool                          { return true }
func (m *mockService) Status() map[string]*manager.RepoStatus { return nil }
func (m *mockService) Store() *state.Store                    { return m.store }
func (m *mockService) TriggerSync(_ context.Context, repoName string) error {
	m.synced = append(m.synced, repoName)
	return nil
}

func (m *mockService) FindRepoByURL(_ context.Context, rawURL string) (*state.Repo, error) {
	for _, r := range m.repos {
		if r.URL == rawURL {
			return r, nil
		}
	}
	return nil, state.ErrNotFound
}

func newMockService(t *testing.T, repos map[string]*state.Repo) *mockService {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := state.New(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return &mockService{repos: repos, store: store}
}

func newTestWebhookServer(t *testing.T, svc Service) (*httptest.Server, *http.Client) {
	t.Helper()
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test", "1.0.0"))

	s := &Server{
		svc:    svc,
		cfg:    ServerConfig{},
		logger: slog.Default(),
	}

	huma.Register(api, huma.Operation{
		OperationID: "webhook",
		Method:      http.MethodPost,
		Path:        webhookPath,
	}, s.handleWebhook)

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	client := ts.Client()
	return ts, client
}

func postWebhook(t *testing.T, client *http.Client, url, body string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url+webhookPath, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response) (string, string) {
	t.Helper()
	defer resp.Body.Close()
	var result struct {
		Status string `json:"status"`
		Repo   string `json:"repo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result.Status, result.Repo
}

func TestWebhookEndpointNoMatchingRepo(t *testing.T) {
	svc := newMockService(t, map[string]*state.Repo{})
	ts, client := newTestWebhookServer(t, svc)

	resp := postWebhook(t, client, ts.URL,
		`{"ref":"refs/heads/main","after":"abc123def456","repository":{"clone_url":"https://github.com/test/nonexistent.git"}}`,
		nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf(wantStatusFmt, resp.StatusCode, http.StatusOK)
	}
	status, _ := decodeResponse(t, resp)
	if status != "accepted" {
		t.Errorf(wantFieldFmt, status, "accepted")
	}
}

func TestWebhookEndpointTagRefRejected(t *testing.T) {
	svc := newMockService(t, map[string]*state.Repo{})
	ts, client := newTestWebhookServer(t, svc)

	resp := postWebhook(t, client, ts.URL,
		`{"ref":"refs/tags/v1.0.0","after":"abc123","repository":{"clone_url":"https://github.com/test/repo.git"}}`,
		map[string]string{headerGitHubEvent: "push"})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(wantStatusFmt, resp.StatusCode, http.StatusBadRequest)
	}
	resp.Body.Close()
}

func TestWebhookEndpointBranchMismatch(t *testing.T) {
	svc := newMockService(t, map[string]*state.Repo{
		"myrepo": {Name: "myrepo", URL: "https://github.com/org/repo.git", Branch: "main"},
	})
	ts, client := newTestWebhookServer(t, svc)

	resp := postWebhook(t, client, ts.URL,
		`{"ref":"refs/heads/develop","after":"abc123def456","repository":{"clone_url":"https://github.com/org/repo.git"}}`,
		map[string]string{headerGitHubEvent: "push"})

	if resp.StatusCode != http.StatusOK {
		t.Errorf(wantStatusFmt, resp.StatusCode, http.StatusOK)
	}
	status, repo := decodeResponse(t, resp)
	if status != "branch ignored" {
		t.Errorf(wantFieldFmt, status, "branch ignored")
	}
	if repo != "myrepo" {
		t.Errorf(wantFieldFmt, repo, "myrepo")
	}
}

func TestWebhookEndpointSyncTriggered(t *testing.T) {
	svc := newMockService(t, map[string]*state.Repo{
		"myrepo": {Name: "myrepo", URL: "https://github.com/org/repo.git", Branch: "main"},
	})
	ts, client := newTestWebhookServer(t, svc)

	resp := postWebhook(t, client, ts.URL,
		`{"ref":"refs/heads/main","after":"abc123def456","repository":{"clone_url":"https://github.com/org/repo.git"}}`,
		map[string]string{headerGitHubEvent: "push"})

	if resp.StatusCode != http.StatusOK {
		t.Errorf(wantStatusFmt, resp.StatusCode, http.StatusOK)
	}
	status, repo := decodeResponse(t, resp)
	if status != "sync triggered" {
		t.Errorf(wantFieldFmt, status, "sync triggered")
	}
	if repo != "myrepo" {
		t.Errorf(wantFieldFmt, repo, "myrepo")
	}
}

func TestWebhookEndpointInvalidJSON(t *testing.T) {
	svc := newMockService(t, map[string]*state.Repo{})
	ts, client := newTestWebhookServer(t, svc)

	resp := postWebhook(t, client, ts.URL, `not json`, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(wantStatusFmt, resp.StatusCode, http.StatusBadRequest)
	}
	resp.Body.Close()
}
