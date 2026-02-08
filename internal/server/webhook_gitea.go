package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type giteaPush struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		SSHURL   string `json:"ssh_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
}

func parseGitea(body []byte) (*webhookPayload, error) {
	var push giteaPush
	if err := json.Unmarshal(body, &push); err != nil {
		return nil, fmt.Errorf("parse gitea payload: %w", err)
	}

	branch := strings.TrimPrefix(push.Ref, "refs/heads/")
	if branch == push.Ref {
		return nil, fmt.Errorf("not a branch push: %s", push.Ref)
	}

	repoURL := push.Repository.CloneURL
	if repoURL == "" {
		repoURL = push.Repository.HTMLURL
	}

	return &webhookPayload{
		RepoURL: repoURL,
		Branch:  branch,
		Commit:  push.After,
	}, nil
}

func validateGiteaSignature(secret, signature string, body []byte) error {
	if signature == "" {
		return fmt.Errorf("missing X-Gitea-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
