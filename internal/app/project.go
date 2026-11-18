package app

import (
	"fmt"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

func (a *App) projectPage(c *ui.Context) {
	p := a.currentProject
	w, _ := c.Size()
	infoWidth := float32(320)
	if w < 1280 {
		infoWidth = 288
	}
	ui.Column(c).Fill().Padding(16, 20, 18, 20).Gap(10).Children(func() {
		a.detailCrumb(c)
		a.detailHead(c, p)
		ui.Row(c).Grow(1).Gap(14).AlignItems(ui.Stretch).Children(func() {
			left := a.card(c, 0, func() { a.projectDocsCard(c, p) })
			left.Grow(1)
			right := a.card(c, 0, func() { a.projectInfoCard(c, p) })
			right.Width(infoWidth)
		})
	})
}

func (a *App) detailCrumb(c *ui.Context) {
	ui.Row(c).Height(22).Gap(6).Children(func() {
		b := ui.Box(c).Size(20, 20).Radius(4).Center().Focusable().Role(ui.RoleButton).Label("返回项目组")
		b.OnClick(a.backToGroups)
		b.Children(func() { ui.Icon(c, icons.arrowLeft).FontSize(12).TextColor(colText2) })
		ui.Text(c, "项目组管理").FontSize(11).TextColor(colText3)
		ui.Text(c, "/").FontSize(11.5).TextColor(colText4)
		ui.Text(c, a.currentProject.GroupName).FontSize(11.5).TextColor(colText3)
		ui.Text(c, "/").FontSize(11.5).TextColor(colText4)
		ui.Text(c, a.currentProject.Name).FontSize(11).Bold()
	})
}

func (a *App) detailHead(c *ui.Context, p model.Project) {
	ui.Row(c).Height(76).Gap(12).Children(func() {
		ui.Box(c).Size(46, 46).Radius(8).Background(colPrimary1).Center().Children(func() {
			ui.Icon(c, icons.code).FontSize(22).TextColor(colPrimary)
		})
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Row(c).Gap(9).Children(func() {
				ui.Text(c, p.Name).FontSize(20).Bold()
				a.versionChip(c, p.Version)
			})
			ui.Text(c, p.Description).FontSize(11.5).TextColor(colText3).SingleLine()
			ui.Row(c).Gap(5).Children(func() {
				for i, tag := range p.Tags {
					if i > 4 {
						break
					}
					fg, bg := colPrimary, colPrimary1
					if i == 1 {
						fg, bg = colCyan, colCyan1
					}
					a.tag(c, tag, fg, bg)
				}
			})
		})
		ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
			a.secondaryButton(c, icons.folderOpen, "打开目录", func() {
				a.openProjectFolder(c, p)
			})
			a.secondaryButton(c, icons.edit, "编辑项目", a.openEditProject)
			a.ghostButton(c, icons.more, "更多", a.askDeleteProject)
		})
	})
}

func (a *App) openProjectFolder(c *ui.Context, p model.Project) {
	if p.Path == "" {
		c.Toast("尚未登记项目路径")
		return
	}
	if _, err := os.Stat(p.Path); err != nil {
		c.Toast("项目路径不存在，无法打开")
		return
	}
	go func() {
		if err := mygo.Shell.OpenPath(p.Path); err != nil && a.win != nil {
			a.win.Update(func() { a.notice = "打开目录失败：" + err.Error() })
		}
	}()
}

func (a *App) projectDocsCard(c *ui.Context, p model.Project) {
	docs := a.projectDocs()
	ui.Row(c).Height(46).Padding(0, 16).Children(func() {
		ui.Text(c, "项目文档").FontSize(13).Bold()
		ui.Text(c, itoa(len(docs))).FontSize(10.5).TextColor(colText3)
		ui.Spacer(c)
		a.linkMore(c, "管理排序", func() {})
	})
	ui.Row(c).Height(40).Padding(0, 14).Gap(8).BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Children(func() {
		ui.SearchField(c, &a.groupQuery).Width(210).Height(28).Placeholder("搜索文档标题或正文…")
		ui.Spacer(c)
		a.primaryButton(c, icons.plus, "新建文档", func() {
			a.newDoc(model.OwnerProject, p.ID)
		})
	})
	if len(docs) == 0 {
		a.emptyState(c, icons.file, "这个项目还没有文档", "新建需求、设计或开发计划，文档会跟随项目。", "新建文档", func() {
			a.newDoc(model.OwnerProject, p.ID)
		})
		return
	}
	a.detailDocTable.Selected = &a.docRow
	a.detailDocTable.Key = func(row int) any { return docs[row].ID }
	a.detailDocTable.Label = func(row int) string { return docs[row].Title }
	cols := []ui.TableColumn{
		{Title: "标题"},
		{Title: "类型", Width: 100},
		{Title: "字数", Width: 78},
		{Title: "更新时间", Width: 92},
		{Title: "", Width: 26},
	}
	tbl := ui.Table(c, &a.detailDocTable, cols, len(docs), func(row, col int) {
		d := docs[row]
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
	if tbl.Changed() && a.detailDocTable.Selected != nil && *a.detailDocTable.Selected >= 0 && *a.detailDocTable.Selected < len(docs) {
		a.openDoc(docs[*a.detailDocTable.Selected].ID)
	}
	words := 0
	for _, d := range docs {
		words += d.Words
	}
	ui.Row(c).Height(30).Padding(0, 16).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, fmt.Sprintf("共 %d 篇 · 合计 %d 字", len(docs), words)).FontSize(10).TextColor(colText3)
		ui.Spacer(c)
		ui.Text(c, "最后更新于 "+model.Relative(docs[0].UpdatedAt)).FontSize(10.5).TextColor(colText3)
	})
}

func (a *App) projectInfoCard(c *ui.Context, p model.Project) {
	ui.Row(c).Height(42).Padding(0, 16).BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, "项目信息").FontSize(13).Bold()
		ui.Spacer(c)
		a.ghostButton(c, icons.edit, "编辑项目", a.openEditProject)
	})
	ui.Scroll(c).Grow(1).Padding(12, 14).Children(func() {
		a.sectionLabel(c, "基本信息")
		a.infoRow(c, "所属项目组", func() {
			ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(6, 6).Radius(3).Background(groupColor(a, p.GroupID))
				ui.Text(c, p.GroupName).FontSize(11.5).Bold()
			})
		})
		a.infoRow(c, "版本", func() { a.versionChip(c, p.Version) })
		a.infoRow(c, "项目路径", func() {
			path := p.Path
			if path == "" {
				path = "未登记"
			}
			box := ui.Row(c).Gap(5).Padding(5, 7).Radius(6).Background(colSurface2).Border(1, colBorder2)
			box.Children(func() {
				ui.Text(c, path).FontSize(9.6).TextColor(colText2).SingleLine()
				ui.Spacer(c)
				b := ui.Box(c).Size(20, 20).Radius(4).Center().Focusable().Role(ui.RoleButton).Label("复制路径")
				b.OnClick(func() { c.WriteClipboard(path); c.Toast("路径已复制到剪贴板") })
				b.Children(func() { ui.Icon(c, icons.copy).FontSize(12).TextColor(colText3) })
			})
		})
		a.infoRow(c, "创建时间", func() { ui.Text(c, model.DateTime(p.CreatedAt)).FontSize(11).TextColor(colText2) })
		a.infoRow(c, "更新时间", func() { ui.Text(c, model.Relative(p.UpdatedAt)).FontSize(11).TextColor(colText2) })
		ui.Box(c).Height(1).Background(colBorder2).Margin(12, 0, 12, 0)
		a.sectionLabel(c, "标签")
		ui.Row(c).Gap(5).Padding(2, 0, 8, 0).Children(func() {
			if len(p.Tags) == 0 {
				ui.Text(c, "暂无标签").FontSize(11).TextColor(colText3)
			}
			for i, tag := range p.Tags {
				fg, bg := colPrimary, colPrimary1
				if i == 1 {
					fg, bg = colCyan, colCyan1
				}
				a.tag(c, tag, fg, bg)
			}
		})
		ui.Box(c).Height(1).Background(colBorder2).Margin(12, 0, 12, 0)
		a.sectionLabel(c, "备注")
		note := p.Description
		if note == "" {
			note = "暂无备注"
		}
		ui.Box(c).Background(colSurface2).Border(1, colBorder2).Radius(4).Padding(8, 9).Children(func() {
			ui.Text(c, note).FontSize(11).TextColor(colText2)
		})
		ui.Row(c).Gap(10).Padding(12, 0, 0, 0).Children(func() {
			a.infoStat(c, itoa(p.DocumentCount), "项目文档")
			words := 0
			for _, d := range a.projectDocs() {
				words += d.Words
			}
			a.infoStat(c, fmt.Sprintf("%.1fk", float64(words)/1000), "文档总字数")
		})
	})
}

func (a *App) infoRow(c *ui.Context, label string, value func()) {
	ui.Row(c).Padding(6, 0).Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Box(c).Width(72).Children(func() {
			ui.Text(c, label).FontSize(11).TextColor(colText3)
		})
		ui.Column(c).Grow(1).Children(value)
	})
}

func (a *App) infoStat(c *ui.Context, value, label string) {
	ui.Box(c).Grow(1).Background(colSurface2).Border(1, colBorder2).Radius(8).Padding(10, 11).Children(func() {
		ui.Text(c, value).FontSize(18).Bold()
		ui.Text(c, label).FontSize(10).TextColor(colText3).Padding(2, 0, 0, 0)
	})
}

func groupColor(a *App, groupID *int64) ui.Color {
	if groupID == nil {
		return colText4
	}
	if g, ok := a.groupByID(*groupID); ok {
		return ui.Hex(g.Color)
	}
	return colPrimary
}
