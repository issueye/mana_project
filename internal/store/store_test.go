package store

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"mana/internal/model"
)

func testStore(t *testing.T) (*Store, context.Context) {
	t.Helper()
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, ctx
}

func TestMigrateIdempotent(t *testing.T) {
	s, ctx := testStore(t)
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := s.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != len(entries) {
		t.Fatalf("migrations = %d, want one row per script (%d)", n, len(entries))
	}
}

func TestGroupAndProjectCRUD(t *testing.T) {
	s, ctx := testStore(t)
	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "客户端", Color: "#1677FF"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreateProject(ctx, model.ProjectInput{
		Name: "Mana Desktop", GroupID: &g.ID, Tags: []string{"Go", "MyGo"},
		Path: `E:\codes\go_projects\mana_project`, Version: "v0.9.2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.GroupName != "客户端应用" || len(p.Tags) != 2 {
		t.Fatalf("unexpected project: %#v", p)
	}
	p, err = s.UpdateProject(ctx, p.ID, model.ProjectInput{
		Name: "Mana Desktop", GroupID: &g.ID, Tags: []string{"Go"},
		Path: `E:\codes\go_projects\mana_project`, Version: "v1.0.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != "v1.0.0" || len(p.Tags) != 1 {
		t.Fatalf("update failed: %#v", p)
	}
}

func TestUniqueConstraints(t *testing.T) {
	s, ctx := testStore(t)
	if _, err := s.CreateGroup(ctx, model.GroupInput{Name: "后端服务"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateGroup(ctx, model.GroupInput{Name: "后端服务"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate group err = %v, want conflict", err)
	}
	g, _ := s.GetGroup(ctx, 1)
	if _, err := s.CreateProject(ctx, model.ProjectInput{Name: "Atlas", GroupID: &g.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateProject(ctx, model.ProjectInput{Name: "atlas", GroupID: &g.ID}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate project err = %v, want conflict", err)
	}
	if _, err := s.CreateProject(ctx, model.ProjectInput{Name: "Solo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateProject(ctx, model.ProjectInput{Name: "solo"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate ungrouped err = %v, want conflict", err)
	}
}

func TestDeleteGroupMovesProjectsAndDeletesGroupDocs(t *testing.T) {
	s, ctx := testStore(t)
	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "客户端应用"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreateProject(ctx, model.ProjectInput{Name: "Mana", GroupID: &g.ID})
	if err != nil {
		t.Fatal(err)
	}
	groupDoc, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerGroup, OwnerID: g.ID, Title: "组规范", Kind: "design", Content: "组级文档",
	})
	if err != nil {
		t.Fatal(err)
	}
	projectDoc, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "需求", Kind: "requirement", Content: "项目文档",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteGroup(ctx, g.ID); err != nil {
		t.Fatal(err)
	}
	p, err = s.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.GroupID != nil || p.GroupName != "未分组" {
		t.Fatalf("project not moved: %#v", p)
	}
	if _, err := s.GetDocument(ctx, groupDoc.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("group document still exists: %v", err)
	}
	if _, err := s.GetDocument(ctx, projectDoc.ID); err != nil {
		t.Fatalf("project document lost: %v", err)
	}
}

func TestDeleteProjectCascadesDocuments(t *testing.T) {
	s, ctx := testStore(t)
	p, err := s.CreateProject(ctx, model.ProjectInput{Name: "Temp"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "Doc", Content: "x",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteProject(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetProject(ctx, p.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("project still exists: %v", err)
	}
	docs, err := s.ListDocuments(ctx, model.OwnerProject, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Fatalf("documents remain: %#v", docs)
	}
}

func TestDocumentContentAndCounts(t *testing.T) {
	s, ctx := testStore(t)
	g, _ := s.CreateGroup(ctx, model.GroupInput{Name: "G"})
	p, _ := s.CreateProject(ctx, model.ProjectInput{Name: "P", GroupID: &g.ID})
	d, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "需求", Kind: "requirement",
		Content: "第一行\n第二行", Link: "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Words == 0 {
		t.Fatal("word count is zero")
	}
	full, err := s.GetDocument(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if full.Content != "第一行\n第二行" {
		t.Fatalf("content = %q", full.Content)
	}
	if full.OwnerName != "P" {
		t.Fatalf("owner name = %q", full.OwnerName)
	}
	n, err := s.CountDocuments(ctx, model.OwnerProject, p.ID)
	if err != nil || n != 1 {
		t.Fatalf("count = %d, %v", n, err)
	}
	// Project metadata includes the document count.
	p, _ = s.GetProject(ctx, p.ID)
	if p.DocumentCount != 1 {
		t.Fatalf("document count = %d", p.DocumentCount)
	}
}
