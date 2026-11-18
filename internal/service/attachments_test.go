package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mana/internal/model"
	"mana/internal/store"
)

func fileService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	st.SetFilesDir(t.TempDir())
	t.Cleanup(func() { st.Close() })
	return New(st), ctx
}

func TestSaveDocumentFileUploadsAndExtracts(t *testing.T) {
	svc, ctx := fileService(t)
	g, err := svc.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(t.TempDir(), "发布说明.txt")
	if err := os.WriteFile(src, []byte("版本 v1.0 已发布"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := svc.SaveDocumentFile(ctx, model.OwnerGroup, g.ID, src)
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "发布说明" || d.FileName != "发布说明.txt" {
		t.Fatalf("uploaded = %#v", d)
	}
	if d.FileSize != int64(len("版本 v1.0 已发布")) {
		t.Fatalf("size = %d", d.FileSize)
	}
	if !strings.Contains(d.Content, "版本 v1.0") {
		t.Fatalf("content = %q", d.Content)
	}

	// The stored copy exists where the service says it does.
	path, err := svc.DocumentFilePath(d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stored file: %v", err)
	}

	// Uploading the same name again deconflicts the title.
	d2, err := svc.SaveDocumentFile(ctx, model.OwnerGroup, g.ID, src)
	if err != nil {
		t.Fatal(err)
	}
	if d2.Title != "发布说明 (2)" {
		t.Fatalf("second title = %q", d2.Title)
	}

	// Deleting removes the stored tree.
	if err := svc.DeleteDocument(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("attachment survived delete: %v", err)
	}
}

func TestSaveDocumentFileRejectsOversized(t *testing.T) {
	svc, ctx := fileService(t)
	g, err := svc.CreateGroup(ctx, model.GroupInput{Name: "客户端应用", Description: "", Color: "#33586B"})
	if err != nil {
		t.Fatal(err)
	}
	// Pretend the file is huge without writing 100 MB: a directory stat
	// short-circuits first, so check the size gate via a sparse file.
	src := filepath.Join(t.TempDir(), "big.txt")
	f, err := os.Create(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(MaxUpload + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := svc.SaveDocumentFile(ctx, model.OwnerGroup, g.ID, src); err == nil {
		t.Fatal("oversized upload should fail")
	}
}
