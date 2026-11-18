package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(strings.TrimSpace(sub)))
}

func (a *App) groupsPage(c *ui.Context) {
	ui.Box(c).Fill().Padding(16, 18).Children(func() {
		ui.Row(c).Fill().Gap(14).AlignItems(ui.Stretch).Children(func() {
			ui.Box(c).Width(280).FillHeight().Children(func() {
				a.groupPanel(c)
			})
			ui.Box(c).Grow(1).FillHeight().Children(func() {
				a.groupMain(c)
			})
		})
	})
}

func (a *App) groupPanel(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).FillHeight().Background(t.Surface).Border(1, colBorder).Radius(10).
		Children(func() {
			ui.Row(c).Height(46).Padding(0, 14).BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Children(func() {
				ui.Row(c).Gap(7).Children(func() {
					ui.Text(c, "项目组").FontSize(13).Bold()
					ui.Text(c, itoa(len(a.groups))).FontSize(10.5).TextColor(colText3)
				})
				ui.Spacer(c)
				a.ghostButton(c, icons.plus, "新建项目组", a.openNewGroup)
			})
			ui.Box(c).Padding(10, 12, 4, 12).Children(func() {
				ui.SearchField(c, &a.groupListQuery).Height(30).Placeholder("搜索项目组…")
			})
			ui.Scroll(c).Grow(1).Padding(4, 10, 8, 10).Children(func() {
				a.groupNavRow(c, "全部项目", icons.layers, len(a.projects), a.groupFilter == 0, func() { a.openGroup(0) })
				a.sectionLabel(c, "项目组")
				for _, g := range a.groups {
					group := g
					if a.groupListQuery != "" && !containsFold(group.Name, a.groupListQuery) {
						continue
					}
					a.groupNavRow(c, group.Name, icons.folder, group.ProjectCount, a.groupFilter == group.ID,
						func() { a.openGroup(group.ID) })
				}
				ui.Box(c).Height(1).Background(colBorder2).BorderStyle(ui.BorderDashed).Margin(9, 8, 7, 8)
				ungrouped := 0
				for _, p := range a.projects {
					if p.GroupID == nil {
						ungrouped++
					}
				}
				a.groupNavRow(c, "未分组", icons.folderOpen, ungrouped, a.groupFilter == -1, func() { a.openGroup(-1) })
			})
			ui.Column(c).Padding(8, 12, 12, 12).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
				button := ui.Row(c).Height(30).Radius(4).Center().
					Border(1, colBorder).BorderStyle(ui.BorderDashed).
					Focusable().Role(ui.RoleButton).Label("新建项目组")
				button.OnClick(a.openNewGroup)
				button.Children(func() {
					ui.Icon(c, icons.plus).FontSize(13).TextColor(colText3)
					ui.Text(c, "新建项目组").FontSize(11.5).TextColor(colText3)
				})
			})
		})
}

func (a *App) groupNavRow(c *ui.Context, name string, icon *ui.SVG, count int, active bool, fn func()) {
	row := ui.Row(c).Height(38).Padding(0, 8).Gap(8).Radius(4).
		Focusable().Role(ui.RoleButton).Label(name)
	if active {
		row.Background(colPrimary1)
	} else if row.Hovered() {
		row.Background(ui.Hex("#F7F9FC"))
	}
	row.OnClick(fn)
	fg := colText2
	if active {
		fg = colPrimary
	}
	row.Children(func() {
		ui.Box(c).Size(22, 22).Radius(4).Background(map[bool]ui.Color{true: colSurface, false: ui.Hex("#F2F4F7")}[active]).Center().Children(func() {
			ui.Icon(c, icon).FontSize(13).TextColor(fg)
		})
		ui.Text(c, name).FontSize(12).TextColor(fg)
		ui.Spacer(c)
		badgeBg := ui.Hex("#F2F4F7")
		badgeFg := colText3
		if active {
			badgeBg = colSurface
			badgeFg = colPrimary
		}
		ui.Row(c).Height(18).Padding(0, 5).Radius(9).Background(badgeBg).Center().Children(func() {
			ui.Text(c, itoa(count)).FontSize(9.5).TextColor(badgeFg)
		})
	})
}

func (a *App) groupMain(c *ui.Context) {
	t := c.Theme()
	name, desc, color := a.currentGroup()
	projects := a.visibleProjects()
	docs := a.groupDocs()
	ui.Column(c).FillHeight().Background(t.Surface).Border(1, colBorder).Radius(10).
		Children(func() {
			a.groupHead(c, name, desc, color, len(projects), len(docs))
			a.groupTabs(c, len(projects), len(docs))
			if a.groupTab == 1 && a.groupFilter > 0 {
				a.groupDocsTable(c, docs)
			} else {
				a.groupProjectsTable(c, projects)
			}
		})
}

func (a *App) groupHead(c *ui.Context, name, desc, color string, projectCount, docCount int) {
	w, _ := c.Size()
	narrow := w < 1280
	ui.Row(c).Padding(14, 18, 12, 18).BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Gap(14).Children(func() {
		ui.Box(c).Size(42, 42).Radius(8).Background(ui.Hex(color).Alpha(0.10)).Center().Children(func() {
			ui.Icon(c, icons.folder).FontSize(20).TextColor(ui.Hex(color))
		})
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Text(c, name).FontSize(17.5).Bold().SingleLine()
			ui.Text(c, desc).FontSize(11.5).TextColor(colText3).SingleLine()
			ui.Row(c).Gap(10).Children(func() {
				a.metaItem(c, icons.grid, fmt.Sprintf("%d 个项目", projectCount))
				a.metaItem(c, icons.file, fmt.Sprintf("%d 篇文档", docCount))
				if !narrow {
					a.metaItem(c, icons.clock, groupUpdated(a.visibleProjects(), a.groupDocs()))
				}
			})
		})
		ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
			if a.groupFilter > 0 {
				if narrow {
					a.ghostButton(c, icons.edit, "编辑", a.openEditGroup)
				} else {
					a.secondaryButton(c, icons.edit, "编辑", a.openEditGroup)
				}
				a.ghostButton(c, icons.more, "更多", func() {
					a.askDeleteGroup()
				})
			}
			a.primaryButton(c, icons.plus, "新建项目", a.openNewProject)
		})
	})
}

func (a *App) metaItem(c *ui.Context, icon *ui.SVG, text string) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, icon).FontSize(11).TextColor(colText3)
		ui.Text(c, text).FontSize(10.5).TextColor(colText3)
	})
}

func (a *App) groupTabs(c *ui.Context, projectCount, docCount int) {
	w, _ := c.Size()
	narrow := w < 1280
	ui.Row(c).Height(40).Padding(0, 14).BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Children(func() {
		if a.groupFilter > 0 {
			ui.Tabs(c, &a.groupTab, fmt.Sprintf("项目 %d", projectCount), fmt.Sprintf("文档 %d", docCount))
			if a.groupTab == 1 {
				a.primaryButton(c, icons.plus, "新建文档", func() {
					a.newDoc(model.OwnerGroup, a.groupFilter)
				})
			}
		} else {
			ui.Text(c, fmt.Sprintf("项目 %d", projectCount)).FontSize(12.5).Bold().TextColor(colPrimary)
		}
		ui.Spacer(c)
		if !narrow {
			ui.Select(c, &a.tagFilter, a.tagOptions()).Width(120).Height(28)
			ui.Box(c).Width(7)
		}
		searchWidth := float32(210)
		if narrow {
			searchWidth = 160
		}
		ui.SearchField(c, &a.groupQuery).Width(searchWidth).Height(28).Placeholder("搜索当前组…")
		if !narrow {
			ui.Box(c).Width(7)
			ui.Select(c, &a.sortMode, []string{"更新时间", "名称"}).Width(90).Height(28)
		}
	})
}

func (a *App) groupProjectsTable(c *ui.Context, list []model.Project) {
	if len(list) == 0 {
		title := "这个组还没有项目"
		sub := "新建一个项目，或从其他组移入已有项目。"
		if a.groupFilter == 0 {
			title = "还没有任何项目"
			sub = "点击右上角「新建项目」，开始登记第一个项目。"
		}
		a.emptyState(c, icons.folderOpen, title, sub, "新建项目", a.openNewProject)
		return
	}
	a.projectTable.Selected = &a.projectRow
	a.projectTable.Key = func(row int) any { return list[row].ID }
	a.projectTable.Label = func(row int) string { return list[row].Name }
	cols := []ui.TableColumn{
		{Title: "项目"},
		{Title: "版本", Width: 72},
		{Title: "标签", Width: 150},
		{Title: "更新时间", Width: 88},
		{Title: "", Width: 26},
	}
	tbl := ui.Table(c, &a.projectTable, cols, len(list), func(row, col int) {
		p := list[row]
		switch col {
		case 0:
			a.clickRow(c, 48, func() { a.openProject(p.ID) }, func() {
				ui.Row(c).Gap(10).Children(func() {
					icon, fg, bg := projectIcon(p)
					a.iconTile(c, icon, 28, fg, bg)
					ui.Column(c).Gap(1).Children(func() {
						ui.Text(c, p.Name).FontSize(12.5).Bold().SingleLine()
						ui.Text(c, p.Path).FontSize(9.5).TextColor(colText4).SingleLine()
					})
				})
			})
		case 1:
			a.clickRow(c, 48, func() { a.openProject(p.ID) }, func() { a.versionChip(c, p.Version) })
		case 2:
			a.clickRow(c, 48, func() { a.openProject(p.ID) }, func() { a.tagsCell(c, p.Tags) })
		case 3:
			a.clickRow(c, 48, func() { a.openProject(p.ID) }, func() {
				ui.Text(c, model.Relative(p.UpdatedAt)).FontSize(10.5).TextColor(colText3)
			})
		case 4:
			ui.Icon(c, icons.more).FontSize(15).TextColor(colText3)
		}
	}).FillWidth().Padding(4, 12).Grow(1)
	if tbl.Changed() && a.projectRow >= 0 && a.projectRow < len(list) {
		id := list[a.projectRow].ID
		a.projectRow = -1
		a.openProject(id)
	}
	ui.Row(c).Height(30).Padding(0, 18).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, fmt.Sprintf("共 %d 个项目", len(list))).FontSize(10).TextColor(colText3)
		ui.Spacer(c)
		ui.Text(c, "按更新时间排序").FontSize(10.5).TextColor(colText3)
	})
	a.groupDocsPreview(c)
}

func (a *App) groupDocsPreview(c *ui.Context) {
	if a.groupFilter <= 0 {
		return
	}
	docs := a.groupDocs()
	if len(docs) == 0 {
		return
	}
	ui.Row(c).Height(30).Padding(0, 16).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Row(c).Gap(6).Children(func() {
			ui.Text(c, "组文档").FontSize(12).Bold()
			ui.Text(c, itoa(len(docs))).FontSize(9.5).TextColor(colText3)
		})
		ui.Spacer(c)
		a.linkMore(c, "查看全部", func() { a.groupTab = 1 })
	})
	if len(docs) > 3 {
		docs = docs[:3]
	}
	for _, d := range docs {
		doc := d
		row := ui.Row(c).Height(32).Padding(0, 16).Gap(8).
			Focusable().Role(ui.RoleButton).Label(doc.Title)
		if row.Hovered() {
			row.Background(ui.Hex("#F8F6FF"))
		}
		row.OnClick(func() { a.openDoc(doc.ID) })
		row.Children(func() {
			ui.Icon(c, icons.note).FontSize(13).TextColor(colText3)
			ui.Text(c, doc.Title).FontSize(11.5).TextColor(colText2).SingleLine()
			a.kindTag(c, doc.Kind)
			ui.Spacer(c)
			ui.Text(c, model.Relative(doc.UpdatedAt)).FontSize(10).TextColor(colText4)
		})
	}
}

func (a *App) groupDocsTable(c *ui.Context, list []model.Document) {
	if len(list) == 0 {
		a.emptyState(c, icons.file, "这个组还没有组文档", "组文档适合存放组规范、发布流程与评审记录。", "新建文档", func() {
			a.newDoc(model.OwnerGroup, a.groupFilter)
		})
		return
	}
	a.groupDocTable.Selected = &a.docRow
	a.groupDocTable.Key = func(row int) any { return list[row].ID }
	a.groupDocTable.Label = func(row int) string { return list[row].Title }
	cols := []ui.TableColumn{
		{Title: "文档"},
		{Title: "类型", Width: 110},
		{Title: "字数", Width: 84},
		{Title: "更新时间", Width: 104},
		{Title: "", Width: 26},
	}
	tbl := ui.Table(c, &a.groupDocTable, cols, len(list), func(row, col int) {
		d := list[row]
		switch col {
		case 0:
			a.clickRow(c, 46, func() { a.openDoc(d.ID) }, func() {
				ui.Row(c).Gap(8).Children(func() {
					ui.Icon(c, icons.note).FontSize(15).TextColor(colText3)
					ui.Text(c, d.Title).FontSize(12).Bold().SingleLine()
				})
			})
		case 1:
			a.clickRow(c, 46, func() { a.openDoc(d.ID) }, func() { a.kindTag(c, d.Kind) })
		case 2:
			a.clickRow(c, 46, func() { a.openDoc(d.ID) }, func() {
				ui.Text(c, fmt.Sprintf("%d 字", d.Words)).FontSize(10.5).TextColor(colText3)
			})
		case 3:
			a.clickRow(c, 46, func() { a.openDoc(d.ID) }, func() {
				ui.Text(c, model.Relative(d.UpdatedAt)).FontSize(10.5).TextColor(colText3)
			})
		case 4:
			doc := d
			ui.Box(c).Size(24, 24).Radius(5).Center().Focusable().Role(ui.RoleButton).Label("文档操作").OnClick(func() {
				a.askDeleteDoc(doc.ID)
			}).Children(func() {
				ui.Icon(c, icons.more).FontSize(15).TextColor(colText3)
			})
		}
	}).FillWidth().Padding(4, 12).Grow(1)
	if tbl.Changed() && a.docRow >= 0 && a.docRow < len(list) {
		id := list[a.docRow].ID
		a.docRow = -1
		a.openDoc(id)
	}
	ui.Row(c).Height(30).Padding(0, 18).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, fmt.Sprintf("共 %d 篇文档", len(list))).FontSize(10).TextColor(colText3)
		ui.Spacer(c)
		ui.Text(c, "组内文档独立于项目文档").FontSize(10.5).TextColor(colText3)
	})
}

func groupUpdated(projects []model.Project, docs []model.Document) string {
	var latest time.Time
	for _, p := range projects {
		if p.UpdatedAt.After(latest) {
			latest = p.UpdatedAt
		}
	}
	for _, d := range docs {
		if d.UpdatedAt.After(latest) {
			latest = d.UpdatedAt
		}
	}
	if latest.IsZero() {
		return "暂无更新"
	}
	return model.Relative(latest) + "更新"
}
