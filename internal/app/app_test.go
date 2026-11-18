package app

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"

	"mana/internal/service"
	"mana/internal/store"
)

func demoApp(t *testing.T) *App {
	t.Helper()
	st, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SeedDemo(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return NewWithService(service.New(st))
}

func saveShot(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	dir := filepath.Join("testdata", "screens")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestHomeRenders(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if _, ok := tt.Find("最近更新项目"); !ok {
		t.Fatal("home did not render the recent projects card")
	}
	if _, ok := tt.Find("Mana Desktop"); !ok {
		t.Fatal("home did not render the sample project")
	}
	saveShot(t, tt, "home")
}

func TestGroupPageRenders(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("项目组管理"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("客户端应用"); !ok {
		t.Fatal("group page did not render")
	}
	saveShot(t, tt, "groups")
}

func TestProjectAndEditorRender(t *testing.T) {
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
	if _, ok := tt.Find("项目信息"); !ok {
		t.Fatal("project detail did not render")
	}
	saveShot(t, tt, "project")
	if err := tt.Click("需求文档 v1.0"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	t.Logf("page after click: %v, docID=%d, title=%q", a.page, a.docFormID, a.docForm.Title)
	if _, ok := tt.Find("Markdown 实时预览"); !ok {
		for _, text := range tt.Texts() {
			t.Logf("text: %q", text)
		}
		t.Fatal("editor did not render")
	}
	saveShot(t, tt, "editor")
}

func TestDialogsRender(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("新建项目"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("所属项目组"); !ok {
		t.Fatal("project dialog did not render")
	}
	saveShot(t, tt, "dialog-project")
}

func TestGroupDocsRender(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 900)
	tt.Frame()
	if err := tt.Click("项目组管理"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("文档 5"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("客户端设计规范"); !ok {
		t.Fatal("group documents did not render")
	}
	saveShot(t, tt, "groups-docs")
}

func TestNativeWindowContentSize(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1440, 868)
	tt.Frame()
	a.page = PageGroups
	if len(a.groups) > 0 {
		a.groupFilter = a.groups[0].ID
	}
	tt.Frame()
	saveShot(t, tt, "native-groups")
	a.openProject(a.projects[0].ID)
	tt.Frame()
	saveShot(t, tt, "native-project")
	a.openDoc(a.docs[0].ID)
	tt.Frame()
	if _, ok := tt.Find("Ctrl S 保存"); !ok {
		t.Fatal("editor status bar clipped at native content height")
	}
	saveShot(t, tt, "native-editor")
}

func TestMinimumPagesRender(t *testing.T) {
	a := demoApp(t)
	tt := ui.NewTester(a.View, 1120, 720)
	tt.Frame()
	saveShot(t, tt, "home-min")
	a.page = PageGroups
	if len(a.groups) > 0 {
		a.groupFilter = a.groups[0].ID
	}
	tt.Frame()
	if tt.FramePending() {
		t.Fatal("groups view did not settle")
	}
	saveShot(t, tt, "groups-min")

	a.openProject(a.projects[0].ID)
	tt.Frame()
	if _, ok := tt.Find("项目信息"); !ok {
		t.Fatal("project detail missing at minimum size")
	}
	saveShot(t, tt, "project-min")

	a.openDoc(a.docs[0].ID)
	tt.Frame()
	if _, ok := tt.Find("Markdown 实时预览"); !ok {
		t.Fatal("editor missing at minimum size")
	}
	saveShot(t, tt, "editor-min")
}
