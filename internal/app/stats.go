package app

import (
	"fmt"
	"sort"
	"time"

	"github.com/egoist/mygo/ui"
)

// statsPage is the whole-index view: the same four-cell ledger as home,
// then where the projects and documents actually sit.
func (a *App) statsPage(c *ui.Context) {
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Children(func() {
			a.statsHead(c)
			a.ledgerRow(c)
			hairline(c)
			ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
				a.statsMain(c)
				w, _ := c.Size()
				if w >= 1000 {
					a.statsSide(c)
				}
			})
		})
	})
}

func (a *App) statsHead(c *ui.Context) {
	ui.Column(c).Padding(18, 24, 14, 24).Gap(4).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "工作台").FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, "统计 01").FontSize(10).FontWeight(600).TextColor(colSeal)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Baseline).Children(func() {
			ui.Text(c, "数据统计").FontSize(29).FontWeight(700)
			monoText(c, "SNAPSHOT · "+time.Now().Format("15:04"), 10.5).TextColor(colInk3)
		})
		ui.Text(c, "整张索引的分布：项目落在哪些组，文档是什么类型，最近 7 天谁在动。").
			FontSize(11.5).TextColor(colInk3)
	})
}

// statsMain is the left column: group distribution and the week activity.
func (a *App) statsMain(c *ui.Context) {
	ui.Column(c).Grow(1).Padding(12, 24, 16, 24).Gap(10).Children(func() {
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "项目组分布")
			ui.Spacer(c)
			monoText(c, itoaPad(len(a.groups))+" 组", 9.5).TextColor(colInk4)
		})
		a.statsGroupBars(c)

		ui.Box(c).Height(1).Background(colLine2).Margin(6, 0)
		ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "近 7 天活跃")
			ui.Spacer(c)
			monoText(c, "项目 · 文档", 9).TextColor(colInk4)
		})
		a.statsWeekBars(c)
	})
}

// statsSide is the right column: document types and the top tags.
func (a *App) statsSide(c *ui.Context) {
	w, _ := c.Size()
	sideWidth := float32(360)
	if w < 1200 {
		sideWidth = 300
	}
	ui.Column(c).Width(sideWidth).Background(colSheet2).
		BorderWidth(0, 0, 0, 1).BorderColor(colLine2).
		Padding(12, 20, 12, 20).Gap(2).Children(func() {
			ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
				sectionTitle(c, "文档类型")
				ui.Spacer(c)
				monoText(c, itoaPad(len(a.docs))+" 篇", 9.5).TextColor(colInk4)
			})
			a.docTypeLedger(c)
			ui.Box(c).Height(1).Background(colLine2).Margin(8, 0, 10, 0)
			ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
				sectionTitle(c, "标签 Top")
				ui.Spacer(c)
				monoText(c, itoaPad(len(a.tagCounts()))+" 个", 9.5).TextColor(colInk4)
			})
			for _, tc := range a.tagCounts() {
				ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
					ui.Text(c, tc.tag).FontSize(10.5).TextColor(colInk2).SingleLine()
					ui.Spacer(c)
					monoText(c, itoa(tc.count), 10).FontWeight(600).TextColor(colSeal)
				})
			}
			ui.Spacer(c)
			ui.Row(c).Height(26).Gap(4).AlignItems(ui.Center).Children(func() {
				projects, groups, docs, week := a.stats()
				ui.Text(c, fmt.Sprintf("索引共 %d 项目 · %d 组 · %d 文档 · 本周 %d 项更新", projects, groups, docs, week)).
					FontSize(9.5).TextColor(colInk3)
			})
		})
}

type tagCount struct {
	tag   string
	count int
}

// tagCounts returns the most used tags across projects.
func (a *App) tagCounts() []tagCount {
	counts := map[string]int{}
	for _, p := range a.projects {
		for _, t := range p.Tags {
			counts[t]++
		}
	}
	out := make([]tagCount, 0, len(counts))
	for tag, n := range counts {
		out = append(out, tagCount{tag, n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].tag < out[j].tag
	})
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

// statsGroupBars places every group on two hairline scales.
func (a *App) statsGroupBars(c *ui.Context) {
	maxProjects, maxDocs := 1, 1
	for _, g := range a.groups {
		if g.ProjectCount > maxProjects {
			maxProjects = g.ProjectCount
		}
		if g.DocumentCount > maxDocs {
			maxDocs = g.DocumentCount
		}
	}
	ungrouped := len(a.projectsInGroup(nil))
	if ungrouped > maxProjects {
		maxProjects = ungrouped
	}
	ui.Column(c).Gap(2).Children(func() {
		draw := func(name string, color ui.Color, projects, docs int) {
			ui.Row(c).Height(30).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(7, 7).Radius(1).Background(color)
				ui.Text(c, name).FontSize(11).TextColor(colInk).SingleLine()
				ui.Spacer(c)
				a.statsBar(c, projects, maxProjects, colInk3)
				a.statsBar(c, docs, maxDocs, colInk4)
				ui.Box(c).Width(46).Children(func() {
					monoText(c, fmt.Sprintf("%02d · %02d", projects, docs), 9.5).TextColor(colInk3)
				})
			})
		}
		for _, g := range a.groups {
			group := g
			draw(group.Name, ui.Hex(group.Color), group.ProjectCount, group.DocumentCount)
		}
		if ungrouped > 0 {
			draw("未分组", colInk4, ungrouped, 0)
		}
	})
}

func (a *App) statsBar(c *ui.Context, value, max int, color ui.Color) {
	width := float32(120)
	if max > 0 {
		width = 120 * float32(value) / float32(max)
	}
	ui.Box(c).Width(120).Height(3).Background(colLine2).Children(func() {
		ui.Box(c).Width(width).Height(3).Background(color)
	})
}

// statsWeekBars counts, per day, the projects and documents touched.
func (a *App) statsWeekBars(c *ui.Context) {
	now := time.Now()
	type dayRow struct {
		label    string
		projects int
		docs     int
	}
	days := make([]dayRow, 7)
	for i := 0; i < 7; i++ {
		d := now.AddDate(0, 0, -(6 - i))
		days[i].label = [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}[d.Weekday()]
	}
	for _, p := range a.projects {
		for i := range days {
			if sameDay(p.UpdatedAt, now.AddDate(0, 0, -(6-i))) {
				days[i].projects++
			}
		}
	}
	for _, d := range a.docs {
		for i := range days {
			if sameDay(d.UpdatedAt, now.AddDate(0, 0, -(6-i))) {
				days[i].docs++
			}
		}
	}
	max := 1
	for _, d := range days {
		if d.projects > max {
			max = d.projects
		}
		if d.docs > max {
			max = d.docs
		}
	}
	ui.Row(c).Height(96).Gap(4).AlignItems(ui.Stretch).Children(func() {
		for _, d := range days {
			day := d
			ui.Column(c).Grow(1).Gap(4).AlignItems(ui.Center).Justify(ui.End).Padding(0, 0, 4, 0).Children(func() {
				ui.Row(c).Gap(3).AlignItems(ui.End).Children(func() {
					a.weekBar(c, day.projects, max, colSeal)
					a.weekBar(c, day.docs, max, colInk4)
				})
				monoText(c, fmt.Sprintf("%02d", day.projects+day.docs), 9).TextColor(colInk3)
				ui.Text(c, day.label).FontSize(9).TextColor(colInk4)
			})
		}
	})
}

func (a *App) weekBar(c *ui.Context, value, max int, color ui.Color) {
	height := float32(4)
	if max > 0 {
		height = 44 * float32(value) / float32(max)
	}
	if height < 2 {
		height = 2
	}
	ui.Box(c).Width(8).Height(height).Radius(1).Background(color)
}

// statsContext is the context panel of the statistics page: the type
// ledger and a compact group list.
func (a *App) statsContext(c *ui.Context) {
	ui.Column(c).FillHeight().Padding(14, 14, 12, 14).Gap(10).Children(func() {
		a.panelHead(c, "统计目录", itoaPad(len(a.groups)), nil)
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			microLabel(c, "文档类型 · " + fmt.Sprintf("%02d", len(a.docs)))
			ui.Spacer(c)
			microLabel(c, "TYPE")
		})
		a.docTypeLedger(c)
		ui.Row(c).Padding(8, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			microLabel(c, "项目组 · " + fmt.Sprintf("%02d", len(a.groups)))
			ui.Spacer(c)
			microLabel(c, "GROUPS")
		})
		for _, g := range a.groups {
			group := g
			a.contextGroupRow(c, icons.folder, group.Name, group.ProjectCount, false, func() { a.openGroup(group.ID) })
		}
		ui.Spacer(c)
		a.dashedButton(c, "打开首页", func() { a.goHome() })
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(5, 5).Radius(1).Background(colMoss)
			ui.Text(c, "统计来自当前内存快照").FontSize(9).TextColor(colInk4)
		})
	})
}
