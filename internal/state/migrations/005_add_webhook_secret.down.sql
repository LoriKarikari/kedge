CREATE TABLE repos_backup (
    name TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    auth_type TEXT DEFAULT NULL,
    auth_ssh_key_path TEXT DEFAULT NULL,
    auth_username TEXT DEFAULT NULL,
    auth_password_env TEXT DEFAULT NULL
);

INSERT INTO repos_backup (name, url, branch, created_at, auth_type, auth_ssh_key_path, auth_username, auth_password_env)
SELECT name, url, branch, created_at, auth_type, auth_ssh_key_path, auth_username, auth_password_env FROM repos;

DROP TABLE repos;

ALTER TABLE repos_backup RENAME TO repos;
