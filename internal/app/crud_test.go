package app

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"mana/internal/service"
	"mana/internal/store"
)

func emptyApp(t *testing.T) *App {
	t.Helper()
	st, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return NewWithService(service.New(st))
}

func TestEmptyHome(t *testing.T) {
	a := emptyApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if _, ok := tt.Find("还没有项目"); !ok {
		t.Fatal("empty home did not render its empty state")
	}
	if _, ok := tt.Find("新建项目"); !ok {
		t.Fatal("empty home has no action")
	}
	saveShot(t, tt, "home-empty")
}

func TestCreateGroupThroughDialog(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("新建项目组"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("创建项目组"); !ok {
		t.Fatal("new group dialog missing")
	}
	tt.Type("AI 实验室")
	tt.Frame()
	if err := tt.Click("创建项目组"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.groups) != 6 {
		t.Fatalf("groups = %d, want 6", len(a.groups))
	}
	if a.groups[5].Name != "AI 实验室" {
		t.Fatalf("new group = %q", a.groups[5].Name)
	}
}

func TestCreateProjectValidation(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("新建项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("创建项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.showProjectDialog {
		t.Fatal("dialog closed on invalid name")
	}
	if a.fieldErr["name"] == "" {
		t.Fatal("no name error")
	}
	// The name field is focused; type a valid name and submit.
	if err := tt.Click("项目名称 *"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Aster Web")
	tt.Frame()
	if err := tt.Click("创建项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.showProjectDialog {
		t.Fatal("dialog stayed open after valid submit")
	}
	if len(a.projects) != 19 {
		t.Fatalf("projects = %d, want 19", len(a.projects))
	}
}

func TestDeleteDocument(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("项目组管理"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Mana Desktop"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("文档操作"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("确认删除"); !ok {
		t.Fatal("delete dialog missing")
	}
	if err := tt.Click("确认删除"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.docs) != 41 {
		t.Fatalf("docs = %d, want 41", len(a.docs))
	}
}

func TestDeleteProject(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	a.openProject(a.projects[0].ID)
	tt.Frame()
	if err := tt.Click("删除项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("确认删除"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.projects) != 17 {
		t.Fatalf("projects = %d, want 17", len(a.projects))
	}
	if a.page != PageGroups {
		t.Fatalf("page = %v, want groups", a.page)
	}
}

func TestDeleteGroup(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("项目组管理"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("删除项目组"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("确认删除"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.groups) != 4 {
		t.Fatalf("groups = %d, want 4", len(a.groups))
	}
	// The five projects of the deleted group now live in ungrouped.
	ungrouped := 0
	for _, p := range a.projects {
		if p.GroupID == nil {
			ungrouped++
		}
	}
	if ungrouped != 7 {
		t.Fatalf("ungrouped = %d, want 7", ungrouped)
	}
}

func TestCreateAndSaveDocument(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	a.openProject(a.projects[0].ID)
	tt.Frame()
	if err := tt.Click("新建文档"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.page != PageEditor {
		t.Fatalf("page = %v, want editor", a.page)
	}
	tt.Type("发布说明")
	tt.Frame()
	if a.docForm.Title != "发布说明" {
		t.Fatalf("title = %q", a.docForm.Title)
	}
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.page != PageProject {
		t.Fatalf("page after save = %v, want project", a.page)
	}
	if len(a.docs) != 43 {
		t.Fatalf("docs = %d, want 43", len(a.docs))
	}
}
