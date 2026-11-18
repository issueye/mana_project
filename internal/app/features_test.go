package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

// testCtx is a plain background context for the service calls the tests
// make outside the frame.
func testCtx() context.Context { return context.Background() }


// The command palette opens with CTRL K, lists the built-in actions, and
// searches projects and documents once two characters are typed.
func TestCommandPalette(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyK)
	tt.Frame()
	if !a.paletteOpen {
		t.Fatal("CTRL K did not open the palette")
	}
	if _, ok := tt.Find("操作"); !ok {
		t.Fatal("palette did not list its actions")
	}

	tt.Type("API")
	tt.Frame()
	if _, ok := tt.Find("Atlas API"); !ok {
		t.Fatal("palette did not find the project by name")
	}
	if _, ok := tt.Find("API 错误码规范"); !ok {
		t.Fatal("palette did not find the document by title")
	}
	saveShot(t, tt, "palette")

	// Down once selects the document row; Enter opens it in the editor.
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if a.paletteSel != 1 {
		t.Fatalf("paletteSel = %d, want 1", a.paletteSel)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.paletteOpen {
		t.Fatal("palette stayed open after activating a row")
	}
	if a.page != PageEditor {
		t.Fatalf("page = %v, want the editor", a.page)
	}
	if a.docForm.Title != "API 错误码规范" {
		t.Fatalf("opened %q", a.docForm.Title)
	}
}

// Activating the first action with an empty query opens the project
// dialog.
func TestCommandPaletteRunsActions(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	a.openPalette()
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.paletteOpen {
		t.Fatal("palette stayed open")
	}
	if !a.showProjectDialog {
		t.Fatal("first palette action did not open the new-project dialog")
	}
}

// Both move entries lead to the dialog, and confirming re-homes the
// project.
func TestMoveProjectDialog(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()

	target := a.projects[0]
	a.openProject(target.ID)
	tt.Frame()
	if _, ok := tt.Find("移入其他组"); !ok {
		t.Fatal("record card lost the move action")
	}

	a.openMoveProject(target.ID)
	tt.Frame()
	if !a.showMoveDialog {
		t.Fatal("move dialog did not open")
	}
	if _, ok := tt.Find("目标项目组"); !ok {
		t.Fatal("move dialog did not render")
	}
	if a.moveChoice != target.GroupName {
		t.Fatalf("moveChoice = %q, want the current group", a.moveChoice)
	}

	a.moveChoice = "后端服务"
	tt.Frame()
	saveShot(t, tt, "move-dialog")
	if err := tt.Click("移动项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.showMoveDialog {
		t.Fatal("move dialog stayed open")
	}
	p, ok := a.projectByID(target.ID)
	if !ok || p.GroupName != "后端服务" {
		t.Fatalf("project after move = %#v", p)
	}
}

// The settings sheet shows the data facts and the backup affordance.
func TestSettingsDialog(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	a.openSettings()
	tt.Frame()
	if !a.showSettings {
		t.Fatal("settings dialog did not open")
	}
	for _, text := range []string{"数据文件", "数据库大小", "立即备份", "在文件管理器中显示", "Mana v1.0.0"} {
		if _, ok := tt.Find(text); !ok {
			t.Fatalf("settings dialog lost %q", text)
		}
	}
	saveShot(t, tt, "settings")
}

// An uploaded file document opens into the file viewer: the paper header
// shows the file facts and the body shows the extracted text.
func TestFileDocumentViewer(t *testing.T) {
	a := demoApp(t)
	a.store.SetFilesDir(t.TempDir())
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()

	src := filepath.Join(t.TempDir(), "发布说明.txt")
	if err := os.WriteFile(src, []byte("版本 v1.0 已发布\n包含两行内容"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := a.svc.SaveDocumentFile(testCtx(), model.OwnerProject, a.projects[0].ID, src)
	if err != nil {
		t.Fatal(err)
	}
	a.reload()

	// The project table marks the attachment with its kind and size.
	a.openProject(a.projects[0].ID)
	tt.Frame()
	if _, ok := tt.Find("上传文档"); !ok {
		t.Fatal("project docs header lost the upload action")
	}
	if _, ok := tt.Find("发布说明"); !ok {
		t.Fatal("uploaded document missing from the table")
	}
	saveShot(t, tt, "project-files")

	a.openDoc(d.ID)
	tt.Frame()
	if _, ok := tt.Find("发布说明.txt"); !ok {
		t.Fatal("file viewer header missing the file name")
	}
	if _, ok := tt.Find("版本 v1.0 已发布"); !ok {
		t.Fatal("file viewer body missing the extracted text")
	}
	if _, ok := tt.Find("在系统中打开"); !ok {
		t.Fatal("file viewer lost the system opener")
	}
	saveShot(t, tt, "file-viewer")
}

// The project page gains 文件 and 仓库 tabs: the browser lists the
// working tree, the panel detects the git repository.
func TestProjectFilesTab(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()

	// Point the first project at a real directory with content.
	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte("# hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "internal", "app.go"), []byte("package app"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := a.projects[0]
	if _, err := a.svc.UpdateProject(testCtx(), p.ID, model.ProjectInput{
		Name: p.Name, GroupID: p.GroupID, Tags: p.Tags,
		Description: p.Description, Path: workDir, Version: p.Version,
	}); err != nil {
		t.Fatal(err)
	}
	a.reload()
	if fresh, ok := a.projectByID(p.ID); ok {
		p = fresh
	}
	a.openProject(p.ID)
	tt.Frame()

	if _, ok := tt.Find("文件"); !ok {
		t.Fatal("project page lost the files tab")
	}
	if err := tt.Click("文件"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("README.md"); !ok {
		t.Fatal("file browser did not list the project root")
	}
	saveShot(t, tt, "files-root")

	// Enter the folder, then go back up.
	if err := tt.Click("internal"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("app.go"); !ok {
		t.Fatal("file browser did not enter the folder")
	}
	saveShot(t, tt, "files-sub")
	if err := tt.Click("上一级"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("README.md"); !ok {
		t.Fatal("file browser did not return to the root")
	}
}

func TestProjectRepoTab(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	// A git repository with one commit and a modification.
	workDir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init")
	run("config", "user.name", "Mana")
	run("config", "user.email", "mana@example.com")
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "初始提交")
	if err := os.WriteFile(filepath.Join(workDir, "todo.txt"), []byte("待办"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := a.projects[0]
	if _, err := a.svc.UpdateProject(testCtx(), p.ID, model.ProjectInput{
		Name: p.Name, GroupID: p.GroupID, Tags: p.Tags,
		Description: p.Description, Path: workDir, Version: p.Version,
	}); err != nil {
		t.Fatal(err)
	}
	a.reload()
	if fresh, ok := a.projectByID(p.ID); ok {
		p = fresh
	}
	a.openProject(p.ID)
	tt.Frame()

	if err := tt.Click("仓库"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.repo.isRepo {
		t.Fatalf("repo not detected: %#v", a.repo)
	}
	if _, ok := tt.Find("改动"); !ok {
		t.Fatal("repo panel missing the status section")
	}
	if _, ok := tt.Find("todo.txt"); !ok {
		t.Fatal("repo panel missing the untracked file")
	}
	if _, ok := tt.Find("初始提交"); !ok {
		t.Fatal("repo panel missing the history")
	}
	saveShot(t, tt, "repo")

	// Stage the untracked file through the panel.
	if err := tt.Click("暂存 todo.txt"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	staged := false
	for _, f := range a.repo.files {
		if f.Path == "todo.txt" && f.Staged {
			staged = true
		}
	}
	if !staged {
		t.Fatalf("staging did not stick: %#v", a.repo.files)
	}

	// Commit it and watch history grow.
	a.repo.commitMsg = "登记待办清单"
	a.commitRepo(p)
	tt.Frame()
	if len(a.repo.commits) != 2 {
		t.Fatalf("commits = %d, want 2 (err=%q dirty=%q)", len(a.repo.commits), a.repo.errText, a.repo.dirtyMsg)
	}

	// Create a release tag through the form.
	a.repo.tagName = "v0.2.0"
	a.repo.tagMessage = "第二个内部版本"
	a.createRepoTag(p)
	tt.Frame()
	found := false
	for _, tg := range a.repo.tags {
		if tg.Name == "v0.2.0" {
			found = true
		}
	}
	if !found {
		t.Fatalf("tag missing: %#v", a.repo.tags)
	}
	saveShot(t, tt, "repo-after")
}

// The statistics page reuses the home ledger and adds the distributions.
func TestStatsPage(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	a.page = PageStats
	tt.Frame()
	if _, ok := tt.Find("数据统计"); !ok {
		t.Fatal("stats page did not render its header")
	}
	for _, text := range []string{"项目组分布", "近 7 天活跃", "文档类型", "标签 Top"} {
		if _, ok := tt.Find(text); !ok {
			t.Fatalf("stats page lost %q", text)
		}
	}
	if _, ok := tt.Find("客户端应用"); !ok {
		t.Fatal("stats page did not list the groups")
	}
	saveShot(t, tt, "stats")
}
