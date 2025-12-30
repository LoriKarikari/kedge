DROP INDEX IF EXISTS idx_deployments_repo;

-- SQLite doesn't support DROP COLUMN directly, need to recreate table
CREATE TABLE deployments_backup AS SELECT id, commit_hash, compose_content, deployed_at, status, message FROM deployments;
DROP TABLE deployments;
ALTER TABLE deployments_backup RENAME TO deployments;

CREATE INDEX IF NOT EXISTS idx_deployments_commit ON deployments(commit_hash);
CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at DESC);

DROP TABLE IF EXISTS repos;
