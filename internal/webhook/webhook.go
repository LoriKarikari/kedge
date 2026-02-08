package webhook

import (
	"fmt"
	"net/http"
)

type Provider string

const (
	ProviderGitHub  Provider = "github"
	ProviderGitLab  Provider = "gitlab"
	ProviderGitea   Provider = "gitea"
	ProviderGeneric Provider = "generic"
)

type Payload struct {
	Provider Provider
	RepoURL  string
	Branch   string
	Commit   string
}

func DetectProvider(headers http.Header) Provider {
	switch {
	case headers.Get("X-Hub-Signature-256") != "":
		return ProviderGitHub
	case headers.Get("X-Gitlab-Token") != "":
		return ProviderGitLab
	case headers.Get("X-Gitea-Signature") != "":
		return ProviderGitea
	default:
		return ProviderGeneric
	}
}

func Parse(provider Provider, body []byte) (*Payload, error) {
	var (
		payload *Payload
		err     error
	)

	switch provider {
	case ProviderGitHub:
		payload, err = parseGitHub(body)
	case ProviderGitLab:
		payload, err = parseGitLab(body)
	case ProviderGitea:
		payload, err = parseGitea(body)
	case ProviderGeneric:
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

func ValidateSignature(provider Provider, secret string, headers http.Header, body []byte) error {
	if secret == "" {
		return nil
	}

	switch provider {
	case ProviderGitHub:
		return validateGitHubSignature(secret, headers.Get("X-Hub-Signature-256"), body)
	case ProviderGitLab:
		return validateGitLabToken(secret, headers.Get("X-Gitlab-Token"))
	case ProviderGitea:
		return validateGiteaSignature(secret, headers.Get("X-Gitea-Signature"), body)
	case ProviderGeneric:
		return validateGenericSecret(secret, headers.Get("X-Webhook-Secret"))
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}
