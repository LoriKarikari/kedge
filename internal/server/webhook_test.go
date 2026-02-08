package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

const (
	headerGitHubEvent   = "X-GitHub-Event"
	headerHubSignature  = "X-Hub-Signature-256"
	headerGitLabEvent   = "X-Gitlab-Event"
	headerGitLabToken   = "X-Gitlab-Token"
	headerGiteaEvent    = "X-Gitea-Event"
	headerGiteaSig      = "X-Gitea-Signature"
	headerWebhookSecret = "X-Webhook-Secret"
	testCommitFmt       = "commit: got %q, want %q"
)

func computeHMAC(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    webhookProvider
	}{
		{
			name:    "github",
			headers: map[string]string{headerGitHubEvent: "push"},
			want:    providerGitHub,
		},
		{
			name:    "github with signature only",
			headers: map[string]string{headerGitHubEvent: "push", headerHubSignature: "sha256=abc"},
			want:    providerGitHub,
		},
		{
			name:    "gitlab",
			headers: map[string]string{headerGitLabEvent: "Push Hook"},
			want:    providerGitLab,
		},
		{
			name:    "gitea",
			headers: map[string]string{headerGiteaEvent: "push"},
			want:    providerGitea,
		},
		{
			name:    "generic fallback",
			headers: map[string]string{},
			want:    providerGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			got := detectProvider(h)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGitHub(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"after": "abc123",
		"repository": {
			"clone_url": "https://github.com/org/repo.git",
			"ssh_url": "git@github.com:org/repo.git",
			"html_url": "https://github.com/org/repo"
		}
	}`)

	payload, err := parseWebhook(providerGitHub, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Branch != "main" {
		t.Errorf("branch: got %q, want %q", payload.Branch, "main")
	}
	if payload.Commit != "abc123" {
		t.Errorf(testCommitFmt, payload.Commit, "abc123")
	}
	if payload.RepoURL != "https://github.com/org/repo.git" {
		t.Errorf("url: got %q, want %q", payload.RepoURL, "https://github.com/org/repo.git")
	}
	if payload.Provider != providerGitHub {
		t.Errorf("provider: got %q, want %q", payload.Provider, providerGitHub)
	}
}

func TestParseGitLab(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/develop",
		"after": "def456",
		"repository": {
			"url": "https://gitlab.com/org/repo.git",
			"git_http_url": "https://gitlab.com/org/repo.git",
			"git_ssh_url": "git@gitlab.com:org/repo.git"
		}
	}`)

	payload, err := parseWebhook(providerGitLab, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Branch != "develop" {
		t.Errorf("branch: got %q, want %q", payload.Branch, "develop")
	}
	if payload.Commit != "def456" {
		t.Errorf(testCommitFmt, payload.Commit, "def456")
	}
}

func TestParseGitea(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"after": "ghi789",
		"repository": {
			"clone_url": "https://gitea.example.com/org/repo.git",
			"ssh_url": "git@gitea.example.com:org/repo.git",
			"html_url": "https://gitea.example.com/org/repo"
		}
	}`)

	payload, err := parseWebhook(providerGitea, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Commit != "ghi789" {
		t.Errorf(testCommitFmt, payload.Commit, "ghi789")
	}
}

func TestParseGeneric(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"after": "jkl012",
		"repository": {"url": "https://example.com/repo.git"}
	}`)

	payload, err := parseWebhook(providerGeneric, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Commit != "jkl012" {
		t.Errorf(testCommitFmt, payload.Commit, "jkl012")
	}
}

func TestParseTagRefRejected(t *testing.T) {
	body := []byte(`{
		"ref": "refs/tags/v1.0.0",
		"after": "abc123",
		"repository": {"clone_url": "https://github.com/org/repo.git"}
	}`)

	for _, provider := range []webhookProvider{providerGitHub, providerGitLab, providerGitea, providerGeneric} {
		t.Run(string(provider), func(t *testing.T) {
			_, err := parseWebhook(provider, body)
			if err == nil {
				t.Error("expected error for tag ref")
			}
		})
	}
}

func TestValidateGitHubSignature(t *testing.T) {
	secret := "mysecret"
	body := []byte(`{"ref":"refs/heads/main"}`)
	sig := "sha256=" + computeHMAC(secret, string(body))

	headers := http.Header{}
	headers.Set(headerHubSignature, sig)

	if err := validateSignature(providerGitHub, secret, headers, body); err != nil {
		t.Errorf("valid signature rejected: %v", err)
	}

	headers.Set(headerHubSignature, "sha256=invalid")
	if err := validateSignature(providerGitHub, secret, headers, body); err == nil {
		t.Error("invalid signature accepted")
	}
}

func TestValidateGitLabToken(t *testing.T) {
	secret := "mytoken"
	headers := http.Header{}
	headers.Set(headerGitLabToken, secret)

	if err := validateSignature(providerGitLab, secret, headers, nil); err != nil {
		t.Errorf("valid token rejected: %v", err)
	}

	headers.Set(headerGitLabToken, "wrong")
	if err := validateSignature(providerGitLab, secret, headers, nil); err == nil {
		t.Error("invalid token accepted")
	}
}

func TestValidateGiteaSignature(t *testing.T) {
	secret := "mysecret"
	body := []byte(`{"ref":"refs/heads/main"}`)
	sig := computeHMAC(secret, string(body))

	headers := http.Header{}
	headers.Set(headerGiteaSig, sig)

	if err := validateSignature(providerGitea, secret, headers, body); err != nil {
		t.Errorf("valid signature rejected: %v", err)
	}

	headers.Set(headerGiteaSig, "invalid")
	if err := validateSignature(providerGitea, secret, headers, body); err == nil {
		t.Error("invalid signature accepted")
	}
}

func TestValidateGenericSecret(t *testing.T) {
	secret := "mysecret"
	headers := http.Header{}
	headers.Set(headerWebhookSecret, secret)

	if err := validateSignature(providerGeneric, secret, headers, nil); err != nil {
		t.Errorf("valid secret rejected: %v", err)
	}

	headers.Set(headerWebhookSecret, "wrong")
	if err := validateSignature(providerGeneric, secret, headers, nil); err == nil {
		t.Error("invalid secret accepted")
	}
}

func TestValidateEmptySecretSkips(t *testing.T) {
	headers := http.Header{}
	for _, provider := range []webhookProvider{providerGitHub, providerGitLab, providerGitea, providerGeneric} {
		t.Run(string(provider), func(t *testing.T) {
			if err := validateSignature(provider, "", headers, nil); err != nil {
				t.Errorf("empty secret should skip validation: %v", err)
			}
		})
	}
}
