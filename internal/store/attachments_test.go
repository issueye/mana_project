package store

import (
	"os"
	"path/filepath"
	"testing"

	"mana/internal/model"
)

// Migration 002 adds the attachment columns.
func TestAttachmentColumnsExist(t *testing.T) {
	s, ctx := testStore(t)
	rows, err := s.DB().QueryContext(ctx, `PRAGMA table_info(documents)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	found := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt any
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		found[name] = true
	}
	if !found["file_name"] || !found["file_size"] {
		t.Fatalf("documents columns missing file_name/file_size: %v", found)
	}
}

func TestSaveFileCopiesAndCleansUp(t *testing.T) {
	s, ctx := testStore(t)
	dir := t.TempDir()
	s.SetFilesDir(dir)

	g, err := s.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "发布说明.txt")
	body := "第一行\n第二行"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := s.SaveFile(ctx, model.DocumentInput{
		OwnerType: model.OwnerGroup, OwnerID: g.ID,
		Title: "发布说明", Kind: "other", Content: body,
	}, src, "发布说明.txt")
	if err != nil {
		t.Fatal(err)
	}
	if d.FileName != "发布说明.txt" || d.FileSize != int64(len(body)) {
		t.Fatalf("saved = %#v", d)
	}
	stored, err := s.AttachmentFile(d.ID, d.FileName)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(stored)
	if err != nil || string(data) != body {
		t.Fatalf("stored copy = %q, %v", data, err)
	}
	if !filepath.IsAbs(stored) {
		t.Fatalf("stored path %q should be absolute", stored)
	}

	// Removing the attachment clears the whole per-document tree.
	s.RemoveAttachment(d.ID)
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatalf("attachment survived removal: %v", err)
	}
}

func TestSaveFileWithoutFilesDirFails(t *testing.T) {
	s, ctx := testStore(t)
	if _, err := s.SaveFile(ctx, model.DocumentInput{
		OwnerType: model.OwnerGroup, OwnerID: 1, Title: "x", Kind: "other",
	}, "somewhere.txt", "somewhere.txt"); err == nil {
		t.Fatal("upload without a files dir should fail")
	}
}
