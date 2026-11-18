package app

import (
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mana/internal/gitcmd"
	"mana/internal/model"
)

// explorerState is the file browser of a project: the directory being
// shown, its entries, and the generation guard for the async read.
type explorerState struct {
	projectID int64
	dir       string // slash-relative to the project root; "" is the root
	entries   []os.DirEntry
	loaded    bool
	loading   bool
	err       string
	gen       int
}

// openExplorer resets the browser for a project and starts the root read.
func (a *App) openExplorer(p model.Project) {
	a.explorer = explorerState{projectID: p.ID}
	a.loadExplorerDir(p, "")
}

// enterExplorerDir walks into a subdirectory (or "" for the root).
func (a *App) enterExplorerDir(p model.Project, rel string) {
	a.loadExplorerDir(p, rel)
}

// explorerAbsDir joins the relative dir onto the project root, refusing
// anything that climbs out of it.
func (a *App) explorerAbsDir(p model.Project) (string, string) {
	if p.Path == "" {
		return "", ""
	}
	rel := ""
	if a.explorer.dir != "" {
		rel = filepath.Clean(a.explorer.dir)
		if rel == ".." || len(rel) >= 2 && rel[:2] == ".." || filepath.IsAbs(rel) {
			rel = ""
		}
	}
	return filepath.Join(p.Path, filepath.FromSlash(rel)), rel
}

// loadExplorerDir reads a directory off the frame, guarding against
// stale results with a generation counter. rel is slash-relative to the
// project root; "" is the root.
func (a *App) loadExplorerDir(p model.Project, rel string) {
	if rel != "" {
		rel = filepath.ToSlash(filepath.Clean(rel))
		if rel == ".." || len(rel) >= 2 && rel[:2] == ".." || filepath.IsAbs(rel) {
			rel = ""
		}
	}
	a.explorer.dir = rel
	a.explorer.gen++
	id := a.explorer.gen
	a.explorer.loading = true
	a.explorer.loaded = false
	a.explorer.err = ""
	abs, _ := a.explorerAbsDir(p)
	if abs == "" {
		a.explorer.loading = false
		a.explorer.loaded = true
		a.explorer.err = "该项目未登记路径，请先在「编辑项目」中登记"
		return
	}
	if a.win == nil {
		entries, err := readExplorerDir(abs)
		a.explorerApply(id, a.explorer.dir, entries, err)
		return
	}
	go func() {
		entries, err := readExplorerDir(abs)
		a.win.Update(func() { a.explorerApply(id, rel, entries, err) })
	}()
}

func (a *App) explorerApply(id int, rel string, entries []os.DirEntry, err error) {
	if id != a.explorer.gen {
		return
	}
	a.explorer.loading = false
	a.explorer.loaded = true
	a.explorer.dir = rel
	if err != nil {
		a.explorer.entries = nil
		a.explorer.err = err.Error()
		return
	}
	a.explorer.entries = entries
}

// readExplorerDir lists one directory: folders first, then files, dot
// entries hidden.
func readExplorerDir(dir string) ([]os.DirEntry, error) {
	all, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]os.DirEntry, 0, len(all))
	for _, e := range all {
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}
		out = append(out, e)
	}
	gitcmd.SortDirsFirst(out)
	return out, nil
}

// revealExplorerDir shows the current directory in the file manager.
func (a *App) revealExplorerDir(dir string) {
	if dir == "" {
		return
	}
	go func() {
		if err := mygo.Shell.OpenPath(dir); err != nil && a.win != nil {
			a.win.Update(func() { a.notice = "打开目录失败：" + err.Error() })
		}
	}()
}

// openExplorerFile hands a file to the system's default program.
func (a *App) openExplorerFile(path string) {
	go func() {
		if err := mygo.Shell.OpenPath(path); err != nil && a.win != nil {
			a.win.Update(func() { a.notice = "打开失败：" + err.Error() })
		}
	}()
}

// filesView is the 文件 tab: a breadcrumb, the tool row and the listing.
func (a *App) filesView(c *ui.Context, p model.Project) {
	if a.explorer.projectID != p.ID {
		a.openExplorer(p)
	}
	abs, rel := a.explorerAbsDir(p)
	ui.Column(c).Grow(1).Padding(12, 24, 0, 24).Children(func() {
		ui.Row(c).Height(34).Gap(8).AlignItems(ui.Center).Children(func() {
			a.explorerCrumb(c, p, rel)
			ui.Spacer(c)
			if rel != "" {
				a.borderedButton(c, icons.arrowLeft, "上一级", func() {
					parent := ""
					if i := lastSlash(rel); i >= 0 {
						parent = rel[:i]
					}
					a.enterExplorerDir(p, parent)
				})
				ui.Box(c).Width(6)
			}
			a.ghostButton(c, icons.external, "在文件管理器中显示", func() { a.revealExplorerDir(abs) })
			a.ghostButton(c, icons.refresh, "刷新", func() { a.loadExplorerDir(p, rel) })
		})
		hairline(c)

		switch {
		case a.explorer.err != "":
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "无法读取这个目录", a.explorer.err)
			})
		case a.explorer.loading && !a.explorer.loaded:
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "正在读取…", abs)
			})
		case len(a.explorer.entries) == 0:
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "空文件夹", "这里没有可见的文件。")
			})
		default:
			a.explorerList(c, p, abs)
		}
		ui.Box(c).Height(12)
	})
}

// explorerCrumb shows 项目名 ▸ 子 ▸ 子 as clickable steps.
func (a *App) explorerCrumb(c *ui.Context, p model.Project, rel string) {
	crumb := ui.Row(c).Gap(5).AlignItems(ui.Center)
	crumb.Children(func() { a.crumbStep(c, p, p.Name, "", rel == "") })
	if rel == "" {
		return
	}
	parts := splitSlash(rel)
	walked := ""
	for i, part := range parts {
		at := i
		name := part
		walked = joinRel(walked, name)
		last := at == len(parts)-1
		crumb.Children(func() {
			ui.Text(c, "▸").FontSize(9).TextColor(colInk4)
		})
		stepRel := walked
		crumb.Children(func() { a.crumbStep(c, p, name, stepRel, last) })
	}
}

// crumbStep is one clickable breadcrumb segment.
func (a *App) crumbStep(c *ui.Context, p model.Project, name, rel string, last bool) {
	b := ui.Row(c).Height(22).Padding(0, 5).Radius(2).AlignItems(ui.Center).
		Focusable().Role(ui.RoleButton).Label(name)
	if !last && b.Hovered() {
		b.Background(colSheet2)
	}
	b.OnClick(func() { a.enterExplorerDir(p, rel) })
	b.Children(func() {
		t := ui.Text(c, name).FontSize(11).SingleLine()
		if last {
			t.FontWeight(600).TextColor(colInk)
		} else {
			t.TextColor(colInk3)
		}
	})
}

// explorerList draws the directory rows.
func (a *App) explorerList(c *ui.Context, p model.Project, abs string) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Gap(0).Children(func() {
			for i, e := range a.explorer.entries {
				entry := e
				idx := i
				isDir := entry.IsDir()
				row := ui.Row(c).MinHeight(38).Padding(0, 10).Gap(10).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label(entry.Name())
				switch {
				case row.Hovered():
					row.Background(colSheet2)
				case idx%2 == 1:
					row.Background(ui.Hex("#F8F4EA"))
				}
				row.OnClick(func() {
					childRel := joinRel(a.explorer.dir, entry.Name())
					if isDir {
						a.enterExplorerDir(p, childRel)
					} else {
						a.openExplorerFile(filepath.Join(abs, entry.Name()))
					}
				})
				row.Children(func() {
					if isDir {
						ui.Icon(c, icons.folder).FontSize(15).TextColor(colIndigo)
					} else {
						ui.Icon(c, icons.file).FontSize(15).TextColor(colInk3)
					}
					ui.Text(c, entry.Name()).FontSize(12).FontWeight(600).TextColor(colInk).SingleLine()
					ui.Spacer(c)
					if isDir {
						ui.Text(c, "文件夹").FontSize(9.5).TextColor(colInk4)
					} else if info, err := entry.Info(); err == nil {
						monoText(c, byteSize(info.Size()), 9.5).TextColor(colInk3)
						ui.Box(c).Width(14)
						ui.Text(c, model.Relative(info.ModTime())).FontSize(9.5).TextColor(colInk4)
					}
					if isDir {
						ui.Icon(c, icons.chevronRight).FontSize(12).TextColor(colInk4)
					}
				})
			}
		})
	})
}

// joinRel appends a name to a slash-relative path.
func joinRel(rel, name string) string {
	if rel == "" {
		return name
	}
	return rel + "/" + name
}

func splitSlash(rel string) []string {
	var out []string
	cur := ""
	for _, r := range rel {
		if r == '/' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}

func lastSlash(rel string) int {
	for i := len(rel) - 1; i >= 0; i-- {
		if rel[i] == '/' {
			return i
		}
	}
	return -1
}
