package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mana/internal/model"
	"mana/internal/store"
)

func testService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st), ctx
}

func TestGroupValidation(t *testing.T) {
	svc, ctx := testService(t)
	if _, err := svc.CreateGroup(ctx, model.GroupInput{Name: "  "}); err == nil {
		t.Fatal("empty group name accepted")
	}
	if _, err := svc.CreateGroup(ctx, model.GroupInput{Name: strings.Repeat("组", 61)}); err == nil {
		t.Fatal("long group name accepted")
	}
	if _, err := svc.CreateGroup(ctx, model.GroupInput{Name: "客户端"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateGroup(ctx, model.GroupInput{Name: "客户端"}); err == nil {
		t.Fatal("duplicate group accepted")
	} else if !strings.Contains(err.Error(), "已被") {
		t.Fatalf("duplicate message = %v", err)
	}
}

func TestProjectValidation(t *testing.T) {
	svc, ctx := testService(t)
	if _, err := svc.CreateProject(ctx, model.ProjectInput{Name: " "}); err == nil {
		t.Fatal("empty project accepted")
	}
	if _, err := svc.CreateProject(ctx, model.ProjectInput{Name: "P", Path: "relative/path"}); err == nil {
		t.Fatal("relative path accepted")
	}
	if _, err := svc.CreateProject(ctx, model.ProjectInput{
		Name: "P", Tags: []string{"a", "b", "c", "d", "e", "f"},
	}); err == nil {
		t.Fatal("six tags accepted")
	}
	group, err := svc.CreateGroup(ctx, model.GroupInput{Name: "G"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateProject(ctx, model.ProjectInput{Name: "P", GroupID: &group.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateProject(ctx, model.ProjectInput{Name: "p", GroupID: &group.ID}); err == nil {
		t.Fatal("duplicate project accepted")
	} else if !strings.Contains(err.Error(), "同名") {
		t.Fatalf("duplicate message = %v", err)
	}
}

func TestTagNormalization(t *testing.T) {
	svc, ctx := testService(t)
	p, err := svc.CreateProject(ctx, model.ProjectInput{
		Name: "P", Tags: []string{" Go ", "go", "MyGo", "MyGo"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "Go" || p.Tags[1] != "MyGo" {
		t.Fatalf("tags = %#v", p.Tags)
	}
}

func TestDocumentValidationAndConflict(t *testing.T) {
	svc, ctx := testService(t)
	p, err := svc.CreateProject(ctx, model.ProjectInput{Name: "P"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveDocument(ctx, 0, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "",
	}); err == nil {
		t.Fatal("empty title accepted")
	}
	if _, err := svc.SaveDocument(ctx, 0, model.DocumentInput{
		OwnerType: "wat", OwnerID: p.ID, Title: "T",
	}); err == nil {
		t.Fatal("invalid owner accepted")
	}
	if _, err := svc.SaveDocument(ctx, 0, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "需求",
		Content: "hello",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveDocument(ctx, 0, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "需求",
	}); err == nil {
		t.Fatal("duplicate document title accepted")
	}
}

func TestDeleteGroupMigration(t *testing.T) {
	svc, ctx := testService(t)
	g, err := svc.CreateGroup(ctx, model.GroupInput{Name: "G"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.CreateProject(ctx, model.ProjectInput{Name: "P", GroupID: &g.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteGroup(ctx, g.ID); err != nil {
		t.Fatal(err)
	}
	p, err = svc.Project(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.GroupID != nil {
		t.Fatalf("project group = %v, want nil", p.GroupID)
	}
}

func TestFormatError(t *testing.T) {
	if got := FormatError(&FieldError{Message: "请输入"}); got != "请输入" {
		t.Fatalf("format = %q", got)
	}
	if got := FormatError(store.ErrNotFound); !strings.Contains(got, "不存在") {
		t.Fatalf("format = %q", got)
	}
	if !errors.Is(store.ErrNotFound, store.ErrNotFound) {
		t.Fatal("sentinel mismatch")
	}
}
