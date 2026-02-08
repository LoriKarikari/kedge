package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"strings"
)

type genericPush struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		URL string `json:"url"`
	} `json:"repository"`
}

func parseGeneric(body []byte) (*Payload, error) {
	var push genericPush
	if err := json.Unmarshal(body, &push); err != nil {
		return nil, fmt.Errorf("parse generic payload: %w", err)
	}

	branch := strings.TrimPrefix(push.Ref, "refs/heads/")
	if branch == push.Ref {
		return nil, fmt.Errorf("not a branch push: %s", push.Ref)
	}

	return &Payload{
		RepoURL: push.Repository.URL,
		Branch:  branch,
		Commit:  push.After,
	}, nil
}

func validateGenericSecret(secret, headerSecret string) error {
	if headerSecret == "" {
		return fmt.Errorf("missing X-Webhook-Secret header")
	}
	if subtle.ConstantTimeCompare([]byte(secret), []byte(headerSecret)) != 1 {
		return fmt.Errorf("invalid secret")
	}
	return nil
}
