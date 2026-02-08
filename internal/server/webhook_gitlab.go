package server

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"strings"
)

type gitlabPush struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		URL     string `json:"url"`
		GitHTTP string `json:"git_http_url"`
		GitSSH  string `json:"git_ssh_url"`
	} `json:"repository"`
}

func parseGitLab(body []byte) (*webhookPayload, error) {
	var push gitlabPush
	if err := json.Unmarshal(body, &push); err != nil {
		return nil, fmt.Errorf("parse gitlab payload: %w", err)
	}

	branch := strings.TrimPrefix(push.Ref, "refs/heads/")
	if branch == push.Ref {
		return nil, fmt.Errorf("not a branch push: %s", push.Ref)
	}

	repoURL := push.Repository.GitHTTP
	if repoURL == "" {
		repoURL = push.Repository.URL
	}

	return &webhookPayload{
		RepoURL: repoURL,
		Branch:  branch,
		Commit:  push.After,
	}, nil
}

func validateGitLabToken(secret, token string) error {
	if token == "" {
		return fmt.Errorf("missing X-Gitlab-Token header")
	}
	if subtle.ConstantTimeCompare([]byte(secret), []byte(token)) != 1 {
		return fmt.Errorf("invalid token")
	}
	return nil
}
