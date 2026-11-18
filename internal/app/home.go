package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

func (a *App) homePage(c *ui.Context) {
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).Padding(18, 20, 20, 20).Gap(12).Children(func() {
			a.homeHead(c)
			a.statsRow(c)
			ui.Row(c).Height(410).Gap(12).AlignItems(ui.Stretch).Children(func() {
				left := a.card(c, 0, func() { a.recentProjectsCard(c) })
				left.Grow(2)
				right := a.card(c, 0, func() { a.groupsOverviewCard(c) })
				right.Grow(1)
			})
			doc := a.card(c, 0, func() { a.recentDocsCard(c) })
			doc.Height(176)
		})
	})
}

func (a *App) homeHead(c *ui.Context) {
	now := time.Now()
	weekday := [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}[now.Weekday()]
	updatedToday := 0
	for _, p := range a.projects {
		if sameDay(p.UpdatedAt, now) {
			updatedToday++
		}
	}
	newDocsToday := 0
	for _, d := range a.docs {
		if sameDay(d.UpdatedAt, now) {
			newDocsToday++
		}
	}
	ui.Row(c).Height(62).Children(func() {
		ui.Column(c).Gap(2).Children(func() {
			ui.Text(c, fmt.Sprintf("%s · %s", now.Format("2006 年 01 月 02 日"), weekday)).
				FontSize(10.5).TextColor(colText3)
			ui.Text(c, greeting(now.Hour())).FontSize(20).Bold()
			ui.Text(c, fmt.Sprintf("今天有 %d 个项目更新，%d 篇文档有改动。", updatedToday, newDocsToday)).
				FontSize(11.5).TextColor(colText3)
		})
		ui.Spacer(c)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Height(28).Padding(0, 9).Radius(4).Border(1, colBorder).Background(colSurface).Center().Children(func() {
				ui.Icon(c, icons.calendar).FontSize(14).TextColor(colText3)
				ui.Text(c, "本周").FontSize(12).TextColor(colText2)
			})
			ui.Row(c).Height(28).Padding(0, 9).Radius(4).Border(1, colBorder).Background(colSurface).Center().Children(func() {
				ui.Icon(c, icons.clock).FontSize(14).TextColor(colText3)
				ui.Text(c, "更新于 1 分钟前").FontSize(12).TextColor(colText2)
			})
		})
	})
}

func (a *App) statsRow(c *ui.Context) {
	projects, groups, docs, week := a.stats()
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	weekDocs := 0
	activeGroups := map[int64]bool{}
	for _, d := range a.docs {
		if d.UpdatedAt.After(cutoff) {
			weekDocs++
		}
	}
	for _, p := range a.projects {
		if p.UpdatedAt.After(cutoff) && p.GroupID != nil {
			activeGroups[*p.GroupID] = true
		}
	}
	projectsMeta := "开始登记第一个项目"
	if projects > 0 {
		projectsMeta = fmt.Sprintf("%d 个项目本周有更新", week)
	}
	groupsMeta := "还没有项目组"
	if groups > 0 {
		groupsMeta = fmt.Sprintf("%d 个组有项目更新", len(activeGroups))
	}
	docsMeta := "还没有文档"
	if docs > 0 {
		docsMeta = fmt.Sprintf("%d 篇文档本周有改动", weekDocs)
	}
	weekMeta := "还没有更新"
	if week > 0 {
		weekMeta = "近 7 天有更新的项目"
	}
	ui.Row(c).Height(88).Gap(12).Children(func() {
		a.statCard(c, icons.grid, "项目总数", itoa(projects), projectsMeta, colPrimary, colPrimary1).Grow(1)
		a.statCard(c, icons.folder, "项目组", itoa(groups), groupsMeta, colPurple, colPurple1).Grow(1)
		a.statCard(c, icons.file, "文档总数", itoa(docs), docsMeta, colCyan, colCyan1).Grow(1)
		a.statCard(c, icons.refresh, "本周更新", itoa(week), weekMeta, colSuccess, colSuccess1).Grow(1)
	})
}

func (a *App) recentProjectsCard(c *ui.Context) {
	ui.Row(c).Height(44).Padding(0, 16).Children(func() {
		ui.Column(c).Gap(1).Children(func() {
			ui.Text(c, "最近更新项目").FontSize(13).Bold()
			ui.Text(c, "按更新时间排序，点击进入项目详情").FontSize(10.5).TextColor(colText3)
		})
		ui.Spacer(c)
		a.linkMore(c, "全部项目", func() { a.openGroup(0) })
	})
	ui.Box(c).Height(1).Background(colBorder2)

	list := a.recentProjects(6)
	if len(list) == 0 {
		a.emptyState(c, icons.grid, "还没有项目", "点击右上角「新建项目」，开始登记第一个项目。", "新建项目", a.openNewProject)
		return
	}
	a.homeTable.Selected = &a.homeProjectRow
	a.homeTable.Key = func(row int) any { return list[row].ID }
	a.homeTable.Label = func(row int) string { return list[row].Name }
	cols := []ui.TableColumn{
		{Title: "项目"},
		{Title: "版本", Width: 72},
		{Title: "标签", Width: 150},
		{Title: "更新时间", Width: 88},
		{Title: "", Width: 26},
	}
	tbl := ui.Table(c, &a.homeTable, cols, len(list), func(row, col int) {
		p := list[row]
		switch col {
		case 0:
			a.clickRow(c, 40, func() { a.openProject(p.ID) }, func() {
				ui.Row(c).Gap(10).Children(func() {
					icon, fg, bg := projectIcon(p)
					a.iconTile(c, icon, 30, fg, bg)
					ui.Column(c).Gap(1).Children(func() {
						ui.Text(c, p.Name).FontSize(12.5).Bold().SingleLine()
						ui.Text(c, p.GroupName).FontSize(10).TextColor(colText3).SingleLine()
					})
				})
			})
		case 1:
			a.clickRow(c, 40, func() { a.openProject(p.ID) }, func() {
				a.versionChip(c, p.Version)
			})
		case 2:
			a.clickRow(c, 40, func() { a.openProject(p.ID) }, func() {
				a.tagsCell(c, p.Tags)
			})
		case 3:
			a.clickRow(c, 40, func() { a.openProject(p.ID) }, func() {
				ui.Text(c, model.Relative(p.UpdatedAt)).FontSize(10.5).TextColor(colText3)
			})
		case 4:
			ui.Icon(c, icons.more).FontSize(15).TextColor(colText3)
		}
	}).FillWidth().Padding(4, 12).Grow(1)
	if tbl.Changed() && a.homeProjectRow >= 0 && a.homeProjectRow < len(list) {
		id := list[a.homeProjectRow].ID
		a.homeProjectRow = -1
		a.openProject(id)
	}
	ui.Row(c).Height(30).Padding(0, 16).BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, fmt.Sprintf("共 %d 个项目", len(a.projects))).FontSize(10).TextColor(colText3)
		ui.Spacer(c)
		ui.Text(c, "最近同步 "+time.Now().Format("15:04")).FontSize(10).TextColor(colText3)
	})
}

func (a *App) groupsOverviewCard(c *ui.Context) {
	ui.Row(c).Height(44).Padding(0, 16).Gap(6).Children(func() {
		ui.Text(c, "项目组").FontSize(13).Bold()
		ui.Text(c, itoa(len(a.groups))).FontSize(10.5).TextColor(colText3)
		ui.Spacer(c)
		a.linkMore(c, "管理", a.goGroups)
	})
	ui.Box(c).Height(1).Background(colBorder2)
	ui.Column(c).Gap(0).Padding(6, 8, 4, 8).Children(func() {
		for _, g := range a.groups {
			group := g
			row := ui.Row(c).Height(36).Padding(0, 8).Gap(8).Radius(4).
				Focusable().Role(ui.RoleButton).Label(group.Name)
			if row.Hovered() {
				row.Background(ui.Hex("#F7F9FC"))
			}
			row.OnClick(func() { a.openGroup(group.ID) })
			row.Children(func() {
				ui.Box(c).Size(7, 7).Radius(4).Background(ui.Hex(group.Color))
				ui.Text(c, group.Name).FontSize(12).TextColor(colText).SingleLine()
				ui.Spacer(c)
				ui.Text(c, fmt.Sprintf("%d 项目 · %d 文档", group.ProjectCount, group.DocumentCount)).
					FontSize(10).TextColor(colText3)
				ui.Icon(c, icons.chevronRight).FontSize(12).TextColor(colText4)
			})
		}
		ungrouped := 0
		for _, p := range a.projects {
			if p.GroupID == nil {
				ungrouped++
			}
		}
		if ungrouped > 0 {
			row := ui.Row(c).Height(36).Padding(0, 8).Gap(8).Radius(4).
				Focusable().Role(ui.RoleButton).Label("未分组")
			if row.Hovered() {
				row.Background(ui.Hex("#F7F9FC"))
			}
			row.OnClick(func() { a.openGroup(-1) })
			row.Children(func() {
				ui.Box(c).Size(7, 7).Radius(4).Background(colText4)
				ui.Text(c, "未分组").FontSize(12).TextColor(colText)
				ui.Spacer(c)
				ui.Text(c, fmt.Sprintf("%d 项目", ungrouped)).FontSize(10).TextColor(colText3)
				ui.Icon(c, icons.chevronRight).FontSize(12).TextColor(colText4)
			})
		}
	})
	ui.Spacer(c)
	ui.Column(c).Padding(6, 16, 12, 16).Children(func() {
		button := ui.Row(c).Height(30).Radius(4).Center().
			Border(1, colBorder).BorderStyle(ui.BorderDashed).
			Focusable().Role(ui.RoleButton).Label("新建项目组")
		button.OnClick(a.openNewGroup)
		button.Children(func() {
			ui.Icon(c, icons.plus).FontSize(13).TextColor(colText3)
			ui.Text(c, "新建项目组").FontSize(11.5).TextColor(colText3)
		})
	})
}

func (a *App) recentDocsCard(c *ui.Context) {
	ui.Row(c).Height(40).Padding(0, 16).Children(func() {
		ui.Text(c, "最近文档").FontSize(13).Bold()
		ui.Spacer(c)
		a.linkMore(c, "全部文档", func() { a.goGroups() })
	})
	ui.Box(c).Height(1).Background(colBorder2)
	docs := a.recentDocs(4)
	if len(docs) == 0 {
		ui.Row(c).Grow(1).Center().Children(func() {
			ui.Text(c, "还没有文档，打开一个项目即可开始记录。").FontSize(11.5).TextColor(colText3)
		})
		return
	}
	ui.Column(c).Gap(0).Padding(4, 16, 6, 16).Children(func() {
		for i := 0; i < len(docs); i += 2 {
			ui.Row(c).Gap(18).Children(func() {
				for j := i; j < i+2 && j < len(docs); j++ {
					d := docs[j]
					item := ui.Row(c).Height(46).Grow(1).Gap(9).
						Focusable().Role(ui.RoleButton).Label(d.Title)
					item.OnClick(func() { a.openDoc(d.ID) })
					item.Children(func() {
						a.iconTile(c, icons.note, 30, colText3, colSurface2)
						ui.Column(c).Grow(1).Gap(1).Children(func() {
							ui.Text(c, d.Title).FontSize(12).Bold().SingleLine()
							ui.Row(c).Gap(5).Children(func() {
								a.kindTag(c, d.Kind)
								ui.Text(c, d.OwnerName).FontSize(10).TextColor(colText3).SingleLine()
							})
						})
						ui.Text(c, model.Relative(d.UpdatedAt)).FontSize(10).TextColor(colText4)
					})
				}
			})
		}
	})
}

func (a *App) linkMore(c *ui.Context, label string, fn func()) {
	b := ui.Button(c, "").Height(24).Padding(0, 4).Radius(4).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		ui.Text(c, label).FontSize(11.5).TextColor(colText2)
		ui.Icon(c, icons.chevronRight).FontSize(12).TextColor(colText4)
	})
}

func (a *App) tagsCell(c *ui.Context, tags []string) {
	ui.Row(c).Gap(4).Children(func() {
		for i, tag := range tags {
			if i >= 2 {
				break
			}
			color := colPrimary
			bg := colPrimary1
			if i == 1 {
				color = colCyan
				bg = colCyan1
			}
			a.tag(c, tag, color, bg)
		}
		if len(tags) > 2 {
			a.tag(c, fmt.Sprintf("+%d", len(tags)-2), colText2, ui.Hex("#F4F5F7"))
		}
	})
}

func projectIcon(p model.Project) (*ui.SVG, ui.Color, ui.Color) {
	for _, tag := range p.Tags {
		switch strings.ToLower(tag) {
		case "sqlite", "database", "数据":
			return icons.database, ui.Hex("#389E0D"), colSuccess1
		case "flutter", "ios", "android":
			return icons.grid, colCyan, colCyan1
		case "swift", "macos":
			return icons.code, colPurple, colPurple1
		case "python", "ml":
			return icons.gitBranch, colOrange, colOrange1
		case "docker", "ci":
			return icons.terminal, colPurple, colPurple1
		}
	}
	return icons.code, colPrimary, colPrimary1
}

func greeting(hour int) string {
	switch {
	case hour < 6:
		return "夜深了，Lin"
	case hour < 12:
		return "上午好，Lin"
	case hour < 18:
		return "下午好，Lin"
	default:
		return "晚上好，Lin"
	}
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Local().Date()
	y2, m2, d2 := b.Local().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
