package store

import (
	"testing"

	"mana/internal/model"
)

func TestCommentsLifecycle(t *testing.T) {
	s, ctx := testStore(t)
	p, err := s.CreateProject(ctx, model.ProjectInput{Name: "评论项目"})
	if err != nil {
		t.Fatal(err)
	}

	first, err := s.CreateComment(ctx, model.OwnerProject, p.ID, "确定用 SQLite 存储")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateComment(ctx, model.OwnerProject, p.ID, "v0.9.3 聚焦待办"); err != nil {
		t.Fatal(err)
	}

	comments, err := s.ListComments(ctx, model.OwnerProject, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 2 {
		t.Fatalf("comments = %#v", comments)
	}
	if comments[0].Content != "v0.9.3 聚焦待办" || comments[0].CreatedAt.Before(comments[1].CreatedAt) {
		t.Fatalf("newest-first ordering broken: %#v", comments)
	}
	_ = first

	if err := s.DeleteComment(ctx, comments[0].ID); err != nil {
		t.Fatal(err)
	}
	if comments, _ = s.ListComments(ctx, model.OwnerProject, p.ID); len(comments) != 1 {
		t.Fatalf("after delete = %#v", comments)
	}
}

// Comments ride along with their host's delete.
func TestCommentsCascadeWithHost(t *testing.T) {
	s, ctx := testStore(t)
	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreateProject(ctx, model.ProjectInput{Name: "Mana Desktop", GroupID: &g.ID})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID, Title: "需求 v1.0", Kind: "requirement",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.CreateComment(ctx, model.OwnerGroup, g.ID, "组级评论"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateComment(ctx, model.OwnerProject, p.ID, "项目评论"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateComment(ctx, "document", d.ID, "文档评论"); err != nil {
		t.Fatal(err)
	}

	// Deleting the document removes only its own comment.
	if err := s.DeleteDocument(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	all, err := s.ListAllComments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("after document delete = %#v", all)
	}

	// Deleting the project removes its comment; the group's stays.
	if err := s.DeleteProject(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if all, _ = s.ListAllComments(ctx); len(all) != 1 || all[0].Content != "组级评论" {
		t.Fatalf("after project delete = %#v", all)
	}

	// Deleting the group clears the last one.
	if err := s.DeleteGroup(ctx, g.ID); err != nil {
		t.Fatal(err)
	}
	if all, _ = s.ListAllComments(ctx); len(all) != 0 {
		t.Fatalf("after group delete = %#v", all)
	}
}
