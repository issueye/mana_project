package app

import (
	"context"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

// paletteRow is one selectable line of the command palette: a search hit
// or a built-in action.
type paletteRow struct {
	kind      string // "project" | "doc" | "group" | "action"
	id        int64
	actionIdx int
	label     string
	sub       string
}

var paletteActions = []struct {
	label  string
	hint   string
	run    func(*App)
	header string
}{
	{"新建项目", "CTRL N", func(a *App) { a.openNewProject() }, "操作"},
	{"新建项目组", "CTRL SHIFT N", func(a *App) { a.openNewGroup() }, ""},
	{"打开统计", "", func(a *App) { a.page = PageStats }, ""},
	{"备份数据…", "", func(a *App) { a.openSettings() }, ""},
	{"打开设置…", "", func(a *App) { a.openSettings() }, ""},
}

// openPalette resets and shows the command palette.
func (a *App) openPalette() {
	a.paletteOpen = true
	a.paletteQuery = ""
	a.paletteSearched = ""
	a.paletteResults = model.SearchResult{}
	a.paletteSel = 0
}

// paletteDispatch runs the search for the query once per change. Real
// windows search after a 150ms debounce; headless tests search inline.
func (a *App) paletteDispatch() {
	q := strings.TrimSpace(a.paletteQuery)
	if q == a.paletteSearched {
		return
	}
	a.paletteSearched = q
	a.paletteGen++
	if r := []rune(q); len(r) < 2 {
		a.paletteResults = model.SearchResult{}
		a.paletteSel = 0
		return
	}
	id := a.paletteGen
	if a.svc == nil || a.win == nil {
		if a.svc == nil {
			return
		}
		res, err := a.svc.Search(context.Background(), q)
		if err == nil {
			a.paletteResults = res
			a.paletteSel = 0
		}
		return
	}
	go func() {
		time.Sleep(150 * time.Millisecond)
		res, err := a.svc.Search(context.Background(), q)
		if err != nil {
			return
		}
		a.win.Update(func() {
			if a.paletteGen != id {
				return
			}
			a.paletteResults = res
			a.paletteSel = 0
		})
	}()
}

// paletteRows flattens the search hits and the actions into one list the
// arrow keys walk.
func (a *App) paletteRows() []paletteRow {
	rows := make([]paletteRow, 0, 20)
	for _, p := range a.paletteResults.Projects {
		rows = append(rows, paletteRow{kind: "project", id: p.ID, label: p.Name,
			sub: groupOrUngrouped(a, p) + " · " + pathOrNone(p.Path)})
	}
	for _, d := range a.paletteResults.Docs {
		rows = append(rows, paletteRow{kind: "doc", id: d.ID, label: d.Title,
			sub: d.OwnerName + " · " + model.KindLabel(d.Kind)})
	}
	for _, g := range a.paletteResults.Groups {
		rows = append(rows, paletteRow{kind: "group", id: g.ID, label: g.Name,
			sub: "项目组 · " + itoa(g.ProjectCount) + " 个项目"})
	}
	for i := range paletteActions {
		rows = append(rows, paletteRow{kind: "action", actionIdx: i, label: paletteActions[i].label})
	}
	return rows
}

// paletteStep moves the selection within the flattened rows.
func (a *App) paletteStep(delta int) {
	n := len(a.paletteRows())
	if n == 0 {
		return
	}
	a.paletteSel = (a.paletteSel + delta + n) % n
}

// paletteActivate opens the row the palette points at.
func (a *App) paletteActivate(idx int) {
	rows := a.paletteRows()
	if idx < 0 || idx >= len(rows) {
		return
	}
	row := rows[idx]
	a.paletteOpen = false
	switch row.kind {
	case "project":
		a.openProject(row.id)
	case "doc":
		a.openDoc(row.id)
	case "group":
		a.openGroup(row.id)
	case "action":
		paletteActions[row.actionIdx].run(a)
	}
}

// paletteDialog is the modal quick-open panel behind CTRL K. The arrow
// and enter handling sits on the panel itself so it works while the
// search field keeps the focus.
func (a *App) paletteDialog(c *ui.Context) {
	ui.DialogBase(c, &a.paletteOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(23, 21, 15, 0.42))
		panel.Width(640).Background(colSheet).Radius(2).Border(1, colLine).
			Shadow(0, 18, 40, 0, ui.RGBA(23, 21, 15, 0.28))
		panel.OnShortcut(0, ui.KeyDown, func() { a.paletteStep(1) })
		panel.OnShortcut(0, ui.KeyUp, func() { a.paletteStep(-1) })
		a.paletteDispatch()
		ui.Column(c).FillWidth().Children(func() {
			ui.Row(c).Height(52).Padding(0, 16).Gap(9).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icons.search).FontSize(16).TextColor(colInk3)
				ui.TextInputBase(c, &a.paletteQuery).Grow(1).FontSize(14).
					Placeholder("搜索项目、文档、项目组，或执行操作…").AutoFocus().
					OnSubmit(func() { a.paletteActivate(a.paletteSel) })
				monoText(c, "ESC", 9).TextColor(colInk4)
			})
			hairline(c)
			rows := a.paletteRows()
			height := float32(len(rows))*36 + 12
			if height > 380 {
				height = 380
			}
			if len(rows) == 0 {
				ui.Column(c).Height(88).Center().Children(func() {
					ui.Text(c, "未找到与 “"+strings.TrimSpace(a.paletteQuery)+"” 相关的内容").
						FontSize(11.5).TextColor(colInk3)
				})
			} else {
				ui.Scroll(c).Height(height).Children(func() {
					ui.Column(c).Padding(6, 8, 8, 8).Gap(1).Children(func() {
						a.paletteSection(c, &rows, "项目", "project")
						a.paletteSection(c, &rows, "文档", "doc")
						a.paletteSection(c, &rows, "项目组", "group")
						a.paletteSection(c, &rows, "操作", "action")
					})
				})
			}
			hairline(c)
			ui.Row(c).Height(32).Padding(0, 16).AlignItems(ui.Center).Children(func() {
				monoText(c, "↑↓ 选择 · ENTER 打开 · ESC 关闭", 9).TextColor(colInk4)
				ui.Spacer(c)
				monoText(c, itoa(len(rows))+" 项", 9).TextColor(colInk4)
			})
		})
	})
}

// paletteSection draws one titled group of palette rows, keeping the
// selection index aligned with the flat list. Sections without rows are
// skipped entirely.
func (a *App) paletteSection(c *ui.Context, rows *[]paletteRow, title, kind string) {
	all := *rows
	present := false
	for _, row := range all {
		if row.kind == kind {
			present = true
			break
		}
	}
	if !present {
		return
	}
	ui.Row(c).Height(22).Padding(6, 8, 0, 8).Children(func() {
		microLabel(c, title)
	})
	for i, row := range all {
		if row.kind != kind {
			continue
		}
		idx := i
		selected := idx == a.paletteSel
		item := ui.Row(c).Height(36).Padding(0, 10).Gap(9).Radius(2).AlignItems(ui.Center).
			Focusable().Role(ui.RoleButton).Label(row.label)
		switch {
		case selected:
			item.Background(colSealSoft)
		case item.Hovered():
			item.Background(colSheet2)
		}
		item.OnClick(func() { a.paletteActivate(idx) })
		item.Children(func() {
			switch row.kind {
			case "project":
				ui.Icon(c, icons.grid).FontSize(13).TextColor(colInk3)
			case "doc":
				ui.Icon(c, icons.file).FontSize(13).TextColor(colInk3)
			case "group":
				ui.Icon(c, icons.folder).FontSize(13).TextColor(colInk3)
			default:
				ui.Icon(c, icons.chevronRight).FontSize(13).TextColor(colInk3)
			}
			ui.Column(c).Grow(1).Gap(0).Children(func() {
				ui.Text(c, row.label).FontSize(12).FontWeight(600).SingleLine()
				if row.sub != "" {
					ui.Text(c, row.sub).FontSize(9.5).TextColor(colInk4).SingleLine()
				}
			})
			if row.kind == "action" && paletteActions[row.actionIdx].hint != "" {
				monoText(c, paletteActions[row.actionIdx].hint, 9).TextColor(colInk4)
			}
		})
	}
}
