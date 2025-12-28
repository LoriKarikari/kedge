package state

import (
	"context"
	"database/sql"
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	_ "modernc.org/sqlite" // sqlite driver
)

var ErrNotFound = errors.New("not found")

type DeploymentStatus string

const (
	StatusPending    DeploymentStatus = "pending"
	StatusSuccess    DeploymentStatus = "success"
	StatusFailed     DeploymentStatus = "failed"
	StatusSkipped    DeploymentStatus = "skipped"
	StatusRolledBack DeploymentStatus = "rolled_back"

	DefaultListLimit = 100
)

var statusSchema = z.String().OneOf([]string{
	string(StatusPending),
	string(StatusSuccess),
	string(StatusFailed),
	string(StatusSkipped),
	string(StatusRolledBack),
})

var ErrInvalidStatus = errors.New("invalid deployment status")

func (s DeploymentStatus) IsValid() bool {
	str := string(s)
	return statusSchema.Validate(&str) == nil
}

type Deployment struct {
	ID             int64
	CommitHash     string
	ComposeContent string
	DeployedAt     time.Time
	Status         DeploymentStatus
	Message        string
}

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS deployments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	commit_hash TEXT NOT NULL,
	compose_content TEXT NOT NULL,
	deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	status TEXT NOT NULL,
	message TEXT
);

CREATE INDEX IF NOT EXISTS idx_deployments_commit ON deployments(commit_hash);
CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at DESC);
`

func New(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	pragmas := `
		PRAGMA journal_mode=WAL;
		PRAGMA busy_timeout=5000;
		PRAGMA foreign_keys=ON;
	`
	if _, err := db.ExecContext(ctx, pragmas); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveDeployment(ctx context.Context, commit, composeContent string, status DeploymentStatus, message string) (*Deployment, error) {
	if !status.IsValid() {
		return nil, ErrInvalidStatus
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO deployments (commit_hash, compose_content, status, message) VALUES (?, ?, ?, ?)`,
		commit, composeContent, status, message,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetDeployment(ctx, id)
}

func (s *Store) GetDeployment(ctx context.Context, id int64) (*Deployment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE id = ?`,
		id,
	)
	return scanDeployment(row)
}

func (s *Store) GetLastDeployment(ctx context.Context) (*Deployment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, commit_hash, compose_content, deployed_at, status, message FROM deployments ORDER BY id DESC LIMIT 1`,
	)
	return scanDeployment(row)
}

func (s *Store) GetDeploymentByCommit(ctx context.Context, commit string) (*Deployment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE commit_hash = ? ORDER BY id DESC LIMIT 1`,
		commit,
	)
	return scanDeployment(row)
}

func (s *Store) ListDeployments(ctx context.Context, limit int) ([]*Deployment, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, commit_hash, compose_content, deployed_at, status, message FROM deployments ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		d, err := scanDeploymentRows(rows)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}

	return deployments, rows.Err()
}

func (s *Store) UpdateDeploymentStatus(ctx context.Context, id int64, status DeploymentStatus, message string) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE deployments SET status = ?, message = ? WHERE id = ?`,
		status, message, id,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func scanDeployment(row *sql.Row) (*Deployment, error) {
	var d Deployment
	var message sql.NullString
	err := row.Scan(&d.ID, &d.CommitHash, &d.ComposeContent, &d.DeployedAt, &d.Status, &message)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.Message = message.String
	return &d, nil
}

func scanDeploymentRows(rows *sql.Rows) (*Deployment, error) {
	var d Deployment
	var message sql.NullString
	err := rows.Scan(&d.ID, &d.CommitHash, &d.ComposeContent, &d.DeployedAt, &d.Status, &message)
	if err != nil {
		return nil, err
	}
	d.Message = message.String
	return &d, nil
}
