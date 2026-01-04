CREATE TABLE deployments_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_name TEXT NOT NULL DEFAULT 'default',
    commit_hash TEXT NOT NULL,
    compose_content TEXT NOT NULL,
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    message TEXT
);

INSERT INTO deployments_old SELECT id, repo_name, commit_hash, compose_content, deployed_at, status, message FROM deployments;

DROP TABLE deployments;

ALTER TABLE deployments_old RENAME TO deployments;

CREATE INDEX IF NOT EXISTS idx_deployments_commit ON deployments(commit_hash);
CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at DESC);
CREATE INDEX IF NOT EXISTS idx_deployments_repo ON deployments(repo_name);
