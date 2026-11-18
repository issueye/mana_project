package app

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

func (a *App) homePage(c *ui.Context) {
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Children(func() {
			a.homeHead(c)
			a.ledgerRow(c)
			ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
				a.homeMain(c)
				w, _ := c.Size()
				if w >= 1000 {
					a.homeSide(c)
				}
			})
		})
	})
}

// homeHead is the page header: a breadcrumb, the display title with the
// date, one caption line and the actions.
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
	ui.Column(c).Padding(18, 24, 14, 24).Gap(4).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "工作台").FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, "索引 01").FontSize(10).FontWeight(600).TextColor(colSeal)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Baseline).Children(func() {
			ui.Text(c, "本地项目索引").FontSize(29).FontWeight(700)
			monoText(c, now.Format("2006.01.02")+" · "+weekday, 10.5).TextColor(colInk3)
			ui.Spacer(c)
			week := ui.Row(c).Height(30).Padding(0, 10).Radius(2).Gap(6).
				Border(1, colLine).Background(colSheet).AlignItems(ui.Center).
				Focusable().Role(ui.RoleButton).Label("本周")
			week.Children(func() {
				ui.Icon(c, icons.calendar).FontSize(13).TextColor(colInk3)
				ui.Text(c, "本周").FontSize(11.5).TextColor(colInk2)
			})
			ui.Box(c).Width(8)
			a.primaryButton(c, icons.plus, "新建项目", a.openNewProject)
		})
		ui.Text(c, fmt.Sprintf("把项目、项目组与文档收进一份可检索的本地索引。今天有 %d 个项目更新，%d 篇文档有改动。", updatedToday, newDocsToday)).
			FontSize(11.5).TextColor(colInk3)
	})
}

// ledgerRow is the four-cell bookkeeping strip under the header.
func (a *App) ledgerRow(c *ui.Context) {
	projects, groups, docs, week := a.stats()
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	weekDocs := 0
	for _, d := range a.docs {
		if d.UpdatedAt.After(cutoff) {
			weekDocs++
		}
	}
	projectsMeta := "开始登记第一个项目"
	if projects > 0 {
		projectsMeta = fmt.Sprintf("其中 %d 个项目本周有更新", week)
	}
	groupsMeta := "还没有项目组"
	if groups > 0 {
		groupsMeta = "本地索引中的项目组"
	}
	docsMeta := "还没有文档"
	if docs > 0 {
		docsMeta = fmt.Sprintf("%d 篇文档本周有改动", weekDocs)
	}
	weekMeta := "还没有更新"
	if week > 0 {
		weekMeta = "近 7 天有更新的项目"
	}
	ui.Column(c).Children(func() {
		hairline(c)
		ui.Row(c).Height(92).Padding(0, 24).Children(func() {
			a.ledgerCell(c, icons.grid, "项目总数 · PROJECTS", "项", itoa(projects), projectsMeta, true)
			a.ledgerCell(c, icons.folder, "项目组 · GROUPS", "组", itoa(groups), groupsMeta, false)
			a.ledgerCell(c, icons.file, "文档总数 · DOCS", "篇", itoa(docs), docsMeta, false)
			a.ledgerCell(c, icons.refresh, "本周更新 · WEEK", "项", itoa(week), weekMeta, false)
		})
		hairline(c)
	})
}

// homeMain is the left column: the recent project table and the recent
// documents strip.
func (a *App) homeMain(c *ui.Context) {
	list := a.recentProjects(6)
	w, _ := c.Size()
	narrow := w < 1280
	ui.Column(c).Grow(1).Padding(12, 24, 16, 24).Children(func() {
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "最近更新")
			ui.Box(c).Width(8)
			monoText(c, "01 — "+itoaPad(len(list)), 9.5).TextColor(colInk4)
			ui.Spacer(c)
			a.linkMore(c, fmt.Sprintf("全部 %d 项", len(a.projects)), func() { a.openGroup(0) })
		})
		if len(list) == 0 {
			a.emptyState(c, "还没有项目", "从第一个项目开始建立本地索引。登记名称与路径，或先建一个项目组。",
				[2]any{"新建项目", func() { a.openNewProject() }},
				[2]any{"新建项目组", func() { a.openNewGroup() }})
			return
		}
		a.homeTable.Selected = &a.homeProjectRow
		a.homeTable.Key = func(row int) any { return list[row].ID }
		a.homeTable.Label = func(row int) string { return list[row].Name }
		kinds := []struct {
			title string
			width float32
			draw  func(p model.Project, row int)
		}{
			{title: "#", width: 36, draw: func(p model.Project, row int) {
				monoText(c, itoaPad(row+1), 10).TextColor(colInk4)
			}},
			{title: "项目", draw: func(p model.Project, row int) {
				ui.Column(c).Gap(1).Children(func() {
					ui.Text(c, p.Name).FontSize(12.5).FontWeight(600).SingleLine()
					ui.Text(c, p.Description).FontSize(9.5).TextColor(colInk3).SingleLine()
				})
			}},
			{title: "版本", width: 76, draw: func(p model.Project, row int) {
				a.versionChip(c, p.Version)
			}},
			{title: "标签", width: 150, draw: func(p model.Project, row int) {
				tagsText(c, p.Tags, 2)
			}},
			{title: "更新", width: 76, draw: func(p model.Project, row int) {
				ui.Text(c, model.Relative(p.UpdatedAt)).FontSize(10.5).TextColor(colInk3)
			}},
			{title: "", width: 30, draw: func(p model.Project, row int) {
				ui.Icon(c, icons.more).FontSize(14).TextColor(colInk4)
			}},
		}
		if narrow {
			kinds = append(kinds[:3:3], kinds[4]) // drop the tags column
		}
		cols := make([]ui.TableColumn, len(kinds))
		for i, k := range kinds {
			cols[i] = ui.TableColumn{Title: k.title, Width: k.width}
		}
		tbl := ui.Table(c, &a.homeTable, cols, len(list), func(row, col int) {
			p := list[row]
			open := func(build func()) {
				r := ui.Row(c).MinHeight(44).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label(p.Name)
				r.OnClick(func() { a.openProject(p.ID) })
				r.Children(build)
			}
			open(func() { kinds[col].draw(p, row) })
		}).FillWidth().Padding(0, 0).Grow(1)
		if tbl.Changed() && a.homeProjectRow >= 0 && a.homeProjectRow < len(list) {
			id := list[a.homeProjectRow].ID
			a.homeProjectRow = -1
			a.openProject(id)
		}
		ui.Row(c).Height(30).Children(func() {
			ui.Text(c, fmt.Sprintf("共 %d 个项目 · 按更新时间降序", len(a.projects))).FontSize(9.5).TextColor(colInk4)
			ui.Spacer(c)
			monoText(c, "SYNC "+time.Now().Format("15:04"), 9).TextColor(colInk4)
		})

		ui.Box(c).Height(1).Background(colLine2).Margin(6, 0, 10, 0)
		docs := a.recentDocs(4)
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "最近文档")
			ui.Box(c).Width(8)
			monoText(c, itoaPad(len(docs)), 9.5).TextColor(colInk4)
			ui.Spacer(c)
			a.linkMore(c, fmt.Sprintf("全部 %d 篇", len(a.docs)), func() { a.goGroups() })
		})
		if len(docs) == 0 {
			ui.Text(c, "还没有文档，打开一个项目即可开始记录。").FontSize(11).TextColor(colInk3)
			return
		}
		for i := 0; i < len(docs); i += 2 {
			ui.Row(c).Grow(1).Gap(24).Children(func() {
				for j := i; j < i+2 && j < len(docs); j++ {
					d := docs[j]
					item := ui.Row(c).Height(46).Grow(1).Gap(9).AlignItems(ui.Center).
						Focusable().Role(ui.RoleButton).Label(d.Title)
					item.OnClick(func() { a.openDoc(d.ID) })
					item.Children(func() {
						kindDot(c, d.Kind)
						ui.Column(c).Grow(1).Gap(1).Children(func() {
							ui.Text(c, d.Title).FontSize(11.5).FontWeight(600).SingleLine()
							ui.Text(c, d.OwnerName).FontSize(9.5).TextColor(colInk3).SingleLine()
						})
						ui.Text(c, model.Relative(d.UpdatedAt)).FontSize(9.5).TextColor(colInk4)
					})
				}
			})
		}
	})
}

// homeSide is the right column: the group ledger and the week review.
func (a *App) homeSide(c *ui.Context) {
	w, _ := c.Size()
	sideWidth := float32(360)
	if w < 1200 {
		sideWidth = 300
	}
	ui.Column(c).Width(sideWidth).Background(colSheet2).
		BorderWidth(0, 0, 0, 1).BorderColor(colLine2).
		Padding(12, 20, 12, 20).Gap(2).Children(func() {
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "项目组")
			ui.Box(c).Width(6)
			monoText(c, itoaPad(len(a.groups)), 9.5).TextColor(colInk4)
			ui.Spacer(c)
			a.linkMore(c, "管理", a.goGroups)
		})
		for _, g := range a.groups {
			group := g
			row := ui.Row(c).Height(34).Padding(0, 6).Gap(8).Radius(2).
				Focusable().Role(ui.RoleButton).Label(group.Name)
			if row.Hovered() {
				row.Background(colSheet)
			}
			row.OnClick(func() { a.openGroup(group.ID) })
			row.Children(func() {
				ui.Box(c).Size(7, 7).Radius(1).Background(ui.Hex(group.Color))
				ui.Text(c, group.Name).FontSize(11.5).TextColor(colInk).SingleLine()
				ui.Spacer(c)
				ui.Text(c, fmt.Sprintf("%d 项 · %d 篇", group.ProjectCount, group.DocumentCount)).
					FontSize(9.5).TextColor(colInk3)
				ui.Icon(c, icons.chevronRight).FontSize(11).TextColor(colInk4)
			})
		}
		ungrouped := 0
		for _, p := range a.projects {
			if p.GroupID == nil {
				ungrouped++
			}
		}
		if ungrouped > 0 {
			row := ui.Row(c).Height(34).Padding(0, 6).Gap(8).Radius(2).
				Focusable().Role(ui.RoleButton).Label("未分组")
			if row.Hovered() {
				row.Background(colSheet)
			}
			row.OnClick(func() { a.openGroup(-1) })
			row.Children(func() {
				ui.Box(c).Size(7, 7).Radius(1).Background(colInk4)
				ui.Text(c, "未分组").FontSize(11.5).TextColor(colInk)
				ui.Spacer(c)
				ui.Text(c, fmt.Sprintf("%d 项", ungrouped)).FontSize(9.5).TextColor(colInk3)
				ui.Icon(c, icons.chevronRight).FontSize(11).TextColor(colInk4)
			})
		}

		ui.Box(c).Height(1).Background(colLine2).Margin(8, 0, 10, 0)
		review := a.weekReview(4)
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "本周回顾")
			ui.Box(c).Width(6)
			monoText(c, itoaPad(len(review)), 9.5).TextColor(colInk4)
			ui.Spacer(c)
			ui.Text(c, "统计").FontSize(11).TextColor(colInk4)
		})
		for _, item := range review {
			ui.Row(c).Height(32).Gap(9).AlignItems(ui.Center).Children(func() {
				monoText(c, item.day, 9).TextColor(colInk4)
				ui.Text(c, item.text).FontSize(10.5).TextColor(colInk2).SingleLine()
				ui.Spacer(c)
				monoText(c, item.value, 10).FontWeight(600).TextColor(colSeal)
			})
		}
		ui.Spacer(c)
		ui.Row(c).Height(26).Gap(4).AlignItems(ui.Center).Children(func() {
			_, _, _, week := a.stats()
			cutoff := time.Now().Add(-7 * 24 * time.Hour)
			weekDocs := 0
			for _, d := range a.docs {
				if d.UpdatedAt.After(cutoff) {
					weekDocs++
				}
			}
			ui.Text(c, fmt.Sprintf("本周 %d 个项目有更新 · %d 篇文档有改动", week, weekDocs)).
				FontSize(9.5).TextColor(colInk3)
			ui.Icon(c, icons.chevronRight).FontSize(10).TextColor(colInk4)
		})
	})
}

type reviewItem struct {
	day   string
	text  string
	value string
}

// weekReview merges recently touched projects and documents into the
// short activity ledger on the right of the home page.
func (a *App) weekReview(limit int) []reviewItem {
	day := func(t time.Time) string {
		names := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
		return names[t.Weekday()]
	}
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var items []reviewItem
	for _, d := range a.recentDocs(limit) {
		if d.UpdatedAt.Before(cutoff) {
			continue
		}
		items = append(items, reviewItem{day: day(d.UpdatedAt), text: d.Title, value: itoa(d.Words)})
	}
	for _, p := range a.recentProjects(limit) {
		if p.UpdatedAt.Before(cutoff) {
			continue
		}
		text := p.Name + " 更新"
		if p.Version != "" {
			text = p.Name + " 发布 " + p.Version
		}
		items = append(items, reviewItem{day: day(p.UpdatedAt), text: text, value: model.Relative(p.UpdatedAt)})
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Local().Date()
	y2, m2, d2 := b.Local().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
