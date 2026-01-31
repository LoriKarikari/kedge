-- SQLite doesn't support DROP COLUMN in older versions, so we recreate the table
CREATE TABLE repos_backup (
    name TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO repos_backup (name, url, branch, created_at)
SELECT name, url, branch, created_at FROM repos;

DROP TABLE repos;

ALTER TABLE repos_backup RENAME TO repos;
