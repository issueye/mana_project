package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"mana/internal/model"
	"mana/internal/service"
	"mana/internal/store"
)

// Page identifies the visible page.
type Page int

const (
	PageHome Page = iota
	PageGroups
	PageProject
	PageEditor
	PageStats
)

type App struct {
	svc    *service.Service
	store  *store.Store
	win    *mygo.Window
	theme  ui.Theme
	dbPath string

	page        Page
	groupFilter int64 // 0 all, -1 ungrouped, >0 group ID
	groupTab    int   // 0 projects, 1 documents
	projectView int   // 0 documents, 1 files, 2 repository, 3 todos
	projectID   int64
	docID       int64
	editorFrom  Page
	groupReady  bool

	groups   []model.Group
	projects []model.Project
	docs     []model.Document
	todos    []model.Todo

	currentProject model.Project
	currentDoc     model.Document

	// The project page's side views: the file browser (F14) and the Git
	// panel (F15), both loaded off the frame.
	explorer explorerState
	repo     repoState

	// The editor's buffer and state live across frames so the writing
	// paper keeps the caret, scroll and line numbers.
	docBuf      *editor.Buffer
	editorState editor.State

	// The viewer state of a file document (F13): loaded in the
	// background, keyed by document ID.
	viewer      viewerFile
	viewerSheet string

	// The preview card on the project page loads the newest document's
	// body once per document.
	previewDocID int64
	previewBody  string

	projectTable   ui.ListState
	homeTable      ui.ListState
	groupDocTable  ui.ListState
	detailDocTable ui.ListState
	groupList      ui.ListState
	homeProjectRow int
	projectRow     int
	docRow         int

	searchQuery    string
	searchHandle   ui.Handle
	groupQuery     string
	groupListQuery string
	tagFilter      string
	sortMode       string

	// The project page's todo checklist (F17) and its note editor (F16).
	todoInput   string
	todoFilter  int // 0 all, 1 undone, 2 done
	noteEditing bool
	noteDraft   string

	// The command palette (CTRL K) and its debounced search.
	paletteOpen     bool
	paletteQuery    string
	paletteSearched string
	paletteResults  model.SearchResult
	paletteSel      int
	paletteGen      int

	// The settings sheet and its backup flow.
	showSettings bool
	dbSize       int64

	// The move-to-group dialog.
	showMoveDialog bool
	moveProjectID  int64
	moveChoice     string

	showProjectDialog bool
	showGroupDialog   bool
	showDeleteDialog  bool
	showDiscardDialog bool
	projectEditing    int64
	groupEditing      int64
	deleteKind        string
	deleteID          int64
	deleteTitle       string
	deleteMessage     string
	projectForm       model.ProjectInput
	groupForm         model.GroupInput
	groupChoice       string
	docForm           model.DocumentInput
	docFormID         int64
	docInitial        model.DocumentInput
	docKindChoice     string
	editorMode        int
	fieldErr          map[string]string

	loading bool
	err     string
	notice  string
}

// New creates an app with an empty state. Call Start, or NewWithService, before
// showing its view.
func New() *App {
	return &App{
		page:           PageHome,
		fieldErr:       map[string]string{},
		theme:          newTheme(),
		tagFilter:      "全部标签",
		sortMode:       "更新时间",
		homeProjectRow: -1,
		projectRow:     -1,
		docRow:         -1,
	}
}

// NewWithService creates an app for tests with an already opened service.
func NewWithService(svc *service.Service) *App {
	a := New()
	a.svc = svc
	a.store = svc.Store()
	if err := a.Load(); err != nil {
		a.err = service.FormatError(err)
	}
	return a
}

// Start opens the database in the user's data directory and loads the data.
func (a *App) Start() error {
	if a.svc != nil {
		return a.Load()
	}
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return fmt.Errorf("user data directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	st, err := store.Open(context.Background(), filepath.Join(dir, "mana.db"))
	if err != nil {
		return err
	}
	st.SetFilesDir(dir)
	a.store = st
	a.svc = service.New(st)
	a.dbPath = filepath.Join(dir, "mana.db")
	return a.Load()
}

// AttachWindow gives the app the window used for asynchronous updates and
// dialogs.
func (a *App) AttachWindow(win *mygo.Window) { a.win = win }

// Close closes the database.
func (a *App) Close() {
	if a.store != nil {
		a.store.Close()
	}
}

// SeedDemo inserts sample data when the database is empty.
func (a *App) SeedDemo(ctx context.Context) error {
	if a.store == nil {
		return errors.New("app: no database")
	}
	return a.store.SeedDemo(ctx)
}

// Load reads a consistent snapshot from the service.
func (a *App) Load() error {
	if a.svc == nil {
		return errors.New("app: no service")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	snap, err := a.svc.Snapshot(ctx)
	if err != nil {
		return err
	}
	a.applySnapshot(snap)
	return nil
}

func (a *App) applySnapshot(snap model.Snapshot) {
	a.groups = snap.Groups
	a.projects = snap.Projects
	a.docs = snap.Docs
	a.todos = snap.Todos
	a.err = ""
	a.ensureSelection()
}

func (a *App) ensureSelection() {
	if a.projectID != 0 {
		if p, ok := a.projectByID(a.projectID); ok {
			a.currentProject = p
		} else {
			a.projectID = 0
			a.currentProject = model.Project{}
		}
	}
	if a.groupFilter > 0 {
		found := false
		for _, g := range a.groups {
			if g.ID == a.groupFilter {
				found = true
				break
			}
		}
		if !found {
			a.groupFilter = 0
		}
	}
}

func (a *App) reload() {
	if err := a.Load(); err != nil {
		a.notice = service.FormatError(err)
	}
}

func (a *App) projectByID(id int64) (model.Project, bool) {
	for _, p := range a.projects {
		if p.ID == id {
			return p, true
		}
	}
	return model.Project{}, false
}

func (a *App) groupByID(id int64) (model.Group, bool) {
	for _, g := range a.groups {
		if g.ID == id {
			return g, true
		}
	}
	return model.Group{}, false
}

func (a *App) groupByName(name string) (model.Group, bool) {
	for _, g := range a.groups {
		if g.Name == name {
			return g, true
		}
	}
	return model.Group{}, false
}

func (a *App) docByID(id int64) (model.Document, bool) {
	for _, d := range a.docs {
		if d.ID == id {
			return d, true
		}
	}
	return model.Document{}, false
}

// currentGroup returns the selected group's presentation data, including the
// virtual all/ungrouped entries.
func (a *App) currentGroup() (name, description, color string) {
	if a.groupFilter == 0 {
		return "全部项目", "汇总所有项目组与未分组中的项目，用于全局浏览。", "#33586B"
	}
	if a.groupFilter == -1 {
		return "未分组", "尚未归入项目组的项目；可随时移动到任意项目组。", "#A39B87"
	}
	if g, ok := a.groupByID(a.groupFilter); ok {
		return g.Name, g.Description, g.Color
	}
	return "全部项目", "汇总所有项目组与未分组中的项目，用于全局浏览。", "#33586B"
}

func (a *App) visibleProjects() []model.Project {
	out := make([]model.Project, 0, len(a.projects))
	for _, p := range a.projects {
		if a.groupFilter > 0 && (p.GroupID == nil || *p.GroupID != a.groupFilter) {
			continue
		}
		if a.groupFilter == -1 && p.GroupID != nil {
			continue
		}
		if a.tagFilter != "" && a.tagFilter != "全部标签" && !projectHasTag(p, a.tagFilter) {
			continue
		}
		if a.groupQuery != "" && !projectMatches(p, a.groupQuery) {
			continue
		}
		out = append(out, p)
	}
	if a.sortMode == "名称" {
		sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	} else {
		sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	}
	return out
}

func (a *App) tagOptions() []string {
	options := []string{"全部标签"}
	seen := map[string]bool{"全部标签": true}
	for _, p := range a.projects {
		for _, tag := range p.Tags {
			if !seen[tag] {
				seen[tag] = true
				options = append(options, tag)
			}
		}
	}
	return options
}

func projectHasTag(p model.Project, tag string) bool {
	for _, t := range p.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

func (a *App) groupDocs() []model.Document {
	if a.groupFilter <= 0 {
		return nil
	}
	out := make([]model.Document, 0)
	for _, d := range a.docs {
		if d.OwnerType == model.OwnerGroup && d.OwnerID == a.groupFilter {
			if a.groupQuery != "" && !docMatches(d, a.groupQuery) {
				continue
			}
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (a *App) projectDocs() []model.Document {
	out := make([]model.Document, 0)
	for _, d := range a.docs {
		if d.OwnerType == model.OwnerProject && d.OwnerID == a.projectID {
			if a.groupQuery != "" && !docMatches(d, a.groupQuery) {
				continue
			}
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (a *App) recentProjects(limit int) []model.Project {
	projects := append([]model.Project(nil), a.projects...)
	sort.Slice(projects, func(i, j int) bool { return projects[i].UpdatedAt.After(projects[j].UpdatedAt) })
	if len(projects) > limit {
		projects = projects[:limit]
	}
	return projects
}

func (a *App) recentDocs(limit int) []model.Document {
	docs := append([]model.Document(nil), a.docs...)
	sort.Slice(docs, func(i, j int) bool { return docs[i].UpdatedAt.After(docs[j].UpdatedAt) })
	if len(docs) > limit {
		docs = docs[:limit]
	}
	return docs
}

func (a *App) stats() (projects, groups, docs, week int) {
	projects = len(a.projects)
	groups = len(a.groups)
	docs = len(a.docs)
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	for _, p := range a.projects {
		if p.UpdatedAt.After(cutoff) {
			week++
		}
	}
	return
}

func projectMatches(p model.Project, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	if strings.Contains(strings.ToLower(p.Name), q) ||
		strings.Contains(strings.ToLower(p.Path), q) ||
		strings.Contains(strings.ToLower(p.Description), q) {
		return true
	}
	for _, tag := range p.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func docMatches(d model.Document, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(d.Title), q) ||
		strings.Contains(strings.ToLower(d.OwnerName), q)
}

func (a *App) goHome() {
	a.page = PageHome
}

func (a *App) goGroups() {
	a.page = PageGroups
	if !a.groupReady {
		a.groupReady = true
		if len(a.groups) > 0 {
			a.groupFilter = a.groups[0].ID
		}
	}
	if a.groupFilter > 0 {
		if _, ok := a.groupByID(a.groupFilter); !ok {
			a.groupFilter = 0
		}
	}
}

func (a *App) openGroup(id int64) {
	a.groupFilter = id
	a.groupTab = 0
	a.groupQuery = ""
	a.tagFilter = "全部标签"
	a.page = PageGroups
}

func (a *App) openProject(id int64) {
	p, ok := a.projectByID(id)
	if !ok {
		a.notice = "项目不存在或已被删除"
		return
	}
	if a.projectID != id {
		a.projectView = 0
	}
	a.projectID = id
	a.currentProject = p
	a.page = PageProject
	a.groupQuery = ""
	a.homeProjectRow = -1
	a.projectRow = -1
}

func (a *App) backToGroups() {
	a.page = PageGroups
	if p, ok := a.projectByID(a.projectID); ok && p.GroupID != nil {
		a.groupFilter = *p.GroupID
	} else if a.projectID != 0 {
		a.groupFilter = -1
	}
}

// openDoc loads a document into the editor.
func (a *App) openDoc(id int64) {
	if a.svc == nil {
		return
	}
	d, err := a.svc.Document(context.Background(), id)
	if err != nil {
		a.notice = service.FormatError(err)
		return
	}
	a.docID = id
	a.docFormID = id
	a.currentDoc = d
	a.docForm = model.DocumentInput{
		OwnerType: d.OwnerType, OwnerID: d.OwnerID,
		Title: d.Title, Kind: d.Kind, Content: d.Content, Link: d.Link,
	}
	a.docInitial = a.docForm
	a.docKindChoice = model.KindLabel(d.Kind)
	a.editorMode = 0
	a.editorFrom = a.page
	a.page = PageEditor
	a.docRow = -1
	a.fieldErr = map[string]string{}
	a.resetDocBuffer(d.Content)
}

// newDoc opens the editor for a new document owned by ownerType/ownerID.
func (a *App) newDoc(ownerType string, ownerID int64) {
	a.docID = 0
	a.docFormID = 0
	a.currentDoc = model.Document{}
	a.docForm = model.DocumentInput{
		OwnerType: ownerType, OwnerID: ownerID,
		Title: "", Kind: "other", Content: "", Link: "",
	}
	a.docInitial = a.docForm
	a.docKindChoice = "其他"
	a.editorMode = 0
	a.editorFrom = a.page
	a.page = PageEditor
	a.fieldErr = map[string]string{}
	a.resetDocBuffer("")
}

// resetDocBuffer points the writing paper at a fresh buffer with text.
func (a *App) resetDocBuffer(text string) {
	if a.docBuf == nil {
		a.docBuf = editor.NewBuffer(text)
	} else {
		a.docBuf.Reset(text)
	}
	a.editorState = editor.State{}
}

func (a *App) editorDirty() bool {
	return a.docForm.Title != a.docInitial.Title ||
		a.docContent() != a.docInitial.Content ||
		a.docForm.Kind != a.docInitial.Kind ||
		a.docForm.Link != a.docInitial.Link ||
		a.docKindChoice != model.KindLabel(a.docInitial.Kind)
}

func (a *App) backFromEditor() {
	if a.editorDirty() {
		a.showDiscardDialog = true
		return
	}
	a.leaveEditor()
}

func (a *App) leaveEditor() {
	switch a.editorFrom {
	case PageProject:
		a.page = PageProject
	case PageGroups:
		a.page = PageGroups
	default:
		a.page = PageHome
	}
}

func (a *App) saveDoc(c *ui.Context) {
	if a.svc == nil {
		return
	}
	a.syncDocBuffer()
	a.docForm.Kind = model.NormalizeKind(a.docKindChoice)
	a.fieldErr = map[string]string{}
	_, err := a.svc.SaveDocument(context.Background(), a.docFormID, a.docForm)
	if err != nil {
		var fe *service.FieldError
		if errors.As(err, &fe) {
			a.fieldErr[fe.Field] = fe.Message
			c.Toast(fe.Message)
			return
		}
		a.err = service.FormatError(err)
		c.Toast(a.err)
		return
	}
	d, err := a.svc.Snapshot(context.Background())
	if err == nil {
		a.applySnapshot(d)
	}
	a.docID = 0
	a.docInitial = a.docForm
	a.leaveEditor()
	c.Toast("文档已保存")
}

// openNewProject opens the project dialog for the current group.
func (a *App) openNewProject() {
	groupName := "未分组"
	if a.groupFilter > 0 {
		if g, ok := a.groupByID(a.groupFilter); ok {
			groupName = g.Name
		}
	}
	a.projectEditing = 0
	a.projectForm = model.ProjectInput{}
	a.groupChoice = groupName
	a.fieldErr = map[string]string{}
	a.showProjectDialog = true
}

func (a *App) openEditProject() {
	a.projectEditing = a.projectID
	p := a.currentProject
	a.projectForm = model.ProjectInput{
		Name: p.Name, GroupID: p.GroupID, Tags: append([]string(nil), p.Tags...),
		Description: p.Description, Path: p.Path, Version: p.Version,
	}
	a.groupChoice = "未分组"
	if p.GroupID != nil {
		if g, ok := a.groupByID(*p.GroupID); ok {
			a.groupChoice = g.Name
		}
	}
	a.fieldErr = map[string]string{}
	a.showProjectDialog = true
}

func (a *App) chosenGroupID() *int64 {
	if a.groupChoice == "" || a.groupChoice == "未分组" {
		return nil
	}
	for _, g := range a.groups {
		if g.Name == a.groupChoice {
			id := g.ID
			return &id
		}
	}
	return nil
}

func (a *App) submitProject(c *ui.Context) {
	if a.svc == nil {
		return
	}
	a.projectForm.GroupID = a.chosenGroupID()
	var err error
	if a.projectEditing == 0 {
		_, err = a.svc.CreateProject(context.Background(), a.projectForm)
	} else {
		_, err = a.svc.UpdateProject(context.Background(), a.projectEditing, a.projectForm)
	}
	if err != nil {
		var fe *service.FieldError
		if errors.As(err, &fe) {
			a.fieldErr[fe.Field] = fe.Message
			c.Toast(fe.Message)
			return
		}
		c.Toast(service.FormatError(err))
		return
	}
	a.showProjectDialog = false
	a.reload()
	if a.projectEditing != 0 {
		if p, ok := a.projectByID(a.projectID); ok {
			a.currentProject = p
		}
	}
	if a.projectForm.Path != "" {
		if _, err := os.Stat(a.projectForm.Path); err != nil {
			c.Toast("项目已保存；路径不存在，已仅登记")
			return
		}
	}
	c.Toast("项目已保存")
}

func (a *App) submitGroup(c *ui.Context) {
	if a.svc == nil {
		return
	}
	var err error
	if a.groupEditing == 0 {
		_, err = a.svc.CreateGroup(context.Background(), a.groupForm)
	} else {
		_, err = a.svc.UpdateGroup(context.Background(), a.groupEditing, a.groupForm)
	}
	if err != nil {
		var fe *service.FieldError
		if errors.As(err, &fe) {
			a.fieldErr[fe.Field] = fe.Message
			c.Toast(fe.Message)
			return
		}
		c.Toast(service.FormatError(err))
		return
	}
	a.showGroupDialog = false
	a.reload()
	c.Toast("项目组已保存")
}

func (a *App) openNewGroup() {
	a.groupEditing = 0
	a.groupForm = model.GroupInput{Color: "#1677FF"}
	a.fieldErr = map[string]string{}
	a.showGroupDialog = true
}

func (a *App) openEditGroup() {
	a.groupEditing = a.groupFilter
	g := model.Group{Name: "", Description: "", Color: "#1677FF"}
	if a.groupFilter > 0 {
		if gg, ok := a.groupByID(a.groupFilter); ok {
			g = gg
		}
	}
	a.groupForm = model.GroupInput{Name: g.Name, Description: g.Description, Color: g.Color}
	a.fieldErr = map[string]string{}
	a.showGroupDialog = true
}

func (a *App) askDeleteProject() {
	a.deleteKind = "project"
	a.deleteID = a.currentProject.ID
	a.deleteTitle = "删除项目 “" + a.currentProject.Name + "”？"
	a.deleteMessage = "项目会从索引中移除，受影响的记录如下。"
	a.showDeleteDialog = true
}

func (a *App) askDeleteGroup() {
	if a.groupFilter <= 0 {
		return
	}
	a.deleteKind = "group"
	a.deleteID = a.groupFilter
	g, _ := a.groupByID(a.groupFilter)
	a.deleteTitle = "删除项目组 “" + g.Name + "”？"
	a.deleteMessage = "项目组会从索引中移除，受影响的记录如下。"
	a.showDeleteDialog = true
}

func (a *App) askDeleteDoc(id int64) {
	d, ok := a.docByID(id)
	if !ok {
		return
	}
	a.deleteKind = "doc"
	a.deleteID = id
	a.deleteTitle = "删除文档 “" + d.Title + "”？"
	a.deleteMessage = "这篇文档会从记录中移除。"
	a.showDeleteDialog = true
}

func (a *App) confirmDelete(c *ui.Context) {
	if a.svc == nil {
		return
	}
	var err error
	switch a.deleteKind {
	case "project":
		err = a.svc.DeleteProject(context.Background(), a.deleteID)
		if a.page == PageProject {
			a.page = PageGroups
		}
	case "group":
		err = a.svc.DeleteGroup(context.Background(), a.deleteID)
		a.groupFilter = 0
		a.page = PageGroups
	case "doc":
		err = a.svc.DeleteDocument(context.Background(), a.deleteID)
		if a.page == PageEditor {
			a.leaveEditor()
		}
	}
	if err != nil {
		c.Toast(service.FormatError(err))
		return
	}
	a.showDeleteDialog = false
	a.reload()
	c.Toast("已删除")
}

// View is the window's frame function.
func (a *App) View(c *ui.Context) {
	c.SetTheme(&a.theme)
	a.shortcuts(c)
	ui.Column(c).Fill().Children(func() {
		a.windowBar(c)
		ui.Row(c).Grow(1).FillWidth().Children(func() {
			a.iconRail(c)
			a.contextPanel(c)
			ui.Column(c).Grow(1).FillHeight().Background(colSheet).Children(func() {
				a.content(c)
				a.statusBar(c)
			})
		})
	})
	a.dialogs(c)
	if a.notice != "" {
		c.Toast(a.notice)
		a.notice = ""
	}
}

func (a *App) shortcuts(c *ui.Context) {
	if c.Shortcut(ui.Cmd, ui.KeyK) {
		if a.paletteOpen {
			a.paletteOpen = false
		} else {
			a.openPalette()
		}
	}
	if c.Shortcut(ui.Cmd, ui.KeyN) {
		if a.page != PageEditor {
			a.openNewProject()
		}
	}
	if c.Shortcut(ui.Cmd|ui.Shift, ui.KeyN) {
		a.openNewGroup()
	}
	if c.Shortcut(ui.Cmd, ui.KeyF) {
		a.searchHandle.Focus(c)
	}
	if a.page == PageEditor && c.Shortcut(ui.Cmd, ui.KeyS) {
		a.saveDoc(c)
	}
	if a.showProjectDialog && c.Shortcut(ui.Cmd, ui.KeyEnter) {
		a.submitProject(c)
	}
	if a.showGroupDialog && c.Shortcut(ui.Cmd, ui.KeyEnter) {
		a.submitGroup(c)
	}
	if a.showMoveDialog && c.Shortcut(ui.Cmd, ui.KeyEnter) {
		a.submitMove(c)
	}
}

func (a *App) content(c *ui.Context) {
	switch a.page {
	case PageGroups:
		a.groupsPage(c)
	case PageProject:
		a.projectPage(c)
	case PageEditor:
		a.editorPage(c)
	case PageStats:
		a.statsPage(c)
	default:
		a.homePage(c)
	}
}
