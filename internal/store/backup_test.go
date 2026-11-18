package store

import (
	"path/filepath"
	"testing"

	"mana/internal/model"
)

// Backup must produce a file another SQLite connection can open with the
// same records inside.
func TestBackupCreatesQueryableCopy(t *testing.T) {
	s, ctx := testStore(t)
	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateProject(ctx, model.ProjectInput{
		Name: "Mana Desktop", GroupID: &g.ID, Tags: []string{"Go"},
		Path: `E:\codes\go_projects\mana_project`, Version: "v0.9.2",
	}); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "mana-backup.db")
	if err := s.Backup(ctx, dest); err != nil {
		t.Fatalf("backup: %v", err)
	}

	copy, err := Open(ctx, dest)
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer copy.Close()
	projects, err := copy.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Name != "Mana Desktop" {
		t.Fatalf("backup projects = %#v", projects)
	}
	groups, err := copy.ListGroups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Name != "客户端应用" {
		t.Fatalf("backup groups = %#v", groups)
	}
}

// Backing up over an existing file must refuse instead of silently
// overwriting it.
func TestBackupRefusesExistingDestination(t *testing.T) {
	s, ctx := testStore(t)
	dest := filepath.Join(t.TempDir(), "mana-backup.db")
	if err := s.Backup(ctx, dest); err != nil {
		t.Fatalf("first backup: %v", err)
	}
	if err := s.Backup(ctx, dest); err == nil {
		t.Fatal("second backup over the same file should fail")
	}
}

func TestBackupRejectsEmptyDestination(t *testing.T) {
	s, ctx := testStore(t)
	if err := s.Backup(ctx, ""); err == nil {
		t.Fatal("empty destination should fail")
	}
}
