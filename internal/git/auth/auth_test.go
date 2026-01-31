package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	defaultUsername = "x-access-token"
)

func TestConfigResolveNone(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{name: "nil config", config: nil},
		{name: "empty type", config: &Config{Type: TypeNone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := tt.config.Resolve(nil)
			if err != nil {
				t.Errorf("Resolve() error = %v", err)
			}
			if auth != nil {
				t.Errorf("Resolve() = %T, want nil", auth)
			}
		})
	}
}

func TestConfigResolveSSH(t *testing.T) {
	keyPath := createTestSSHKey(t)

	cfg := &Config{
		Type:       TypeSSHKey,
		SSHKeyPath: keyPath,
	}

	auth, err := cfg.Resolve(slog.Default())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if _, ok := auth.(*gitssh.PublicKeys); !ok {
		t.Errorf("Resolve() = %T, want *ssh.PublicKeys", auth)
	}
}

func TestConfigResolveSSHMissingKeyPath(t *testing.T) {
	cfg := &Config{
		Type:       TypeSSHKey,
		SSHKeyPath: "",
	}

	_, err := cfg.Resolve(nil)
	if err == nil {
		t.Error("Resolve() error = nil, want error for missing key path")
	}
}

func TestConfigResolveSSHNonexistentKey(t *testing.T) {
	cfg := &Config{
		Type:       TypeSSHKey,
		SSHKeyPath: "/nonexistent/path/to/key",
	}

	_, err := cfg.Resolve(nil)
	if err == nil {
		t.Error("Resolve() error = nil, want error for nonexistent key file")
	}
}

func TestConfigResolveToken(t *testing.T) {
	const envVar = "TEST_KEDGE_TOKEN"
	const testToken = "test-token-value"

	t.Setenv(envVar, testToken)

	cfg := &Config{
		Type:        TypeToken,
		Username:    defaultUsername,
		PasswordEnv: envVar,
	}

	auth, err := cfg.Resolve(nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	basicAuth, ok := auth.(*http.BasicAuth)
	if !ok {
		t.Fatalf("Resolve() = %T, want *http.BasicAuth", auth)
	}

	if basicAuth.Username != defaultUsername {
		t.Errorf("Username = %q, want %q", basicAuth.Username, defaultUsername)
	}

	if basicAuth.Password != testToken {
		t.Errorf("Password = %q, want %q", basicAuth.Password, testToken)
	}
}

func TestConfigResolveTokenDefaultUsername(t *testing.T) {
	const envVar = "TEST_KEDGE_TOKEN_DEFAULT"
	t.Setenv(envVar, "test-token")

	cfg := &Config{
		Type:        TypeToken,
		Username:    "",
		PasswordEnv: envVar,
	}

	auth, err := cfg.Resolve(nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	basicAuth, ok := auth.(*http.BasicAuth)
	if !ok {
		t.Fatalf("Resolve() = %T, want *http.BasicAuth", auth)
	}

	if basicAuth.Username != defaultUsername {
		t.Errorf("Username = %q, want default %q", basicAuth.Username, defaultUsername)
	}
}

func TestConfigResolveTokenMissingEnvVar(t *testing.T) {
	cfg := &Config{
		Type:        TypeToken,
		PasswordEnv: "",
	}

	_, err := cfg.Resolve(nil)
	if err == nil {
		t.Error("Resolve() error = nil, want error for missing env var name")
	}
}

func TestConfigResolveTokenUnsetEnvVar(t *testing.T) {
	cfg := &Config{
		Type:        TypeToken,
		PasswordEnv: "UNSET_ENV_VAR_THAT_DOES_NOT_EXIST",
	}

	_, err := cfg.Resolve(nil)
	if err == nil {
		t.Error("Resolve() error = nil, want error for unset env var")
	}
}

func TestConfigResolveUnknownType(t *testing.T) {
	cfg := &Config{
		Type: "unknown-type",
	}

	_, err := cfg.Resolve(nil)
	if err == nil {
		t.Error("Resolve() error = nil, want error for unknown auth type")
	}
}

func TestConfigIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{name: "nil config", config: nil, expected: true},
		{name: "empty type", config: &Config{Type: TypeNone}, expected: true},
		{name: "ssh type", config: &Config{Type: TypeSSHKey}, expected: false},
		{name: "token type", config: &Config{Type: TypeToken}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func createTestSSHKey(t *testing.T) string {
	t.Helper()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	pemBlock, err := cryptossh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}

	keyBytes := pem.EncodeToMemory(pemBlock)

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test_key")

	if err := os.WriteFile(keyPath, keyBytes, 0o600); err != nil {
		t.Fatalf("failed to write test key: %v", err)
	}

	return keyPath
}
