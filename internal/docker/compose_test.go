package docker

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testProject    = "test-project"
	testImageNginx = "nginx:latest"
)

func TestLoadProject(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

	content := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: postgres:18
    environment:
      POSTGRES_PASSWORD: secret
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(t.Context(), composePath, testProject)
	if err != nil {
		t.Fatal(err)
	}

	if project.Name != testProject {
		t.Errorf("got project name %q, want %q", project.Name, testProject)
	}

	if len(project.Services) != 2 {
		t.Errorf("got %d services, want 2", len(project.Services))
	}

	web := project.Services["web"]
	if web.Image != testImageNginx {
		t.Errorf("got web image %q, want %q", web.Image, testImageNginx)
	}

	db := project.Services["db"]
	if db.Image != "postgres:18" {
		t.Errorf("got db image %q, want %q", db.Image, "postgres:18")
	}
}

func TestLoadProjectInvalidFile(t *testing.T) {
	_, err := LoadProject(t.Context(), "/nonexistent/compose.yaml", "test")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestServiceImages(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

	content := `
services:
  web:
    image: nginx:latest
  api:
    image: myapp:v1
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(t.Context(), composePath, "test")
	if err != nil {
		t.Fatal(err)
	}

	images := ServiceImages(project)

	if images["web"] != testImageNginx {
		t.Errorf("got web image %q, want %q", images["web"], testImageNginx)
	}
	if images["api"] != "myapp:v1" {
		t.Errorf("got api image %q, want %q", images["api"], "myapp:v1")
	}
}

func TestServiceNames(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

	content := `
services:
  web:
    image: nginx:latest
  api:
    image: myapp:v1
  db:
    image: postgres:18
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(t.Context(), composePath, "test")
	if err != nil {
		t.Fatal(err)
	}

	names := ServiceNames(project)

	if len(names) != 3 {
		t.Errorf("got %d names, want 3", len(names))
	}
}
