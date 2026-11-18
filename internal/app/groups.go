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

// groupsPage is the group ledger: one header, one pair of tabs, one table
// and the group documents strip at the foot. No nested cards.
func (a *App) groupsPage(c *ui.Context) {
	name, desc, color := a.currentGroup()
	projects := a.visibleProjects()
	docs := a.groupDocs()
	ui.Column(c).Fill().Children(func() {
		a.groupHead(c, name, desc, color, len(projects), len(docs))
		a.groupTabs(c, len(projects), len(docs))
		if a.groupTab == 1 && a.groupFilter > 0 {
			a.groupDocsTable(c, docs)
		} else {
			a.groupProjectsTable(c, projects)
		}
	})
}

// groupHead is the page header: identity color, group name, description,
// one meta line and the actions.
func (a *App) groupHead(c *ui.Context, name, desc, color string, projectCount, docCount int) {
	ui.Column(c).Padding(18, 24, 12, 24).Gap(4).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "索引").FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, "项目组 02").FontSize(10).FontWeight(600).TextColor(colSeal)
		})
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(11, 11).Radius(1).Background(ui.Hex(color))
			ui.Text(c, name).FontSize(22).FontWeight(700).SingleLine()
			ui.Spacer(c)
			if a.groupFilter > 0 {
				a.borderedButton(c, icons.edit, "编辑", a.openEditGroup)
				ui.Box(c).Width(7)
				a.dangerButton(c, icons.trash, "删除项目组", a.askDeleteGroup)
				ui.Box(c).Width(7)
			}
			a.primaryButton(c, icons.plus, "新建项目", a.openNewProject)
		})
		ui.Text(c, desc).FontSize(11.5).TextColor(colInk3).SingleLine()
		ui.Row(c).Gap(14).Padding(2, 0, 0, 0).Children(func() {
			a.metaItem(c, icons.grid, fmt.Sprintf("%d 个项目", projectCount))
			a.metaItem(c, icons.file, fmt.Sprintf("%d 篇文档", docCount))
			a.metaItem(c, icons.clock, groupUpdated(a.visibleProjects(), a.groupDocs()))
		})
	})
}

// groupTabs is the underline tab row with the filters on the right.
func (a *App) groupTabs(c *ui.Context, projectCount, docCount int) {
	w, _ := c.Size()
	narrow := w < 1280
	ui.Column(c).Padding(0, 24).Children(func() {
		hairline(c)
		ui.Row(c).Height(42).AlignItems(ui.Center).Children(func() {
			if a.groupFilter > 0 {
				ui.Tabs(c, &a.groupTab, fmt.Sprintf("项目 %02d", projectCount), fmt.Sprintf("文档 %02d", docCount))
				ui.Spacer(c)
				if a.groupTab == 1 {
					a.borderedButton(c, icons.file, "上传文档", func() {
						a.uploadDocument(model.OwnerGroup, a.groupFilter)
					})
					ui.Box(c).Width(8)
					a.primaryButton(c, icons.plus, "新建文档", func() {
						a.newDoc(model.OwnerGroup, a.groupFilter)
					})
					ui.Box(c).Width(10)
				}
			} else {
				sectionTitle(c, fmt.Sprintf("项目 %02d", projectCount))
				ui.Spacer(c)
			}
			ui.SearchField(c, &a.groupQuery).Width(map[bool]float32{true: 170, false: 220}[narrow]).
				Height(30).Placeholder("搜索当前组")
			if !narrow {
				ui.Box(c).Width(8)
				ui.Select(c, &a.tagFilter, a.tagOptions()).Width(120).Height(30)
				ui.Box(c).Width(8)
				ui.Select(c, &a.sortMode, []string{"更新时间", "名称"}).Width(100).Height(30)
			}
		})
		hairline(c)
	})
}

// groupProjectsTable is the project ledger of the current selection.
func (a *App) groupProjectsTable(c *ui.Context, list []model.Project) {
	if len(list) == 0 {
		title := "这个组还没有项目"
		sub := "新建一个项目，或从其他组移入已有项目。"
		if a.groupFilter == 0 {
			title = "还没有任何项目"
			sub = "从第一个项目开始建立本地索引。登记名称与路径，或先建一个项目组。"
		}
		ui.Column(c).Grow(1).Children(func() {
			a.emptyState(c, title, sub, [2]any{"新建项目", func() { a.openNewProject() }})
		})
		return
	}
	w, _ := c.Size()
	narrow := w < 1280
	a.projectTable.Selected = &a.projectRow
	a.projectTable.Key = func(row int) any { return list[row].ID }
	a.projectTable.Label = func(row int) string { return list[row].Name }
	kinds := []struct {
		title string
		width float32
		draw  func(p model.Project, row int)
	}{
		{title: "#", width: 40, draw: func(p model.Project, row int) {
			monoText(c, itoaPad(row+1), 10).TextColor(colInk4)
		}},
		{title: "项目 / 路径", draw: func(p model.Project, row int) {
			ui.Column(c).Gap(2).Children(func() {
				ui.Text(c, p.Name).FontSize(12.5).FontWeight(600).SingleLine()
				monoText(c, pathOrNone(p.Path), 9.5).TextColor(colInk3).SingleLine()
			})
		}},
		{title: "版本", width: 76, draw: func(p model.Project, row int) {
			a.versionChip(c, p.Version)
		}},
		{title: "标签", width: 140, draw: func(p model.Project, row int) {
			tagsText(c, p.Tags, 2)
		}},
		{title: "更新", width: 76, draw: func(p model.Project, row int) {
			ui.Text(c, model.Relative(p.UpdatedAt)).FontSize(10.5).TextColor(colInk3)
		}},
		{title: "", width: 30, draw: func(p model.Project, row int) {
			more := ui.Box(c).Size(24, 24).Radius(2).Center().
				Focusable().Role(ui.RoleButton).Label("移入组").Tooltip("移入组")
			if more.Hovered() {
				more.Background(colSheet2)
			}
			more.OnClick(func() { a.openMoveProject(p.ID) })
			more.Children(func() {
				ui.Icon(c, icons.more).FontSize(14).TextColor(colInk4)
			})
		}},
	}
	if narrow {
		kinds = append(kinds[:3:3], kinds[4]) // drop the tags column
	}
	cols := make([]ui.TableColumn, len(kinds))
	for i, k := range kinds {
		cols[i] = ui.TableColumn{Title: k.title, Width: k.width}
	}
	tbl := ui.Table(c, &a.projectTable, cols, len(list), func(row, col int) {
		p := list[row]
		r := ui.Row(c).MinHeight(56).AlignItems(ui.Center).
			Focusable().Role(ui.RoleButton).Label(p.Name)
		r.OnClick(func() { a.openProject(p.ID) })
		r.Children(func() { kinds[col].draw(p, row) })
	}).FillWidth().Padding(0, 24).Grow(1)
	if tbl.Changed() && a.projectRow >= 0 && a.projectRow < len(list) {
		id := list[a.projectRow].ID
		a.projectRow = -1
		a.openProject(id)
	}
	a.groupFoot(c, fmt.Sprintf("共 %d 个项目 · 点击行进入项目详情", len(list)), "按更新时间排序")
	a.groupDocsStrip(c)
}

// groupDocsTable is the documents ledger of a group.
func (a *App) groupDocsTable(c *ui.Context, list []model.Document) {
	if len(list) == 0 {
		ui.Column(c).Grow(1).Children(func() {
			a.emptyState(c, "这个组还没有组文档", "组文档适合存放组规范、发布流程与评审记录。",
				[2]any{"新建文档", func() { a.newDoc(model.OwnerGroup, a.groupFilter) }})
		})
		return
	}
	a.groupDocTable.Selected = &a.docRow
	a.groupDocTable.Key = func(row int) any { return list[row].ID }
	a.groupDocTable.Label = func(row int) string { return list[row].Title }
	cols := []ui.TableColumn{
		{Title: "#", Width: 40},
		{Title: "文档"},
		{Title: "类型", Width: 110},
		{Title: "字数", Width: 76},
		{Title: "更新", Width: 96},
		{Title: "", Width: 30},
	}
	tbl := ui.Table(c, &a.groupDocTable, cols, len(list), func(row, col int) {
		d := list[row]
		switch col {
		case 0:
			monoText(c, itoaPad(row+1), 10).TextColor(colInk4)
		case 1:
			r := ui.Row(c).MinHeight(48).AlignItems(ui.Center).Gap(8).
				Focusable().Role(ui.RoleButton).Label(d.Title)
			r.OnClick(func() { a.openDoc(d.ID) })
			r.Children(func() {
				ui.Text(c, d.Title).FontSize(12.5).FontWeight(600).SingleLine()
			})
		case 2:
			r := ui.Row(c).MinHeight(48).AlignItems(ui.Center).
				Focusable().Role(ui.RoleButton).Label(model.KindLabel(d.Kind))
			r.OnClick(func() { a.openDoc(d.ID) })
			r.Children(func() {
				if d.FileName != "" {
					monoText(c, fileKindLabel(d), 10).TextColor(colInk3)
				} else {
					a.kindTag(c, d.Kind)
				}
			})
		case 3:
			if d.FileName != "" {
				monoText(c, byteSize(d.FileSize), 10.5).TextColor(colInk3)
			} else {
				monoText(c, itoa(d.Words), 10.5).TextColor(colInk3)
			}
		case 4:
			r := ui.Row(c).MinHeight(48).AlignItems(ui.Center).
				Focusable().Role(ui.RoleButton).Label(model.Relative(d.UpdatedAt))
			r.OnClick(func() { a.openDoc(d.ID) })
			r.Children(func() {
				ui.Text(c, model.Relative(d.UpdatedAt)).FontSize(10.5).TextColor(colInk3)
			})
		case 5:
			doc := d
			ui.Box(c).Size(24, 24).Radius(2).Center().Focusable().Role(ui.RoleButton).Label("文档操作").OnClick(func() {
				a.askDeleteDoc(doc.ID)
			}).Children(func() {
				ui.Icon(c, icons.more).FontSize(14).TextColor(colInk4)
			})
		}
	}).FillWidth().Padding(0, 24).Grow(1)
	if tbl.Changed() && a.docRow >= 0 && a.docRow < len(list) {
		id := list[a.docRow].ID
		a.docRow = -1
		a.openDoc(id)
	}
	a.groupFoot(c, fmt.Sprintf("共 %d 篇文档 · 组文档独立于项目文档", len(list)), "按更新时间排序")
}

// groupFoot is the ledger foot line under a table.
func (a *App) groupFoot(c *ui.Context, left, right string) {
	ui.Row(c).Height(30).Padding(0, 24).BorderWidth(1, 0, 0, 0).BorderColor(colLine2).
		AlignItems(ui.Center).Children(func() {
		ui.Text(c, left).FontSize(9.5).TextColor(colInk4)
		ui.Spacer(c)
		ui.Text(c, right).FontSize(9.5).TextColor(colInk4)
	})
}

// groupDocsStrip is the fixed strip of group documents at the foot of the
// projects tab.
func (a *App) groupDocsStrip(c *ui.Context) {
	if a.groupFilter <= 0 {
		return
	}
	docs := a.groupDocs()
	ui.Column(c).Background(colSheet2).BorderWidth(1, 0, 0, 0).BorderColor(colLine).
		Children(func() {
			ui.Row(c).Height(32).Padding(0, 24).AlignItems(ui.Center).Children(func() {
				sectionTitle(c, "组文档")
				ui.Box(c).Width(6)
				monoText(c, itoaPad(len(docs)), 9.5).TextColor(colInk4)
				ui.Spacer(c)
				a.linkMore(c, "查看全部", func() { a.groupTab = 1 })
			})
			if len(docs) == 0 {
				ui.Row(c).Height(28).Padding(0, 24).Children(func() {
					ui.Text(c, "还没有组文档；组规范与发布流程适合放在这里。").FontSize(10).TextColor(colInk4)
				})
				return
			}
			if len(docs) > 3 {
				docs = docs[:3]
			}
			for _, d := range docs {
				doc := d
				row := ui.Row(c).Height(30).Padding(0, 24).Gap(8).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label(doc.Title)
				if row.Hovered() {
					row.Background(colSheet)
				}
				row.OnClick(func() { a.openDoc(doc.ID) })
				row.Children(func() {
					kindDot(c, doc.Kind)
					ui.Text(c, doc.Title).FontSize(11).TextColor(colInk2).SingleLine()
					ui.Spacer(c)
					ui.Text(c, model.KindLabel(doc.Kind)).FontSize(9.5).TextColor(colInk3)
					ui.Box(c).Width(16)
					ui.Text(c, model.Relative(doc.UpdatedAt)).FontSize(9.5).TextColor(colInk4)
				})
			}
		})
}

func pathOrNone(path string) string {
	if path == "" {
		return "未登记路径"
	}
	return path
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
