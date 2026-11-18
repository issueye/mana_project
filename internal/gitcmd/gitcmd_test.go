package gitcmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// gitAvailable skips the suite when the machine has no git.
func gitAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

// newRepo makes a repository with one commit and returns its path.
func newRepo(t *testing.T) string {
	t.Helper()
	gitAvailable(t)
	dir := t.TempDir()
	ctx := context.Background()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# 样例\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Stage(ctx, dir, nil); err != nil {
		t.Fatal(err)
	}
	// Tests cannot rely on a global git identity; provide one inline.
	cmd := exec.Command("git", "-c", "user.name=Mana 测试", "-c", "user.email=mana@example.com",
		"commit", "-m", "第一次提交")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v: %s", err, out)
	}
	return dir
}

func TestStatusBranchAndFiles(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()

	branch, files, err := Status(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch == "" {
		t.Fatal("branch is empty")
	}
	if len(files) != 0 {
		t.Fatalf("clean tree reported %v", files)
	}

	if err := os.WriteFile(filepath.Join(dir, "新文件.txt"), []byte("内容"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# 改过\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, files, err = Status(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("files = %#v", files)
	}
	byPath := map[string]FileStatus{}
	for _, f := range files {
		byPath[f.Path] = f
	}
	if !byPath["新文件.txt"].Untracked {
		t.Fatalf("untracked detection failed: %#v", byPath)
	}
	if byPath["README.md"].Untracked || byPath["README.md"].Staged {
		t.Fatalf("modification flags wrong: %#v", byPath["README.md"])
	}
}

func TestStageCommitLog(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()

	if err := os.WriteFile(filepath.Join(dir, "功能.txt"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Stage(ctx, dir, []string{"功能.txt"}); err != nil {
		t.Fatal(err)
	}
	_, files, err := Status(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	staged := false
	for _, f := range files {
		if f.Path == "功能.txt" && f.Staged {
			staged = true
		}
	}
	if !staged {
		t.Fatalf("staged flag missing: %#v", files)
	}

	if _, err := Commit(ctx, dir, ""); err == nil {
		t.Fatal("empty message should fail")
	}
	hash, err := Commit(ctx, dir, "新增功能文件")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" || len(hash) > 12 {
		t.Fatalf("hash = %q", hash)
	}

	log, err := Log(ctx, dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(log) != 2 {
		t.Fatalf("log = %#v", log)
	}
	if log[0].Subject != "新增功能文件" || log[0].Author == "" {
		// The author is whatever identity git resolves on this machine.
		t.Fatalf("newest commit = %#v", log[0])
	}
	if log[1].Subject != "第一次提交" {
		t.Fatalf("older commit = %#v", log[1])
	}
	if time.Since(log[0].Time) > time.Hour {
		t.Fatalf("commit time is not recent: %v", log[0].Time)
	}
}

func TestTagsLifecycle(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()

	tags, err := Tags(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("fresh repo has tags: %#v", tags)
	}
	if err := CreateTag(ctx, dir, "", ""); err == nil {
		t.Fatal("empty tag name should fail")
	}
	if err := CreateTag(ctx, dir, "v0.1.0", "首个内部版本"); err != nil {
		t.Fatal(err)
	}
	tags, err = Tags(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Name != "v0.1.0" {
		t.Fatalf("tags = %#v", tags)
	}
	if tags[0].Hash == "" || tags[0].Created.IsZero() {
		t.Fatalf("tag metadata missing: %#v", tags[0])
	}
}

func TestPushWithoutRemoteFails(t *testing.T) {
	dir := newRepo(t)
	if err := Push(context.Background(), dir, ""); err == nil {
		t.Fatal("push without origin should fail")
	}
}

func TestIsRepo(t *testing.T) {
	gitAvailable(t)
	dir := t.TempDir()
	if IsRepo(dir) {
		t.Fatal("empty directory reported as a repo")
	}
	if !IsRepo(newRepo(t)) {
		t.Fatal("initialized repository not detected")
	}
}
