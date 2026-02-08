package server

import (
	"fmt"
	"net/http"
)

type webhookProvider string

const (
	providerGitHub  webhookProvider = "github"
	providerGitLab  webhookProvider = "gitlab"
	providerGitea   webhookProvider = "gitea"
	providerGeneric webhookProvider = "generic"
)

type webhookPayload struct {
	Provider webhookProvider
	RepoURL  string
	Branch   string
	Commit   string
}

func detectProvider(headers http.Header) webhookProvider {
	switch {
	case headers.Get("X-Hub-Signature-256") != "":
		return providerGitHub
	case headers.Get("X-Gitlab-Token") != "":
		return providerGitLab
	case headers.Get("X-Gitea-Signature") != "":
		return providerGitea
	default:
		return providerGeneric
	}
}

func parseWebhook(provider webhookProvider, body []byte) (*webhookPayload, error) {
	var (
		payload *webhookPayload
		err     error
	)

	switch provider {
	case providerGitHub:
		payload, err = parseGitHub(body)
	case providerGitLab:
		payload, err = parseGitLab(body)
	case providerGitea:
		payload, err = parseGitea(body)
	case providerGeneric:
		payload, err = parseGeneric(body)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}
	payload.Provider = provider
	return payload, nil
}

func validateSignature(provider webhookProvider, secret string, headers http.Header, body []byte) error {
	if secret == "" {
		return nil
	}

	switch provider {
	case providerGitHub:
		return validateGitHubSignature(secret, headers.Get("X-Hub-Signature-256"), body)
	case providerGitLab:
		return validateGitLabToken(secret, headers.Get("X-Gitlab-Token"))
	case providerGitea:
		return validateGiteaSignature(secret, headers.Get("X-Gitea-Signature"), body)
	case providerGeneric:
		return validateGenericSecret(secret, headers.Get("X-Webhook-Secret"))
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}
