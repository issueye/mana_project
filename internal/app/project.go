package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

// projectPage is the record card: the header keeps the facts the developer
// needs in one line; the main column switches between the documents, the
// file browser and the Git panel; the right column is the record itself
// behind a hairline.
func (a *App) projectPage(c *ui.Context) {
	p := a.currentProject
	w, _ := c.Size()
	infoWidth := float32(320)
	if w < 1280 {
		infoWidth = 288
	}
	ui.Column(c).Fill().Children(func() {
		a.detailHead(c, p)
		a.projectViewTabs(c, p)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			switch a.projectView {
			case 1:
				a.filesView(c, p)
			case 2:
				a.repoView(c, p)
			case 3:
				a.todosView(c, p)
			default:
				a.projectDocsColumn(c, p)
			}
			a.recordCard(c, p, infoWidth)
		})
	})
}

// projectViewTabs is the 文档 / 文件 / 仓库 / 待办 switch of the main
// column; the todo tab carries the open count.
func (a *App) projectViewTabs(c *ui.Context, p model.Project) {
	ui.Column(c).Padding(0, 24).Children(func() {
		hairline(c)
		ui.Row(c).Height(42).AlignItems(ui.Center).Children(func() {
			ui.Tabs(c, &a.projectView,
				fmt.Sprintf("文档 %02d", p.DocumentCount), "文件", "仓库",
				fmt.Sprintf("待办 %02d", a.undoneTodos(p.ID)))
			ui.Spacer(c)
		})
		hairline(c)
	})
}

// detailHead is the project header: breadcrumb, name with version, the
// description and one meta line with tags, path and freshness.
func (a *App) detailHead(c *ui.Context, p model.Project) {
	groupName := p.GroupName
	if groupName == "" {
		groupName = "未分组"
	}
	ui.Column(c).Padding(18, 24, 14, 24).Gap(4).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "索引").FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, groupName).FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, "项目 01").FontSize(10).FontWeight(600).TextColor(colSeal)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, p.Name).FontSize(29).FontWeight(700).SingleLine()
			a.versionChip(c, p.Version)
			ui.Spacer(c)
			a.borderedButton(c, icons.folderOpen, "打开目录", func() {
				a.openProjectFolder(c, p)
			})
			ui.Box(c).Width(7)
			a.borderedButton(c, icons.edit, "编辑项目", a.openEditProject)
			ui.Box(c).Width(7)
			a.dangerButton(c, icons.trash, "删除项目", a.askDeleteProject)
		})
		ui.Text(c, p.Description).FontSize(11.5).TextColor(colInk3).SingleLine()
		ui.Row(c).Gap(12).Padding(2, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			tagsText(c, p.Tags, 4)
			a.metaDot(c)
			a.pathChip(c, p.Path)
			a.metaDot(c)
			a.metaItem(c, icons.file, fmt.Sprintf("%d 篇文档", p.DocumentCount))
			a.metaDot(c)
			a.metaItem(c, icons.clock, model.Relative(p.UpdatedAt)+"更新")
		})
	})
}

func (a *App) metaDot(c *ui.Context) {
	ui.Box(c).Size(2, 2).Radius(1).Background(colInk4)
}

// pathChip is the monospace path with a copy affordance.
func (a *App) pathChip(c *ui.Context, path string) {
	if path == "" {
		path = "未登记路径"
	}
	chip := ui.Row(c).Height(22).Padding(0, 7).Gap(6).Radius(2).
		Border(1, colLine).Background(colSheet2).AlignItems(ui.Center)
	chip.Children(func() {
		monoText(c, path, 9.5).TextColor(colInk2).SingleLine()
		b := ui.Box(c).Size(16, 16).Radius(2).Center().
			Focusable().Role(ui.RoleButton).Label("复制路径")
		b.OnClick(func() { c.WriteClipboard(path); c.Toast("路径已复制到剪贴板") })
		b.Children(func() { ui.Icon(c, icons.copy).FontSize(10).TextColor(colInk3) })
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

// projectDocsColumn is the left column: the document table and, under it,
// the preview of the latest document.
func (a *App) projectDocsColumn(c *ui.Context, p model.Project) {
	docs := a.projectDocs()
	w, _ := c.Size()
	narrow := w < 1280
	ui.Column(c).Grow(1).Padding(12, 24, 0, 24).Children(func() {
		ui.Row(c).Height(34).AlignItems(ui.Center).Children(func() {
			sectionTitle(c, "项目文档")
			ui.Box(c).Width(6)
			monoText(c, itoaPad(len(docs)), 9.5).TextColor(colInk4)
			ui.Spacer(c)
			ui.SearchField(c, &a.groupQuery).Width(200).Height(30).Placeholder("搜索标题或正文")
			ui.Box(c).Width(8)
			a.borderedButton(c, icons.file, "上传文档", func() {
				a.uploadDocument(model.OwnerProject, p.ID)
			})
			ui.Box(c).Width(8)
			a.primaryButton(c, icons.plus, "新建文档", func() {
				a.newDoc(model.OwnerProject, p.ID)
			})
		})
		if len(docs) == 0 {
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "这个项目还没有文档", "新建需求、设计或开发计划，文档会跟随项目。",
					[2]any{"新建文档", func() { a.newDoc(model.OwnerProject, p.ID) }})
			})
			return
		}
		a.detailDocTable.Selected = &a.docRow
		a.detailDocTable.Key = func(row int) any { return docs[row].ID }
		a.detailDocTable.Label = func(row int) string { return docs[row].Title }
		cols := []ui.TableColumn{
			{Title: "#", Width: 36},
			{Title: "标题"},
			{Title: "类型", Width: 96},
			{Title: "字数", Width: 70},
			{Title: "更新", Width: 80},
			{Title: "", Width: 30},
		}
		tbl := ui.Table(c, &a.detailDocTable, cols, len(docs), func(row, col int) {
			d := docs[row]
			switch col {
			case 0:
				monoText(c, itoaPad(row+1), 10).TextColor(colInk4)
			case 1:
				r := ui.Row(c).MinHeight(46).AlignItems(ui.Center).Gap(8).
					Focusable().Role(ui.RoleButton).Label(d.Title)
				r.OnClick(func() { a.openDoc(d.ID) })
				r.Children(func() {
					kindDot(c, d.Kind)
					ui.Text(c, d.Title).FontSize(12.5).FontWeight(600).SingleLine()
				})
			case 2:
				r := ui.Row(c).MinHeight(46).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label(model.KindLabel(d.Kind))
				r.OnClick(func() { a.openDoc(d.ID) })
				r.Children(func() {
					if d.FileName != "" {
						monoText(c, fileKindLabel(d), 10).TextColor(colInk3)
					} else {
						kindText(c, d.Kind)
					}
				})
			case 3:
				r := ui.Row(c).MinHeight(46).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label(itoa(d.Words))
				r.OnClick(func() { a.openDoc(d.ID) })
				r.Children(func() {
					if d.FileName != "" {
						monoText(c, byteSize(d.FileSize), 10).TextColor(colInk3)
					} else {
						monoText(c, itoa(d.Words), 10.5).TextColor(colInk3)
					}
				})
			case 4:
				r := ui.Row(c).MinHeight(46).AlignItems(ui.Center).
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
		}).FillWidth().Grow(1)
		if tbl.Changed() && a.detailDocTable.Selected != nil && *a.detailDocTable.Selected >= 0 && *a.detailDocTable.Selected < len(docs) {
			a.openDoc(docs[*a.detailDocTable.Selected].ID)
		}
		words := 0
		for _, d := range docs {
			words += d.Words
		}
		ui.Row(c).Height(30).BorderWidth(1, 0, 0, 0).BorderColor(colLine2).
			AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("共 %d 篇 · 合计 %d 字", len(docs), words)).FontSize(9.5).TextColor(colInk4)
			ui.Spacer(c)
			ui.Text(c, "最后更新于 "+model.Relative(docs[0].UpdatedAt)).FontSize(9.5).TextColor(colInk4)
		})
		if !narrow {
			a.docPreview(c, docs)
		}
	})
}

// docPreview shows the newest document of the project as a summary card.
// The snapshot carries metadata only, so the body is fetched once per
// document, off the frame, and the window refreshes when it arrives.
func (a *App) docPreview(c *ui.Context, docs []model.Document) {
	latest := docs[0]
	a.requestDocBody(latest.ID)
	summary := docSummary(a.previewBody, 2)
	if summary == "" {
		summary = "打开文档查看全文。"
	}
	ui.Column(c).Margin(10, 0, 16, 0).Background(colSheet2).
		Border(1, colLine2).Radius(2).Children(func() {
			ui.Row(c).Height(32).Padding(0, 14).AlignItems(ui.Center).Children(func() {
				sectionTitle(c, "文档预览")
				ui.Box(c).Width(8)
				monoText(c, "REQ · "+versionOrDash(a.currentProject.Version), 9).TextColor(colInk4)
				ui.Spacer(c)
				a.linkMore(c, "打开文档", func() { a.openDoc(latest.ID) })
			})
			ui.Row(c).Gap(10).Padding(4, 14, 10, 14).AlignItems(ui.Start).Children(func() {
				ui.Box(c).Width(2).FillHeight().Background(colSeal)
				ui.Column(c).Gap(4).Children(func() {
					ui.Text(c, latest.Title).FontSize(12.5).FontWeight(600)
					ui.Text(c, summary).FontSize(10.5).TextColor(colInk2)
				})
			})
			hairline(c)
			ui.Row(c).Height(26).Padding(0, 14).AlignItems(ui.Center).Children(func() {
				if latest.FileName != "" {
					monoText(c, strings.ToUpper(fileKindLabel(latest)), 9).TextColor(colInk4)
					ui.Box(c).Width(10)
					monoText(c, byteSize(latest.FileSize), 9).TextColor(colInk4)
				} else {
					monoText(c, "MARKDOWN", 9).TextColor(colInk4)
					ui.Box(c).Width(10)
					monoText(c, fmt.Sprintf("%d 字", latest.Words), 9).TextColor(colInk4)
				}
				ui.Box(c).Width(10)
				monoText(c, "更新于 "+model.Relative(latest.UpdatedAt), 9).TextColor(colInk4)
			})
		})
}

// requestDocBody loads a document's markdown in the background for the
// preview card, once per document.
func (a *App) requestDocBody(id int64) {
	if a.previewDocID == id {
		return
	}
	a.previewDocID = id
	a.previewBody = ""
	if a.svc == nil || a.win == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		d, err := a.svc.Document(ctx, id)
		if err == nil && a.previewDocID == id {
			a.win.Update(func() { a.previewBody = d.Content })
		}
	}()
}

// docSummary flattens the markdown into a short plain-text summary.
func docSummary(content string, lines int) string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimLeft(trimmed, "#>-* ")
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
		if len(out) >= lines {
			break
		}
	}
	joined := strings.Join(out, " ")
	runes := []rune(joined)
	if len(runes) > 120 {
		joined = string(runes[:120]) + "…"
	}
	return joined
}

// recordCard is the right column: the project's facts behind one hairline,
// not inside a card.
func (a *App) recordCard(c *ui.Context, p model.Project, width float32) {
	docs := a.projectDocs()
	words := 0
	for _, d := range docs {
		words += d.Words
	}
	ui.Column(c).Width(width).Background(colSheet2).
		BorderWidth(0, 0, 0, 1).BorderColor(colLine).
		Padding(14, 18, 14, 18).Gap(2).Children(func() {
			ui.Row(c).Height(30).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "记录卡").FontSize(12.5).FontWeight(600)
				ui.Box(c).Width(7)
				monoText(c, fmt.Sprintf("RECORD %04d", p.ID), 9).TextColor(colInk4)
				ui.Spacer(c)
				ui.Row(c).Height(18).Padding(0, 6).Radius(2).
					Border(1, colSeal).AlignItems(ui.Center).Children(func() {
					monoText(c, "LOCAL", 8.5).TextColor(colSeal)
				})
			})
			a.recordRow(c, "所属项目组", func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(7, 7).Radius(1).Background(groupColor(a, p.GroupID))
					ui.Text(c, groupOrUngrouped(a, p)).FontSize(11).FontWeight(600)
					ui.Spacer(c)
					b := ui.Row(c).Height(20).Padding(0, 7).Radius(2).Gap(4).AlignItems(ui.Center).
						Focusable().Role(ui.RoleButton).Label("移入其他组")
					if b.Hovered() {
						b.Background(colSheet)
					}
					b.OnClick(func() { a.openMoveProject(p.ID) })
					b.Children(func() {
						ui.Icon(c, icons.chevronRight).FontSize(10).TextColor(colInk4)
						ui.Text(c, "移动").FontSize(9.5).TextColor(colInk3)
					})
				})
			})
			a.recordRow(c, "版本", func() { a.versionChip(c, p.Version) })
			a.recordRow(c, "项目路径", func() {
				monoText(c, pathOrNone(p.Path), 9.5).TextColor(colInk2).SingleLine()
			})
			a.recordRow(c, "创建时间", func() {
				monoText(c, model.DateTime(p.CreatedAt), 9.5).TextColor(colInk2)
			})
			a.recordRow(c, "更新时间", func() {
				monoText(c, model.DateTime(p.UpdatedAt), 9.5).TextColor(colInk2)
			})
			a.recordRow(c, "标签", func() { tagsText(c, p.Tags, 0) })
			a.recordRow(c, "待办", func() {
				todos := a.projectTodos(p.ID)
				done := len(todos) - a.undoneTodos(p.ID)
				ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
					ui.Text(c, fmt.Sprintf("%d 项未完成 · %d 项已完成", a.undoneTodos(p.ID), done)).
						FontSize(10.5).TextColor(colInk2)
				})
			})
			hairline(c)
			a.noteSection(c, p)
			hairline(c)
			a.commentSection(c, p)
			hairline(c)
			ui.Column(c).Padding(10, 0, 0, 0).Gap(6).Children(func() {
				microLabel(c, "文档分布 · " + itoaPad(len(docs)))
				a.docDistribution(c, docs)
			})
			ui.Spacer(c)
			hairline(c)
			ui.Row(c).Padding(10, 0, 0, 0).Children(func() {
				ui.Column(c).Grow(1).Gap(1).Children(func() {
					monoText(c, itoa(len(docs)), 22).FontWeight(600)
					ui.Text(c, "项目文档").FontSize(9.5).TextColor(colInk3)
				})
				ui.Box(c).Width(1).Height(34).Background(colLine2).Margin(0, 14)
				ui.Column(c).Grow(1).Gap(1).Children(func() {
					monoText(c, wordsCount(words), 22).FontWeight(600)
					ui.Text(c, "文档总字数").FontSize(9.5).TextColor(colInk3)
				})
			})
		})
}

// recordRow is one label/value line of the record card.
func (a *App) recordRow(c *ui.Context, label string, value func()) {
	ui.Row(c).MinHeight(26).Padding(4, 0).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Width(66).Children(func() {
			ui.Text(c, label).FontSize(10).TextColor(colInk3)
		})
		ui.Column(c).Grow(1).Children(value)
	})
}

// docDistribution is the per-kind bar ledger inside the record card.
func (a *App) docDistribution(c *ui.Context, docs []model.Document) {
	total := len(docs)
	if total == 0 {
		ui.Text(c, "暂无文档").FontSize(10).TextColor(colInk4)
		return
	}
	counts := map[string]int{}
	var order []string
	for _, d := range docs {
		k := model.KindLabel(d.Kind)
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	for _, k := range order {
		n := counts[k]
		ui.Row(c).Height(20).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(5, 5).Radius(1).Background(kindColor(kindOf(k)))
			ui.Text(c, k).FontSize(10).TextColor(colInk2)
			ui.Spacer(c)
			ui.Box(c).Width(80).Height(2).Background(colLine2).Children(func() {
				ui.Box(c).Width(80 * float32(n) / float32(total)).Height(2).Background(colInk3)
			})
			ui.Box(c).Width(8)
			monoText(c, itoa(n), 9.5).TextColor(colInk3)
		})
	}
}

func kindOf(label string) string {
	switch label {
	case "需求":
		return "requirement"
	case "设计":
		return "design"
	case "开发计划":
		return "plan"
	case "接口":
		return "api"
	case "会议纪要":
		return "meeting"
	default:
		return "other"
	}
}

func wordsCount(words int) string {
	if words >= 1000 {
		return fmt.Sprintf("%.1fk", float64(words)/1000)
	}
	return itoa(words)
}

func groupColor(a *App, groupID *int64) ui.Color {
	if groupID == nil {
		return colInk4
	}
	if g, ok := a.groupByID(*groupID); ok {
		return ui.Hex(g.Color)
	}
	return colSeal
}
