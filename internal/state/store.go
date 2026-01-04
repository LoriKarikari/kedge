package state

import (
	"context"
	"database/sql"
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite" // sqlite driver

	"github.com/LoriKarikari/kedge/internal/state/migrations"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidStatus = errors.New("invalid deployment status")
)

const DefaultListLimit = 100

type Store struct {
	db *sql.DB
}

type Repo struct {
	Name      string
	URL       string
	Branch    string
	CreatedAt time.Time
}

type Deployment struct {
	ID             int64
	RepoName       string
	CommitHash     string
	ComposeContent string
	DeployedAt     time.Time
	Status         DeploymentStatus
	Message        string
}

type DeploymentStatus string

const (
	StatusPending    DeploymentStatus = "pending"
	StatusSuccess    DeploymentStatus = "success"
	StatusFailed     DeploymentStatus = "failed"
	StatusSkipped    DeploymentStatus = "skipped"
	StatusRolledBack DeploymentStatus = "rolled_back"
)

var statusSchema = z.String().OneOf([]string{
	string(StatusPending),
	string(StatusSuccess),
	string(StatusFailed),
	string(StatusSkipped),
	string(StatusRolledBack),
})

func (s DeploymentStatus) IsValid() bool {
	str := string(s)
	return statusSchema.Validate(&str) == nil
}

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

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveRepo(ctx context.Context, name, url, branch string) (*Repo, error) {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO repos (name, url, branch) VALUES (?, ?, ?)`,
		name, url, branch,
	)
	if err != nil {
		return nil, err
	}
	return s.GetRepo(ctx, name)
}

func (s *Store) GetRepo(ctx context.Context, name string) (*Repo, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT name, url, branch, created_at FROM repos WHERE name = ?`,
		name,
	)
	var r Repo
	err := row.Scan(&r.Name, &r.URL, &r.Branch, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ListRepos(ctx context.Context) ([]*Repo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT name, url, branch, created_at FROM repos ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*Repo
	for rows.Next() {
		var r Repo
		if err := rows.Scan(&r.Name, &r.URL, &r.Branch, &r.CreatedAt); err != nil {
			return nil, err
		}
		repos = append(repos, &r)
	}
	return repos, rows.Err()
}

func (s *Store) DeleteRepo(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM repos WHERE name = ?`, name)
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

func (s *Store) SaveDeployment(ctx context.Context, repoName, commit, composeContent string, status DeploymentStatus, message string) (*Deployment, error) {
	if !status.IsValid() {
		return nil, ErrInvalidStatus
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO deployments (repo_name, commit_hash, compose_content, status, message) VALUES (?, ?, ?, ?, ?)`,
		repoName, commit, composeContent, status, message,
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
		`SELECT id, repo_name, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE id = ?`,
		id,
	)
	return scanDeployment(row)
}

func (s *Store) GetLastDeployment(ctx context.Context, repoName string) (*Deployment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, repo_name, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE repo_name = ? ORDER BY id DESC LIMIT 1`,
		repoName,
	)
	return scanDeployment(row)
}

func (s *Store) GetDeploymentByCommit(ctx context.Context, repoName, commit string) (*Deployment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, repo_name, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE repo_name = ? AND commit_hash = ? ORDER BY id DESC LIMIT 1`,
		repoName, commit,
	)
	return scanDeployment(row)
}

func (s *Store) ListDeployments(ctx context.Context, repoName string, limit int) ([]*Deployment, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, repo_name, commit_hash, compose_content, deployed_at, status, message FROM deployments WHERE repo_name = ? ORDER BY id DESC LIMIT ?`,
		repoName, limit,
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

func runMigrations(db *sql.DB) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func scanDeployment(row *sql.Row) (*Deployment, error) {
	var d Deployment
	var message sql.NullString
	err := row.Scan(&d.ID, &d.RepoName, &d.CommitHash, &d.ComposeContent, &d.DeployedAt, &d.Status, &message)
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
	err := rows.Scan(&d.ID, &d.RepoName, &d.CommitHash, &d.ComposeContent, &d.DeployedAt, &d.Status, &message)
	if err != nil {
		return nil, err
	}
	d.Message = message.String
	return &d, nil
}
