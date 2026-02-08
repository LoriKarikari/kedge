package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/LoriKarikari/kedge/internal/webhook"
)

type WebhookInput struct {
	RawBody          []byte `json:"-"`
	XHubSignature256 string `header:"X-Hub-Signature-256"`
	XGitlabToken     string `header:"X-Gitlab-Token"`
	XGiteaSignature  string `header:"X-Gitea-Signature"`
	XWebhookSecret   string `header:"X-Webhook-Secret"`
}

type WebhookOutput struct {
	Body struct {
		Status string `json:"status"`
		Repo   string `json:"repo,omitempty"`
	}
}

func (s *Server) handleWebhook(ctx context.Context, input *WebhookInput) (*WebhookOutput, error) {
	headers := buildHeaders(input)
	provider := webhook.DetectProvider(headers)

	payload, err := webhook.Parse(provider, input.RawBody)
	if err != nil {
		s.logger.Warn("webhook parse failed", slog.Any("error", err))
		return nil, huma.Error400BadRequest("invalid webhook payload")
	}

	repo, err := s.svc.FindRepoByURL(ctx, payload.RepoURL)
	if err != nil {
		s.logger.Debug("no matching repo for webhook", slog.String("url", payload.RepoURL))
		return webhookResponse("accepted", ""), nil
	}

	secret := resolveSecret(repo, s.cfg.Webhook.SecretEnv)
	if err := webhook.ValidateSignature(provider, secret, headers, input.RawBody); err != nil {
		s.logger.Warn("webhook signature validation failed", slog.String("repo", repo.Name), slog.Any("error", err))
		return nil, huma.Error401Unauthorized("invalid signature")
	}

	if repo.Branch != payload.Branch {
		s.logger.Debug("webhook branch mismatch", slog.String("repo", repo.Name), slog.String("expected", repo.Branch), slog.String("got", payload.Branch))
		return webhookResponse("branch ignored", repo.Name), nil
	}

	existing, err := s.svc.Store().GetDeploymentByCommit(ctx, repo.Name, payload.Commit)
	if err == nil && existing != nil && existing.Status == state.StatusSuccess {
		s.logger.Info("webhook skipped, commit already deployed", slog.String("repo", repo.Name), slog.String("commit", payload.Commit[:8]))
		return webhookResponse("already deployed", repo.Name), nil
	}

	s.logger.Info("webhook triggering sync", slog.String("repo", repo.Name), slog.String("commit", payload.Commit[:8]))
	go func() {
		syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := s.svc.TriggerSync(syncCtx, repo.Name); err != nil {
			s.logger.Error("webhook sync failed", slog.String("repo", repo.Name), slog.Any("error", err))
		}
	}()

	return webhookResponse("sync triggered", repo.Name), nil
}

func buildHeaders(input *WebhookInput) http.Header {
	h := http.Header{}
	if input.XHubSignature256 != "" {
		h.Set("X-Hub-Signature-256", input.XHubSignature256)
	}
	if input.XGitlabToken != "" {
		h.Set("X-Gitlab-Token", input.XGitlabToken)
	}
	if input.XGiteaSignature != "" {
		h.Set("X-Gitea-Signature", input.XGiteaSignature)
	}
	if input.XWebhookSecret != "" {
		h.Set("X-Webhook-Secret", input.XWebhookSecret)
	}
	return h
}

func resolveSecret(repo *state.Repo, globalSecretEnv string) string {
	if repo.WebhookSecretEnv != "" {
		if v := os.Getenv(repo.WebhookSecretEnv); v != "" {
			return v
		}
	}
	if globalSecretEnv != "" {
		return os.Getenv(globalSecretEnv)
	}
	return ""
}

func webhookResponse(status, repo string) *WebhookOutput {
	out := &WebhookOutput{}
	out.Body.Status = status
	out.Body.Repo = repo
	return out
}
