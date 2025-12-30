CREATE TABLE IF NOT EXISTS repos (
    name TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE deployments ADD COLUMN repo_name TEXT NOT NULL DEFAULT 'default';

CREATE INDEX IF NOT EXISTS idx_deployments_repo ON deployments(repo_name);
