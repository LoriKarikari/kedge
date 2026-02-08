package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

const (
	headerHubSignature  = "X-Hub-Signature-256"
	headerGitLabToken   = "X-Gitlab-Token"
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
		want    Provider
	}{
		{
			name:    "github",
			headers: map[string]string{headerHubSignature: "sha256=abc"},
			want:    ProviderGitHub,
		},
		{
			name:    "gitlab",
			headers: map[string]string{headerGitLabToken: "token"},
			want:    ProviderGitLab,
		},
		{
			name:    "gitea",
			headers: map[string]string{headerGiteaSig: "abc"},
			want:    ProviderGitea,
		},
		{
			name:    "generic fallback",
			headers: map[string]string{},
			want:    ProviderGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			got := DetectProvider(h)
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

	payload, err := Parse(ProviderGitHub, body)
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
	if payload.Provider != ProviderGitHub {
		t.Errorf("provider: got %q, want %q", payload.Provider, ProviderGitHub)
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

	payload, err := Parse(ProviderGitLab, body)
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

	payload, err := Parse(ProviderGitea, body)
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

	payload, err := Parse(ProviderGeneric, body)
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

	for _, provider := range []Provider{ProviderGitHub, ProviderGitLab, ProviderGitea, ProviderGeneric} {
		t.Run(string(provider), func(t *testing.T) {
			_, err := Parse(provider, body)
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

	if err := ValidateSignature(ProviderGitHub, secret, headers, body); err != nil {
		t.Errorf("valid signature rejected: %v", err)
	}

	headers.Set(headerHubSignature, "sha256=invalid")
	if err := ValidateSignature(ProviderGitHub, secret, headers, body); err == nil {
		t.Error("invalid signature accepted")
	}
}

func TestValidateGitLabToken(t *testing.T) {
	secret := "mytoken"
	headers := http.Header{}
	headers.Set(headerGitLabToken, secret)

	if err := ValidateSignature(ProviderGitLab, secret, headers, nil); err != nil {
		t.Errorf("valid token rejected: %v", err)
	}

	headers.Set(headerGitLabToken, "wrong")
	if err := ValidateSignature(ProviderGitLab, secret, headers, nil); err == nil {
		t.Error("invalid token accepted")
	}
}

func TestValidateGiteaSignature(t *testing.T) {
	secret := "mysecret"
	body := []byte(`{"ref":"refs/heads/main"}`)
	sig := computeHMAC(secret, string(body))

	headers := http.Header{}
	headers.Set(headerGiteaSig, sig)

	if err := ValidateSignature(ProviderGitea, secret, headers, body); err != nil {
		t.Errorf("valid signature rejected: %v", err)
	}

	headers.Set(headerGiteaSig, "invalid")
	if err := ValidateSignature(ProviderGitea, secret, headers, body); err == nil {
		t.Error("invalid signature accepted")
	}
}

func TestValidateGenericSecret(t *testing.T) {
	secret := "mysecret"
	headers := http.Header{}
	headers.Set(headerWebhookSecret, secret)

	if err := ValidateSignature(ProviderGeneric, secret, headers, nil); err != nil {
		t.Errorf("valid secret rejected: %v", err)
	}

	headers.Set(headerWebhookSecret, "wrong")
	if err := ValidateSignature(ProviderGeneric, secret, headers, nil); err == nil {
		t.Error("invalid secret accepted")
	}
}

func TestValidateEmptySecretSkips(t *testing.T) {
	headers := http.Header{}
	for _, provider := range []Provider{ProviderGitHub, ProviderGitLab, ProviderGitea, ProviderGeneric} {
		t.Run(string(provider), func(t *testing.T) {
			if err := ValidateSignature(provider, "", headers, nil); err != nil {
				t.Errorf("empty secret should skip validation: %v", err)
			}
		})
	}
}
