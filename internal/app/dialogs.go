package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func (a *App) dialogs(c *ui.Context) {
	a.projectDialog(c)
	a.groupDialog(c)
	a.deleteDialog(c)
	a.discardDialog(c)
}

func (a *App) dialogBackdrop(backdrop ui.Element) {
	backdrop.Background(ui.RGBA(15, 23, 42, 0.42))
}

func (a *App) styleDialog(panel ui.Element, width float32) {
	panel.Width(width).Background(colSurface).Radius(6).Border(1, colBorder)
}

func (a *App) dialogHead(c *ui.Context, title string, close func()) {
	ui.Row(c).Height(40).Padding(0, 18).
		BorderWidth(0, 0, 1, 0).BorderColor(colBorderStrong).
		Children(func() {
			ui.Text(c, title).FontSize(15.5).Bold().TextColor(colPrimary).Grow(1)
			a.ghostButton(c, icons.close, "关闭", close)
		})
}

func (a *App) dialogFoot(c *ui.Context, cancelLabel, confirmLabel string, confirm func()) {
	ui.Row(c).Height(44).Padding(0, 18).
		BorderWidth(1, 0, 0, 0).BorderColor(colBorderStrong).
		Justify(ui.End).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Button(c, cancelLabel).Height(30).Padding(0, 14).OnClick(func() {
			a.showProjectDialog = false
			a.showGroupDialog = false
			a.showDeleteDialog = false
			a.showDiscardDialog = false
		})
		ui.PrimaryButton(c, confirmLabel).Height(30).Padding(0, 16).OnClick(confirm)
	})
}

func (a *App) projectDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showProjectDialog, func(backdrop, panel ui.Element) {
		a.dialogBackdrop(backdrop)
		a.styleDialog(panel, 640)
		title := "新建项目"
		confirm := "创建项目"
		if a.projectEditing != 0 {
			title = "编辑项目"
			confirm = "保存"
		}
		a.dialogHead(c, title, func() {
			a.showProjectDialog = false
		})
		ui.Form(c, func() {
			ui.Field(c, "项目名称 *", func() {
				ui.TextInput(c, &a.projectForm.Name).Placeholder("例如 Mana Desktop").AutoFocus()
			}).Error(a.fieldErr["name"])
			options := []string{"未分组"}
			for _, g := range a.groups {
				options = append(options, g.Name)
			}
			ui.Field(c, "所属项目组", func() {
				ui.Select(c, &a.groupChoice, options)
			}).Error(a.fieldErr["group"])
			ui.Field(c, "版本", func() {
				ui.TextInput(c, &a.projectForm.Version).Placeholder("例如 v1.0.0")
			}).Error(a.fieldErr["version"])
			ui.Field(c, "标签", func() {
				ui.TokenField(c, &a.projectForm.Tags, []string{"Go", "MyGo", "SQLite", "桌面端", "Flutter"})
			}).Error(a.fieldErr["tags"])
			ui.Field(c, "项目路径", func() {
				ui.Row(c).Gap(8).Children(func() {
					ui.TextInput(c, &a.projectForm.Path).Grow(1).Placeholder(`E:\codes\go_projects\mana_project`)
					a.secondaryButton(c, icons.folderOpen, "浏览", func() {
						a.chooseFolder()
					})
				})
			}).Error(a.fieldErr["path"])
			ui.Field(c, "备注", func() {
				ui.TextArea(c, &a.projectForm.Description).Height(64).Placeholder("目标、范围、注意事项…")
			}).Error(a.fieldErr["description"])
		}).Padding(14, 20, 20, 20)
		a.dialogFoot(c, "取消", confirm, func() { a.submitProject(c) })
	})
}

func (a *App) groupDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showGroupDialog, func(backdrop, panel ui.Element) {
		a.dialogBackdrop(backdrop)
		a.styleDialog(panel, 520)
		title := "新建项目组"
		confirm := "创建项目组"
		if a.groupEditing != 0 {
			title = "编辑项目组"
			confirm = "保存"
		}
		a.dialogHead(c, title, func() {
			a.showGroupDialog = false
		})
		ui.Form(c, func() {
			ui.Field(c, "项目组名称 *", func() {
				ui.TextInput(c, &a.groupForm.Name).Placeholder("例如 客户端应用").AutoFocus()
			}).Error(a.fieldErr["name"])
			ui.Field(c, "描述", func() {
				ui.TextArea(c, &a.groupForm.Description).Height(82).Placeholder("这个组负责什么…")
			}).Error(a.fieldErr["description"])
			ui.Field(c, "标识色", func() {
				ui.Row(c).Gap(10).Children(func() {
					for _, col := range []string{"#1677FF", "#52C41A", "#722ED1", "#FA8C16", "#13C2C2", "#EB2F96"} {
						color := col
						swatch := ui.Box(c).Size(28, 28).Radius(14).Background(ui.Hex(color)).
							Center().Focusable().Role(ui.RoleButton).Label("标识色 " + color)
						if a.groupForm.Color == color {
							swatch.Border(3, colSurface).InsetShadow(0, 0, 2, 4, ui.Hex(color))
							swatch.Children(func() {
								ui.Icon(c, icons.check).FontSize(13).TextColor(colSurface)
							})
						}
						swatch.OnClick(func() { a.groupForm.Color = color })
					}
				})
			}).Error(a.fieldErr["color"])
		}).Padding(14, 20, 20, 20)
		a.dialogFoot(c, "取消", confirm, func() { a.submitGroup(c) })
	})
}

func (a *App) deleteDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showDeleteDialog, func(backdrop, panel ui.Element) {
		a.dialogBackdrop(backdrop)
		a.styleDialog(panel, 448)
		ui.Column(c).Padding(18, 20, 20, 20).Children(func() {
			ui.Row(c).Gap(14).AlignItems(ui.Start).Children(func() {
				ui.Box(c).Size(42, 42).Radius(21).Background(colDanger1).Center().Children(func() {
					ui.Icon(c, icons.warning).FontSize(21).TextColor(colDanger)
				})
				ui.Column(c).Grow(1).Gap(5).Children(func() {
					ui.Text(c, a.deleteTitle).FontSize(15.5).Bold()
					ui.Text(c, a.deleteMessage).FontSize(11.5).TextColor(colText2)
					ui.Box(c).Background(colSurface2).Border(1, colBorder2).Radius(7).Padding(8, 10).Margin(4, 0, 0, 0).Children(func() {
						ui.Text(c, a.deleteImpact()).FontSize(11).TextColor(colText2)
					})
				})
			})
		})
		ui.Row(c).Height(44).Padding(0, 18).BorderWidth(1, 0, 0, 0).BorderColor(colBorderStrong).
			Justify(ui.End).Gap(8).Children(func() {
			ui.Button(c, "取消").Height(30).Padding(0, 14).OnClick(func() { a.showDeleteDialog = false })
			ui.Button(c, "确认删除").Height(30).Padding(0, 16).Background(colDanger).TextColor(colSurface).
				OnClick(func() { a.confirmDelete(c) })
		})
	})
}

func (a *App) deleteImpact() string {
	switch a.deleteKind {
	case "group":
		return "项目会移至未分组，项目文档不受影响。"
	default:
		return "此操作不可恢复。"
	}
}

func (a *App) discardDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showDiscardDialog, func(backdrop, panel ui.Element) {
		a.dialogBackdrop(backdrop)
		a.styleDialog(panel, 420)
		ui.Column(c).Padding(18, 20, 20, 20).Children(func() {
			ui.Text(c, "放弃未保存的修改？").FontSize(16).Bold()
			ui.Text(c, "当前文档有未保存的修改，返回后将丢失。").FontSize(11.5).TextColor(colText3).Padding(5, 0, 0, 0)
		})
		ui.Row(c).Height(44).Padding(0, 18).BorderWidth(1, 0, 0, 0).BorderColor(colBorderStrong).
			Justify(ui.End).Gap(8).Children(func() {
			ui.Button(c, "继续编辑").Height(30).Padding(0, 14).OnClick(func() { a.showDiscardDialog = false })
			ui.Button(c, "放弃修改").Height(30).Padding(0, 16).Background(colDanger).TextColor(colSurface).OnClick(func() {
				a.showDiscardDialog = false
				a.docForm = a.docInitial
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
