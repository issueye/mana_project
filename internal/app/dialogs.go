package app

import (
	"fmt"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func (a *App) dialogs(c *ui.Context) {
	a.projectDialog(c)
	a.groupDialog(c)
	a.deleteDialog(c)
	a.discardDialog(c)
	a.moveDialog(c)
	a.settingsDialog(c)
	a.paletteDialog(c)
}

// dialogShell wraps a panel in the index look: a dimming ink veil, a sheet
// with a 2px radius, one soft shadow and a 3px seal bar on top.
func (a *App) dialogShell(c *ui.Context, backdrop, panel ui.Element, width float32, top func(), body func()) {
	backdrop.Background(ui.RGBA(23, 21, 15, 0.56))
	panel.Width(width).Background(colSheet).Radius(2).
		Border(1, colLine).
		Shadow(0, 18, 40, 0, ui.RGBA(23, 21, 15, 0.28))
	ui.Column(c).FillWidth().Children(func() {
		ui.Box(c).Height(3).FillWidth().Background(colSeal)
		top()
		body()
	})
}

// dialogTop is the head of a record dialog: the micro kicker, the title,
// the caption and the close affordance.
func (a *App) dialogTop(c *ui.Context, kicker, title, caption string, close func()) {
	ui.Column(c).FillWidth().Padding(14, 20, 12, 20).Gap(3).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, kicker).FontSize(9).FontWeight(600).TextColor(colSeal).LetterSpacing(1.5)
			ui.Spacer(c)
			b := ui.Box(c).Size(24, 24).Radius(2).Center().
				Focusable().Role(ui.RoleButton).Label("关闭").Tooltip("关闭")
			if b.Hovered() {
				b.Background(colSheet2)
			}
			b.OnClick(close)
			b.Children(func() { ui.Icon(c, icons.close).FontSize(13).TextColor(colInk3) })
		})
		ui.Text(c, title).FontSize(20).FontWeight(700)
		ui.Text(c, caption).FontSize(10.5).TextColor(colInk3)
	})
}

// dialogFoot is the footer: the shortcuts on the left, the actions on the
// right. danger turns the confirm button seal.
func (a *App) dialogFoot(c *ui.Context, hint, cancelLabel, confirmLabel string, danger bool, confirm func()) {
	ui.Row(c).Height(48).Padding(0, 20).BorderWidth(1, 0, 0, 0).BorderColor(colLine2).
		AlignItems(ui.Center).Children(func() {
		monoText(c, hint, 9).TextColor(colInk4)
		ui.Spacer(c)
		ui.Button(c, cancelLabel).Height(30).Padding(0, 13).Radius(2).OnClick(func() {
			a.showProjectDialog = false
			a.showGroupDialog = false
			a.showDeleteDialog = false
			a.showDiscardDialog = false
		})
		ui.Box(c).Width(8)
		if danger {
			b := ui.Button(c, "").Height(30).Padding(0, 14).Radius(2).
				Background(colSeal).TextColor(colSheet).
				Focusable().Role(ui.RoleButton).Label(confirmLabel)
			b.OnClick(confirm)
			b.Children(func() {
				ui.Text(c, confirmLabel).FontSize(12).TextColor(colSheet)
			})
		} else {
			ui.PrimaryButton(c, confirmLabel).Height(30).Padding(0, 14).Radius(2).OnClick(confirm)
		}
	})
}

// projectDialog registers a local project.
func (a *App) projectDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showProjectDialog, func(backdrop, panel ui.Element) {
		title := "新建项目"
		confirm := "创建项目"
		kicker := "NEW RECORD · 项目"
		if a.projectEditing != 0 {
			title = "编辑项目"
			confirm = "保存修改"
			kicker = "EDIT RECORD · 项目"
		}
		a.dialogShell(c, backdrop, panel, 660,
			func() {
				a.dialogTop(c, kicker, title, "登记一个本地项目。路径可稍后补，名称与所属组决定它在索引中的位置。",
					func() { a.showProjectDialog = false })
			},
			func() {
				ui.Form(c, func() {
					ui.Row(c).AlignItems(ui.Center).Children(func() {
						ui.Field(c, "项目名称 *", func() {
							ui.TextInput(c, &a.projectForm.Name).Placeholder("例如 Mana Console").AutoFocus()
						}).Grow(1).Error(a.fieldErr["name"])
						ui.Box(c).Width(14)
						ui.Field(c, "版本", func() {
							ui.TextInput(c, &a.projectForm.Version).Placeholder("例如 v1.0.0")
						}).Width(180).Error(a.fieldErr["version"])
					})
					ui.Field(c, "所属项目组", func() {
						options := []string{"未分组"}
						for _, g := range a.groups {
							options = append(options, g.Name)
						}
						ui.Select(c, &a.groupChoice, options)
					}).Error(a.fieldErr["group"])
					ui.Field(c, "标签", func() {
						ui.TokenField(c, &a.projectForm.Tags, []string{"Go", "MyGo", "SQLite", "桌面端", "Flutter"})
					}).Error(a.fieldErr["tags"])
					ui.Field(c, "项目路径", func() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							ui.TextInput(c, &a.projectForm.Path).Grow(1).Placeholder(`E:\codes\go_projects\mana_console`)
							a.secondaryButton(c, icons.folderOpen, "浏览", a.chooseFolder)
						})
						hint := "登记路径后可直接「打开目录」。"
						hintColor := colInk4
						if a.projectForm.Path != "" {
							if _, err := os.Stat(a.projectForm.Path); err != nil {
								hint = "该路径不存在，保存时仅登记，不会创建目录。"
								hintColor = colAmber
							} else {
								hint = "路径有效，保存后可直接打开目录。"
							}
						}
						ui.Text(c, hint).FontSize(9.5).TextColor(hintColor).Padding(3, 0, 0, 0)
					}).Error(a.fieldErr["path"])
					ui.Field(c, "备注", func() {
						ui.TextArea(c, &a.projectForm.Description).Height(64).Placeholder("目标、范围、注意事项…")
					}).Error(a.fieldErr["description"])
				}).Padding(4, 20, 14, 20)
				a.dialogFoot(c, "CTRL ENTER 创建 · ESC 取消", "取消", confirm, false, func() { a.submitProject(c) })
			})
	})
}

// groupDialog creates or edits a project group.
func (a *App) groupDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showGroupDialog, func(backdrop, panel ui.Element) {
		title := "新建项目组"
		confirm := "创建项目组"
		kicker := "NEW GROUP · 项目组"
		if a.groupEditing != 0 {
			title = "编辑项目组"
			confirm = "保存修改"
			kicker = "EDIT GROUP · 项目组"
		}
		a.dialogShell(c, backdrop, panel, 520,
			func() {
				a.dialogTop(c, kicker, title, "项目组把相关项目收在同一页索引里，标识色用于目录与页眉。",
					func() { a.showGroupDialog = false })
			},
			func() {
				ui.Form(c, func() {
					ui.Field(c, "项目组名称 *", func() {
						ui.TextInput(c, &a.groupForm.Name).Placeholder("例如 客户端应用").AutoFocus()
					}).Error(a.fieldErr["name"])
					ui.Field(c, "描述", func() {
						ui.TextArea(c, &a.groupForm.Description).Height(72).Placeholder("这个组负责什么…")
					}).Error(a.fieldErr["description"])
					ui.Field(c, "标识色", func() {
						ui.Row(c).Gap(10).Children(func() {
							for _, col := range []string{"#33586B", "#4E6444", "#8C6227", "#AE4229", "#1B1913", "#A39B87"} {
								color := col
								swatch := ui.Box(c).Size(26, 26).Radius(2).Background(ui.Hex(color)).
									Center().Focusable().Role(ui.RoleButton).Label("标识色 " + color)
								if a.groupForm.Color == color {
									swatch.Border(1, colInk).InsetShadow(0, 0, 2, 2, ui.Hex(color))
									swatch.Children(func() {
										ui.Icon(c, icons.check).FontSize(13).TextColor(colSheet)
									})
								}
								swatch.OnClick(func() { a.groupForm.Color = color })
							}
						})
					}).Error(a.fieldErr["color"])
				}).Padding(4, 20, 14, 20)
				a.dialogFoot(c, "CTRL ENTER 创建 · ESC 取消", "取消", confirm, false, func() { a.submitGroup(c) })
			})
	})
}

// deleteDialog states the cascade in numbers before anything is removed.
func (a *App) deleteDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showDeleteDialog, func(backdrop, panel ui.Element) {
		a.dialogShell(c, backdrop, panel, 486,
			func() {
				ui.Row(c).Height(12)
			},
			func() {
				ui.Column(c).FillWidth().Padding(8, 20, 14, 20).Gap(10).Children(func() {
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Box(c).Size(38, 38).Radius(2).Background(colSealSoft).Center().Children(func() {
							ui.Icon(c, icons.warning).FontSize(19).TextColor(colSeal)
						})
						ui.Column(c).Grow(1).Gap(4).Children(func() {
							ui.Text(c, a.deleteTitle).FontSize(15).FontWeight(700)
							ui.Text(c, a.deleteMessage).FontSize(11).TextColor(colInk3)
						})
					})
					impact := a.deleteImpact()
					if len(impact) > 0 {
						ui.Row(c).Background(colSheet2).Border(1, colLine2).Radius(2).
							Padding(9, 12).Gap(10).AlignItems(ui.Center).Children(func() {
							ui.Box(c).Width(2).FillHeight().Background(colSeal)
							ui.Column(c).Gap(2).Children(func() {
								for _, line := range impact {
									ui.Text(c, line).FontSize(10.5).TextColor(colInk2)
								}
							})
						})
					}
				})
				a.dialogFoot(c, "ESC 取消 · ENTER 确认", "取消", "确认删除", true, func() { a.confirmDelete(c) })
			})
	})
}

// deleteImpact lists, line by line, what the deletion touches.
func (a *App) deleteImpact() []string {
	switch a.deleteKind {
	case "group":
		return []string{
			fmt.Sprintf("组内 %d 个项目将移至「未分组」。", a.pendingGroupProjects()),
			fmt.Sprintf("%d 篇组文档将一并删除，无法恢复。", len(a.groupDocs())),
		}
	case "project":
		return []string{
			fmt.Sprintf("项目下 %d 篇文档将一并删除，无法恢复。", len(a.projectDocs())),
		}
	default:
		return []string{"删除后无法恢复。"}
	}
}

// pendingGroupProjects counts the projects a group deletion moves out.
func (a *App) pendingGroupProjects() int {
	if g, ok := a.groupByID(a.deleteID); ok {
		return g.ProjectCount
	}
	return 0
}

// discardDialog keeps an unsaved document from disappearing silently.
func (a *App) discardDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showDiscardDialog, func(backdrop, panel ui.Element) {
		a.dialogShell(c, backdrop, panel, 420,
			func() {
				ui.Row(c).Height(12)
			},
			func() {
				ui.Column(c).FillWidth().Padding(8, 20, 14, 20).Gap(4).Children(func() {
					ui.Text(c, "放弃未保存的修改？").FontSize(15).FontWeight(700)
					ui.Text(c, "当前文档有未保存的修改，返回后将丢失。").FontSize(11).TextColor(colInk3)
				})
				a.dialogFoot(c, "ESC 取消 · ENTER 确认", "继续编辑", "放弃修改", true, func() {
					a.showDiscardDialog = false
					a.docForm = a.docInitial
					a.resetDocBuffer(a.docInitial.Content)
					a.leaveEditor()
				})
			})
	})
}

func (a *App) chooseFolder() {
	if a.win == nil {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      a.win,
			Title:       "选择项目目录",
			ButtonLabel: "选择",
			Directory:   true,
		})
		if err != nil {
			a.win.Update(func() { a.notice = "打开目录选择器失败：" + err.Error() })
			return
		}
		if len(paths) > 0 {
			a.win.Update(func() { a.projectForm.Path = paths[0] })
		}
	}()
}
