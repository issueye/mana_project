package store

import (
	"testing"

	"mana/internal/model"
)

func TestSearchFindsProjectsDocsGroups(t *testing.T) {
	s, ctx := testStore(t)
	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "后端服务", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.CreateProject(ctx, model.ProjectInput{
		Name: "Atlas API", GroupID: &g.ID, Tags: []string{"网关"},
		Path: `E:\codes\atlas`, Version: "v3.1.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateDocument(ctx, model.DocumentInput{
		OwnerType: model.OwnerProject, OwnerID: p.ID,
		Title: "API 错误码规范", Kind: "api", Content: "# 规范",
	}); err != nil {
		t.Fatal(err)
	}

	// By project name, case-insensitive on ASCII.
	res, err := s.Search(ctx, "atlas")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 || res.Projects[0].Name != "Atlas API" {
		t.Fatalf("projects = %#v", res.Projects)
	}

	// By document title.
	res, err = s.Search(ctx, "错误码")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Docs) != 1 || res.Docs[0].Title != "API 错误码规范" {
		t.Fatalf("docs = %#v", res.Docs)
	}

	// By tag.
	res, err = s.Search(ctx, "网关")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 {
		t.Fatalf("tag search projects = %#v", res.Projects)
	}

	// By path.
	res, err = s.Search(ctx, `E:\codes\atlas`)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 {
		t.Fatalf("path search projects = %#v", res.Projects)
	}

	// By group name.
	res, err = s.Search(ctx, "后端")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Groups) != 1 {
		t.Fatalf("group search groups = %#v", res.Groups)
	}

	// Empty query, no results.
	res, err = s.Search(ctx, "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects)+len(res.Docs)+len(res.Groups) != 0 {
		t.Fatalf("empty query returned %#v", res)
	}
	res, err = s.Search(ctx, "不存在的词")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects)+len(res.Docs)+len(res.Groups) != 0 {
		t.Fatalf("miss query returned %#v", res)
	}
}

func TestSearchLimitsEachSection(t *testing.T) {
	s, ctx := testStore(t)
	for i := 0; i < 8; i++ {
		if _, err := s.CreateProject(ctx, model.ProjectInput{
			Name: "同系列-" + string(rune('一'+i)),
		}); err != nil {
			t.Fatal(err)
		}
	}
	res, err := s.Search(ctx, "同系列")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 5 {
		t.Fatalf("projects = %d, want the cap of 5", len(res.Projects))
	}
}
