package docker

import "time"

type ServiceStatus struct {
	Service   string    `json:"service"`
	Container string    `json:"container"`
	Image     string    `json:"image"`
	State     string    `json:"state"`
	Health    string    `json:"health,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	LabelManaged     = "io.kedge.managed"
	LabelProject     = "io.kedge.project"
	LabelService     = "io.kedge.service"
	LabelCommit      = "io.kedge.commit"
	LabelComposeFile = "com.docker.compose.project.config_files"
)
