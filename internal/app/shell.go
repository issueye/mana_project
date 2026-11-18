package app

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

func (a *App) windowBar(c *ui.Context) {
	ui.Row(c).Height(34).Background(colSurface).
		BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).
		Children(func() {
			ui.Row(c).Grow(1).FillHeight().DragWindow().Padding(0, 12).Gap(6).Children(func() {
				ui.Box(c).Size(16, 16).Radius(4).
					Gradient(ui.Hex("#1677FF"), ui.Hex("#6A5CFF"), 140).Center().
					Children(func() {
						ui.Icon(c, icons.layers).FontSize(9).TextColor(colSurface)
					})
				ui.Text(c, "Mana").FontSize(11.5).Bold()
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

func (a *App) windowControl(c *ui.Context, icon *ui.SVG, label string, danger bool, fn func()) {
	b := ui.Box(c).Size(32, 24).Radius(4).Center().
		Focusable().Role(ui.RoleButton).Label(label)
	hover := ui.Hex("#F3F5F8")
	fg := colText3
	if danger {
		hover = colDanger1
		fg = colDanger
	}
	if b.Hovered() {
		b.Background(hover)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(11).TextColor(fg)
	})
}

func (a *App) sidebar(c *ui.Context) {
	ui.Column(c).
		Width(228).
		FillHeight().
		Background(colSurface).
		BorderWidth(0, 0, 0, 1).
		BorderColor(colBorder2).
		Padding(14, 12, 12, 12).
		Children(func() {
			a.brand(c)
			a.sectionLabel(c, "工作台")
			a.navItem(c, icons.home, "首页", 0, a.page == PageHome, a.goHome)
			count := len(a.groups)
			a.navItem(c, icons.folder, "项目组管理", count, a.page != PageHome, a.goGroups)

			ui.Box(c).Height(1).Background(colBorder2).Margin(12, 5, 12, 5)
			a.sectionLabel(c, "最近访问")
			for _, p := range a.recentProjects(3) {
				project := p
				row := ui.Row(c).Height(32).Padding(0, 8).Gap(7).Radius(4).
					Focusable().Role(ui.RoleButton).Label(project.Name)
				if row.Hovered() {
					row.Background(ui.Hex("#F7F9FC"))
				}
				row.OnClick(func() { a.openProject(project.ID) })
				row.Children(func() {
					ui.Icon(c, icons.code).FontSize(13).TextColor(colText3)
					ui.Text(c, project.Name).FontSize(11.5).TextColor(colText2).SingleLine()
				})
			}

			ui.Spacer(c)
			a.storageCard(c)
			a.userRow(c)
		})
}

func (a *App) brand(c *ui.Context) {
	ui.Row(c).Padding(2, 6, 14, 6).Gap(9).Children(func() {
		ui.Box(c).Size(30, 30).Radius(8).
			Gradient(ui.Hex("#1677FF"), ui.Hex("#6A5CFF"), 140).
			Center().
			Children(func() {
				ui.Icon(c, icons.layers).FontSize(16).TextColor(colSurface)
			})
		ui.Column(c).Gap(1).Children(func() {
			ui.Text(c, "Mana").FontSize(15).Bold()
			ui.Text(c, "开发项目管理").FontSize(10).TextColor(colText3)
		})
	})
}

func (a *App) storageCard(c *ui.Context) {
	a.flatCard(c, 10, func() {
		ui.Row(c).Gap(7).Children(func() {
			ui.Icon(c, icons.database).FontSize(14).TextColor(colPrimary)
			ui.Text(c, "本地数据库").FontSize(11.5).Bold().TextColor(colText2)
		})
		ui.Row(c).Gap(6).Padding(6, 0, 0, 0).Children(func() {
			ui.Box(c).Size(6, 6).Radius(3).Background(colSuccess)
			ui.Text(c, "SQLite · 已连接").FontSize(10).TextColor(colText3)
		})
		ui.Row(c).Padding(8, 0, 0, 0).Children(func() {
			ui.Text(c, "本地存储").FontSize(9.5).TextColor(colText4)
			ui.Spacer(c)
			ui.Text(c, fmt.Sprintf("%d 篇文档", len(a.docs))).FontSize(9.5).TextColor(colText4)
		})
	})
}

func (a *App) userRow(c *ui.Context) {
	ui.Row(c).Padding(10, 4, 2, 4).Gap(8).Children(func() {
		ui.Avatar(c, "Lin", nil).Size(26, 26)
		ui.Column(c).Gap(1).Children(func() {
			ui.Text(c, "Lin").FontSize(11.5).Bold()
			ui.Text(c, "本地工作区").FontSize(9.5).TextColor(colText3)
		})
		ui.Spacer(c)
		a.ghostButton(c, icons.settings, "设置", func() {
			c.Toast("设置面板将在后续版本开放")
		})
	})
}

func (a *App) topbar(c *ui.Context) {
	ui.Row(c).
		Height(56).
		Padding(0, 20).
		Background(colSurface).
		BorderWidth(0, 0, 1, 0).
		BorderColor(colBorder2).
		Children(func() {
			ui.Box(c).Width(280).Children(func() {
				switch a.page {
				case PageHome:
					ui.Text(c, "首页").FontSize(14).Bold()
				case PageGroups:
					ui.Text(c, "项目组管理").FontSize(14).Bold()
				case PageProject:
					groupName := a.currentProject.GroupName
					if groupName == "" {
						groupName = "未分组"
					}
					ui.Row(c).Gap(6).Children(func() {
						ui.Text(c, "项目组管理").FontSize(11.5).TextColor(colText3)
						ui.Text(c, "/").FontSize(12).TextColor(colText4)
						ui.Text(c, groupName).FontSize(11.5).TextColor(colText3)
						ui.Text(c, "/").FontSize(12).TextColor(colText4)
						ui.Text(c, a.currentProject.Name).FontSize(11.5).Bold().SingleLine()
					})
				case PageEditor:
					ui.Text(c, "文档编辑").FontSize(14).Bold()
				}
			})
			ui.Spacer(c)
			search := ui.SearchField(c, &a.searchQuery).
				Width(300).
				Bind(&a.searchHandle).
				Placeholder("搜索项目、文档或标签…")
			if search.Submitted() && a.searchQuery != "" {
				a.groupQuery = a.searchQuery
				a.groupFilter = 0
				a.page = PageGroups
			}
			ui.Box(c).Width(12)
			a.ghostButton(c, icons.bell, "通知", func() { c.Toast("暂无新通知") })
			ui.Box(c).Width(8)
			a.primaryButton(c, icons.plus, "新建项目", a.openNewProject)
			ui.Box(c).Width(10)
			ui.Avatar(c, "Lin", nil).Size(26, 26)
		})
}
