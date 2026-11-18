package app

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

// windowBar is the 38px ink bar: the brand seal, the index name and the
// local status on the left, the shortcuts, settings and window controls on
// the right. It drags the window everywhere but over the controls.
func (a *App) windowBar(c *ui.Context) {
	ui.Row(c).Height(38).Background(colInk).Children(func() {
		ui.Row(c).Grow(1).FillHeight().DragWindow().Padding(0, 14).Gap(10).
			AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(18, 18).Radius(2).Background(colSeal).Center().Children(func() {
				monoText(c, "M", 10).FontWeight(700).TextColor(colPaper)
			})
			ui.Text(c, "MANA").FontSize(11.5).FontWeight(700).TextColor(colBarText).LetterSpacing(1.2)
			ui.Box(c).Width(1).Height(14).Background(ui.RGBA(240, 235, 224, 0.18))
			ui.Text(c, "PROJECT INDEX · LOCAL").FontSize(9).TextColor(colBarMuted).LetterSpacing(1.5)
			ui.Spacer(c)
			ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(6, 6).Radius(1).Background(colMoss)
				ui.Text(c, "SQLITE · 已连接").FontSize(9).TextColor(colBarMuted).LetterSpacing(0.8)
			})
			a.barKey(c, "CTRL K", a.openPalette)
			a.barIcon(c, icons.settings, "设置", a.openSettings)
			a.barAvatar(c)
		})
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			a.windowControl(c, icons.minimize, "最小化", false, func() {
				if a.win != nil {
					a.win.Minimize()
				}
			})
			maximized := a.win != nil && a.win.IsMaximized()
			icon := icons.maximize
			label := "最大化"
			if maximized {
				icon = icons.restore
				label = "还原"
			}
			a.windowControl(c, icon, label, false, func() {
				if a.win == nil {
					return
				}
				if a.win.IsMaximized() {
					a.win.Unmaximize()
				} else {
					a.win.Maximize()
				}
			})
			a.windowControl(c, icons.close, "关闭", true, func() {
				if a.win != nil {
					a.win.Close()
				}
			})
		})
	})
}

// barKey is the bordered shortcut chip on the dark bar.
func (a *App) barKey(c *ui.Context, key string, fn func()) {
	b := ui.Row(c).Height(22).Padding(0, 8).Radius(2).
		Border(1, ui.RGBA(240, 235, 224, 0.22)).AlignItems(ui.Center).
		Focusable().Role(ui.RoleButton).Label("命令面板 " + key)
	if b.Hovered() {
		b.Background(colBarHover)
	}
	b.OnClick(fn)
	b.Children(func() {
		monoText(c, key, 9).TextColor(colBarMuted)
	})
}

func (a *App) barIcon(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(26, 26).Radius(2).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Hovered() {
		b.Background(colBarHover)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(14).TextColor(colBarMuted)
	})
}

func (a *App) barAvatar(c *ui.Context) {
	ui.Box(c).Size(24, 24).Radius(2).Background(colSeal).Center().
		Role(ui.RoleImage).Label("Lin").Children(func() {
		monoText(c, "L", 11).FontWeight(700).TextColor(colPaper)
	})
}

func (a *App) windowControl(c *ui.Context, icon *ui.SVG, label string, danger bool, fn func()) {
	b := ui.Box(c).Size(38, 38).Center().
		Focusable().Role(ui.RoleButton).Label(label)
	hover := colBarHover
	fg := colBarMuted
	if danger {
		hover = colSeal
	}
	if b.Hovered() {
		b.Background(hover)
		fg = colBarText
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(13).TextColor(fg)
	})
}

// iconRail is the 68px column of navigation icons on the ink ground; the
// current page marks itself with a seal tile.
func (a *App) iconRail(c *ui.Context) {
	ui.Column(c).Width(68).FillHeight().Background(colInk).
		Padding(10, 0, 12, 0).AlignItems(ui.Center).Children(func() {
		a.railItem(c, icons.home, "首页", PageHome, a.goHome)
		a.railItem(c, icons.folder, "项目组管理", PageGroups, a.goGroups)
		a.railItem(c, icons.grid, "统计", PageStats, func() { a.page = PageStats })
		a.railItem(c, icons.search, "搜索", Page(98), a.openPalette)
		a.railItem(c, icons.layers, "数据索引", Page(97), func() { c.Toast("数据索引将在后续版本开放") })
		ui.Spacer(c)
		a.railGhost(c, icons.database, "本地数据库", a.openSettings)
		a.railGhost(c, icons.settings, "设置", a.openSettings)
		ui.Box(c).Height(10)
		ui.Box(c).Size(28, 28).Radius(2).Background(colSeal).Center().
			Role(ui.RoleImage).Label("Lin").Children(func() {
			monoText(c, "L", 12).FontWeight(700).TextColor(colPaper)
		})
	})
}

func (a *App) railItem(c *ui.Context, icon *ui.SVG, label string, page Page, fn func()) {
	active := a.page == page
	tile := ui.Box(c).Size(46, 40).Radius(2).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	switch {
	case active:
		tile.Background(colSeal)
	case tile.Hovered():
		tile.Background(colBarHover)
	}
	tile.OnClick(fn)
	fg := colBarMuted
	if active {
		fg = colPaper
	}
	tile.Children(func() {
		ui.Icon(c, icon).FontSize(17).TextColor(fg)
	})
}

func (a *App) railGhost(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(46, 34).Radius(2).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Hovered() {
		b.Background(colBarHover)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(colBarMuted)
	})
}

// contextPanel is the index column that changes with the page: one place
// that always says where in the index the window is. It narrows to 240 on
// the smallest window.
func (a *App) contextPanel(c *ui.Context) {
	width := float32(252)
	if w, _ := c.Size(); w > 0 && w < 1200 {
		width = 240
	}
	ui.Column(c).Width(width).FillHeight().Background(colSheet2).
		BorderWidth(0, 0, 1, 0).BorderColor(colLine).Children(func() {
		switch a.page {
		case PageProject:
			a.projectContext(c)
		case PageEditor:
			a.editorContext(c)
		case PageGroups:
			a.groupsContext(c)
		case PageStats:
			a.statsContext(c)
		default:
			a.homeContext(c)
		}
	})
}

// panelHead is the common header of a context panel: a bold title with a
// monospace count and an optional trailing action.
func (a *App) panelHead(c *ui.Context, title, count string, action func()) {
	ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(12.5).FontWeight(600)
		if count != "" {
			monoText(c, count, 9.5).TextColor(colInk4)
		}
		ui.Spacer(c)
		if action != nil {
			a.ghostButton(c, icons.plus, "新建", action)
		}
	})
}

// homeContext lists the recent projects and the document type ledger.
func (a *App) homeContext(c *ui.Context) {
	ui.Column(c).FillHeight().Padding(14, 14, 12, 14).Gap(10).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Baseline).Children(func() {
			ui.Text(c, "索引目录").FontSize(12.5).FontWeight(600)
			microLabel(c, "MANA")
		})
		search := ui.SearchField(c, &a.searchQuery).
			Height(30).
			Bind(&a.searchHandle).
			Placeholder("搜索项目、文档、标签")
		if search.Submitted() && a.searchQuery != "" {
			a.groupQuery = a.searchQuery
			a.groupFilter = 0
			a.page = PageGroups
		}

		ui.Row(c).Padding(6, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			microLabel(c, "最近项目 · " + fmt.Sprintf("%02d", len(a.recentProjects(6))))
			ui.Spacer(c)
			microLabel(c, "RECENT")
		})
		for i, p := range a.recentProjects(6) {
			project := p
			a.contextProjectRow(c, itoaPad(i+1), project.Name,
				fmt.Sprintf("%s · %s", groupOrUngrouped(a, project), versionOrDash(project.Version)),
				false, func() { a.openProject(project.ID) })
		}

		ui.Row(c).Padding(8, 0, 2, 0).AlignItems(ui.Center).Children(func() {
			microLabel(c, "文档类型 · " + fmt.Sprintf("%02d", len(a.docs)))
			ui.Spacer(c)
			microLabel(c, "TYPE")
		})
		a.docTypeLedger(c)

		ui.Spacer(c)
		a.dashedButton(c, "新建项目组", a.openNewGroup)
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(5, 5).Radius(1).Background(colMoss)
			ui.Text(c, fmt.Sprintf("SQLite · %d 篇文档 · %s 同步", len(a.docs), time.Now().Format("15:04"))).
				FontSize(9).TextColor(colInk4).SingleLine()
		})
	})
}

// docTypeLedger draws the document kinds with hairline bars and counts.
func (a *App) docTypeLedger(c *ui.Context) {
	total := len(a.docs)
	if total == 0 {
		ui.Text(c, "还没有文档").FontSize(10.5).TextColor(colInk4)
		return
	}
	counts := map[string]int{}
	var order []string
	for _, d := range a.docs {
		k := model.KindLabel(d.Kind)
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	ui.Column(c).Gap(4).Children(func() {
		for _, k := range order {
			n := counts[k]
			frac := float32(n) / float32(total)
			ui.Row(c).Height(22).AlignItems(ui.Center).Children(func() {
				ui.Text(c, k).FontSize(10.5).TextColor(colInk2)
				ui.Spacer(c)
				ui.Box(c).Width(72).Height(2).Background(colLine2).Children(func() {
					ui.Box(c).Width(72 * frac).Height(2).Background(colInk3)
				})
				ui.Box(c).Width(8)
				monoText(c, fmt.Sprintf("%02d", n), 9.5).TextColor(colInk3)
			})
		}
	})
}

// groupsContext lists the groups with their identity colors and a small
// summary of the current selection.
func (a *App) groupsContext(c *ui.Context) {
	ui.Column(c).FillHeight().Padding(14, 14, 12, 14).Gap(8).Children(func() {
		a.panelHead(c, "项目组", fmt.Sprintf("%02d", len(a.groups)), a.openNewGroup)
		ui.SearchField(c, &a.groupListQuery).Height(30).Placeholder("搜索项目组")

		ui.Row(c).Padding(4, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			microLabel(c, "全部索引")
			ui.Spacer(c)
			microLabel(c, "ALL")
		})
		a.contextGroupRow(c, icons.layers, "全部项目", len(a.projects), a.groupFilter == 0, func() { a.openGroup(0) })

		ui.Row(c).Padding(6, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			microLabel(c, "项目组 · " + fmt.Sprintf("%02d", len(a.groups)))
			ui.Spacer(c)
			microLabel(c, "GROUPS")
		})
		ungrouped := 0
		for _, p := range a.projects {
			if p.GroupID == nil {
				ungrouped++
			}
		}
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Gap(1).Children(func() {
				for _, g := range a.groups {
					group := g
					if a.groupListQuery != "" && !containsFold(group.Name, a.groupListQuery) {
						continue
					}
					a.contextGroupRow(c, icons.folder, group.Name, group.ProjectCount,
						a.groupFilter == group.ID, func() { a.openGroup(group.ID) })
				}
				ui.Box(c).Height(1).Background(colLine2).BorderStyle(ui.BorderDashed).Margin(8, 2)
				a.contextGroupRow(c, icons.folderOpen, "未分组", ungrouped, a.groupFilter == -1, func() { a.openGroup(-1) })
			})
		})

		name, _, _ := a.currentGroup()
		ui.Column(c).Gap(3).Padding(8, 0, 4, 0).Children(func() {
			microLabel(c, "组概览 · SUMMARY")
			a.contextSummaryRow(c, "当前", name)
			a.contextSummaryRow(c, "项目", itoa(len(a.visibleProjects())))
			a.contextSummaryRow(c, "文档", itoa(len(a.groupDocs())))
			a.contextSummaryRow(c, "最近更新", groupUpdated(a.visibleProjects(), a.groupDocs()))
		})
		a.dashedButton(c, "新建项目组", a.openNewGroup)
	})
}

// projectContext is the project directory of the current group.
func (a *App) projectContext(c *ui.Context) {
	groupName, _, _ := a.currentGroup()
	projects := a.visibleProjects()
	ui.Column(c).FillHeight().Padding(14, 14, 12, 14).Gap(8).Children(func() {
		a.panelHead(c, "项目目录", fmt.Sprintf("%02d", len(projects)), nil)
		ui.SearchField(c, &a.groupQuery).Height(30).Placeholder("搜索当前组")
		ui.Row(c).Padding(4, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			ui.Text(c, groupName).FontSize(10).FontWeight(600).TextColor(colInk3)
			ui.Spacer(c)
			microLabel(c, "APPS")
		})
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Gap(1).Children(func() {
				for i, p := range projects {
					project := p
					current := project.ID == a.projectID
					a.contextProjectRow(c, itoaPad(i+1), project.Name,
						fmt.Sprintf("%s · %d 篇文档", versionOrDash(project.Version), project.DocumentCount),
						current, func() { a.openProject(project.ID) })
				}
			})
		})
		back := ui.Row(c).Height(32).Radius(2).Center().Gap(7).
			Border(1, colLine).Background(colSheet).
			Focusable().Role(ui.RoleButton).Label("返回项目组")
		if back.Hovered() {
			back.Background(colSheet2)
		}
		back.OnClick(a.backToGroups)
		back.Children(func() {
			ui.Icon(c, icons.arrowLeft).FontSize(13).TextColor(colInk3)
			ui.Text(c, "返回项目组").FontSize(11.5).TextColor(colInk2)
		})
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(5, 5).Radius(1).Background(colMoss)
			ui.Text(c, fmt.Sprintf("%d 个项目 · %d 篇文档 · 本地索引", len(projects), len(a.projectDocs()))).
				FontSize(9).TextColor(colInk4).SingleLine()
		})
	})
}

// editorContext is the document list of the project being edited.
func (a *App) editorContext(c *ui.Context) {
	owner := a.editorOwnerName()
	docs := a.projectDocs()
	if a.docForm.OwnerType != model.OwnerProject {
		docs = nil
	}
	ui.Column(c).FillHeight().Padding(14, 14, 12, 14).Gap(8).Children(func() {
		a.panelHead(c, "项目文档", fmt.Sprintf("%02d", len(docs)), nil)
		ui.SearchField(c, &a.groupQuery).Height(30).Placeholder("搜索当前项目")
		ui.Row(c).Padding(4, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			ui.Text(c, owner).FontSize(10).FontWeight(600).TextColor(colInk3)
			ui.Spacer(c)
			microLabel(c, "DOCS")
		})
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Gap(1).Children(func() {
				for i, d := range docs {
					doc := d
					current := doc.ID == a.docID
					a.contextDocRow(c, itoaPad(i+1), doc.Title,
						fmt.Sprintf("%s · %d 字", model.KindLabel(doc.Kind), doc.Words),
						current, func() { a.openDoc(doc.ID) })
				}
			})
		})
		back := ui.Row(c).Height(32).Radius(2).Center().Gap(7).
			Border(1, colLine).Background(colSheet).
			Focusable().Role(ui.RoleButton).Label("返回项目")
		if back.Hovered() {
			back.Background(colSheet2)
		}
		back.OnClick(a.backFromEditor)
		back.Children(func() {
			ui.Icon(c, icons.arrowLeft).FontSize(13).TextColor(colInk3)
			ui.Text(c, "返回项目").FontSize(11.5).TextColor(colInk2)
		})
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(5, 5).Radius(1).Background(colMoss)
			ui.Text(c, "Markdown · UTF-8 · 本地保存").
				FontSize(9).TextColor(colInk4).SingleLine()
		})
	})
}

// contextProjectRow is one numbered row of a context panel list.
func (a *App) contextProjectRow(c *ui.Context, index, title, sub string, active bool, fn func()) {
	row := ui.Row(c).MinHeight(40).Padding(6, 8).Gap(8).Radius(2).
		Focusable().Role(ui.RoleButton).Label(title)
	switch {
	case active:
		row.Background(colSealSoft).BorderWidth(0, 0, 0, 2).BorderColor(colSeal)
	case row.Hovered():
		row.Background(colSheet)
	}
	row.OnClick(fn)
	titleColor := colInk
	if active {
		titleColor = colSeal
	}
	row.Children(func() {
		monoText(c, index, 9).TextColor(colInk4).Padding(2, 0, 0, 0)
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, title).FontSize(11.5).FontWeight(600).TextColor(titleColor).SingleLine()
			ui.Text(c, sub).FontSize(9.5).TextColor(colInk3).SingleLine()
		})
	})
}

func (a *App) contextDocRow(c *ui.Context, index, title, sub string, active bool, fn func()) {
	a.contextProjectRow(c, index, title, sub, active, fn)
}

// contextGroupRow is a group entry with its identity color square.
func (a *App) contextGroupRow(c *ui.Context, icon *ui.SVG, name string, count int, active bool, fn func()) {
	row := ui.Row(c).Height(34).Padding(0, 8).Gap(8).Radius(2).
		Focusable().Role(ui.RoleButton).Label(name)
	switch {
	case active:
		row.Background(colSealSoft).BorderWidth(0, 0, 0, 2).BorderColor(colSeal)
	case row.Hovered():
		row.Background(colSheet)
	}
	row.OnClick(fn)
	nameColor := colInk
	dot := colInk4
	if active {
		nameColor = colSeal
		dot = colSeal
	}
	row.Children(func() {
		if g, ok := a.groupByName(name); ok {
			ui.Box(c).Size(7, 7).Radius(1).Background(ui.Hex(g.Color))
		} else {
			ui.Box(c).Size(7, 7).Radius(1).Background(dot)
		}
		ui.Text(c, name).FontSize(11.5).FontWeight(600).TextColor(nameColor).SingleLine()
		ui.Spacer(c)
		monoText(c, fmt.Sprintf("%02d", count), 9.5).TextColor(colInk4)
	})
}

func (a *App) contextSummaryRow(c *ui.Context, label, value string) {
	ui.Row(c).Height(20).Children(func() {
		ui.Text(c, label).FontSize(10.5).TextColor(colInk3)
		ui.Spacer(c)
		ui.Text(c, value).FontSize(10.5).TextColor(colInk2).SingleLine()
	})
}

// statusBar is the 30px strip at the foot of the main area.
func (a *App) statusBar(c *ui.Context) {
	ui.Row(c).Height(30).Padding(0, 20).Background(colSheet2).
		BorderWidth(1, 0, 0, 0).BorderColor(colLine).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(6, 6).Radius(1).Background(colMoss)
		ui.Text(c, a.statusLeft()).FontSize(9.5).TextColor(colInk3).SingleLine()
		ui.Spacer(c)
		ui.Text(c, a.statusShortcuts()).FontSize(9.5).TextColor(colInk4).SingleLine()
	})
}

func (a *App) statusLeft() string {
	switch a.page {
	case PageGroups:
		name, _, _ := a.currentGroup()
		return fmt.Sprintf("当前组：%s · %d 个项目 · %d 篇组文档", name, len(a.visibleProjects()), len(a.groupDocs()))
	case PageProject:
		return fmt.Sprintf("PROJECT %04d · %d 篇文档 · 最近更新 %s", a.projectID, len(a.projectDocs()), model.Relative(a.currentProject.UpdatedAt))
	case PageEditor:
		if a.currentDoc.FileName != "" {
			return fmt.Sprintf("附件 · %s · %s", a.currentDoc.FileName, byteSize(a.currentDoc.FileSize))
		}
		words := len([]rune(a.docContent()))
		state := "未保存"
		if !a.editorDirty() && a.docFormID != 0 {
			state = "已保存"
		}
		return fmt.Sprintf("Markdown · UTF-8 · %d 字 · %s", words, state)
	case PageStats:
		return fmt.Sprintf("数据统计 · %d 项目 · %d 组 · %d 文档 · 快照 %s", len(a.projects), len(a.groups), len(a.docs), time.Now().Format("15:04"))
	default:
		return fmt.Sprintf("SQLite 已连接 · 本地存储 %d 篇文档 · 最近同步 %s", len(a.docs), time.Now().Format("15:04"))
	}
}

func (a *App) statusShortcuts() string {
	switch a.page {
	case PageEditor:
		return "CTRL S 保存 · CTRL F 搜索"
	case PageProject:
		return "CTRL N 新建文档 · CTRL F 搜索"
	case PageGroups:
		return "CTRL N 新建项目 · CTRL F 搜索"
	case PageStats:
		return "CTRL K 命令 · CTRL N 新建"
	default:
		return "CTRL K 命令 · CTRL N 新建 · CTRL F 搜索"
	}
}

// groupOrUngrouped names the group a project belongs to.
func groupOrUngrouped(a *App, p model.Project) string {
	if p.GroupName != "" {
		return p.GroupName
	}
	return "未分组"
}

func versionOrDash(v string) string {
	if v == "" {
		return "—"
	}
	return v
}

func itoaPad(n int) string {
	return fmt.Sprintf("%02d", n)
}
