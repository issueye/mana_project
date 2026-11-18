package app

import (
	"fmt"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"mana/internal/model"
)

func (a *App) editorPage(c *ui.Context) {
	ui.Column(c).Fill().Padding(14, 18, 16, 18).Gap(8).Children(func() {
		a.editorTop(c)
		a.editorTitle(c)
		a.editorCard(c)
	})
}

func (a *App) editorTop(c *ui.Context) {
	ui.Row(c).Height(28).Gap(6).Children(func() {
		b := ui.Box(c).Size(22, 22).Radius(4).Center().Focusable().Role(ui.RoleButton).Label("返回")
		b.OnClick(a.backFromEditor)
		b.Children(func() { ui.Icon(c, icons.arrowLeft).FontSize(12).TextColor(colText2) })
		ui.Text(c, "项目组管理").FontSize(11).TextColor(colText3)
		ui.Text(c, "/").FontSize(11.5).TextColor(colText4)
		ui.Text(c, a.editorOwnerGroup()).FontSize(11.5).TextColor(colText3)
		ui.Text(c, "/").FontSize(11.5).TextColor(colText4)
		ui.Text(c, a.editorOwnerName()).FontSize(11.5).TextColor(colText3)
		ui.Text(c, "/").FontSize(11.5).TextColor(colText4)
		ui.Text(c, "文档").FontSize(11).Bold()
	})
}

func (a *App) editorTitle(c *ui.Context) {
	ui.Row(c).Padding(4, 0, 6, 0).Gap(10).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Gap(4).Children(func() {
			ui.TextInputBase(c, &a.docForm.Title).FontSize(20).Placeholder("未命名文档").Bold().AutoFocus()
			if err := a.fieldErr["title"]; err != "" {
				ui.Text(c, err).FontSize(10.5).TextColor(colDanger)
			}
			ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
				ui.Select(c, &a.docKindChoice,
					[]string{"需求", "设计", "开发计划", "接口", "会议纪要", "其他"}).Width(96).Height(24)
				ui.Text(c, "·").FontSize(11).TextColor(colText4)
				ui.Text(c, a.editorOwnerName()).FontSize(11.5).TextColor(colText3)
				ui.Text(c, "·").FontSize(11).TextColor(colText4)
				if a.docFormID != 0 {
					ui.Text(c, "最后保存 "+model.Relative(a.currentDoc.UpdatedAt)).FontSize(11).TextColor(colText3)
				} else {
					ui.Text(c, "新文档").FontSize(11).TextColor(colText3)
				}
			})
			linkRow := ui.Row(c).Gap(5).AlignItems(ui.Center)
			linkRow.Children(func() {
				ui.Icon(c, icons.external).FontSize(11).TextColor(colText4)
				ui.TextInput(c, &a.docForm.Link).Width(320).Height(24).Placeholder("关联链接（可选）")
			})
			if err := a.fieldErr["link"]; err != "" {
				ui.Text(c, err).FontSize(10.5).TextColor(colDanger)
			}
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Segmented(c, &a.editorMode, "编辑", "预览").Height(30)
			a.primaryButton(c, icons.check, "保存", func() { a.saveDoc(c) })
			if a.docFormID != 0 {
				a.ghostButton(c, icons.trash, "删除文档", func() { a.askDeleteDoc(a.docFormID) })
			}
		})
	})
}

func (a *App) editorCard(c *ui.Context) {
	ui.Column(c).Grow(1).Background(colSurface).Border(1, colBorder).Radius(6).
		Children(func() {
			ui.Row(c).Height(34).Padding(0, 8).Gap(2).Background(colSurface2).
				BorderWidth(0, 0, 1, 0).BorderColor(colBorder2).Children(func() {
				a.toolButton(c, "H1")
				a.toolButton(c, "H2")
				ui.Box(c).Width(1).Height(16).Background(colBorder2).Margin(0, 5)
				a.toolButton(c, "B")
				a.toolButton(c, "I")
				a.toolButton(c, "</>")
				ui.Box(c).Width(1).Height(16).Background(colBorder2).Margin(0, 5)
				a.toolButton(c, "链接")
				a.toolButton(c, "列表")
				ui.Spacer(c)
				ui.Text(c, "Markdown 实时预览").FontSize(10).TextColor(colText3)
			})
			if a.editorMode == 1 {
				a.editorPreview(c)
			} else {
				ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
					a.editorSource(c)
					ui.Box(c).Width(1).FillHeight().Background(colBorder2)
					a.editorPreview(c)
				})
			}
			a.editorStatus(c)
		})
}

func (a *App) editorSource(c *ui.Context) {
	_, windowHeight := c.Size()
	sourceHeight := windowHeight - 366
	if sourceHeight < 180 {
		sourceHeight = 180
	}
	ui.Column(c).Grow(1).Height(sourceHeight+28).Justify(ui.Start).AlignItems(ui.Start).
		Padding(14, 0, 14, 14).Children(func() {
		ui.TextAreaBase(c, &a.docForm.Content).FillWidth().FontSize(12).Padding(8).Lines(1, 40).AlignSelf(ui.Start)
	})
}

func (a *App) editorPreview(c *ui.Context) {
	_, windowHeight := c.Size()
	sourceHeight := windowHeight - 366
	if sourceHeight < 180 {
		sourceHeight = 180
	}
	ui.Column(c).Grow(1).Height(sourceHeight+28).Padding(18, 22).Children(func() {
		if a.docForm.Content == "" {
			ui.Text(c, "在左侧输入 Markdown，右侧会实时预览。").FontSize(12).TextColor(colText3)
			return
		}
		ui.Scroll(c).Grow(1).Children(func() {
			markdown.View(c, a.docForm.Content, nil).FillWidth()
		})
	})
}

func (a *App) editorStatus(c *ui.Context) {
	words := utf8.RuneCountInString(a.docForm.Content)
	ui.Row(c).Height(28).Padding(0, 12).Gap(12).Background(colSurface2).
		BorderWidth(1, 0, 0, 0).BorderColor(colBorder2).Children(func() {
		ui.Text(c, "Markdown").FontSize(10).TextColor(colText3)
		ui.Text(c, "UTF-8").FontSize(10).TextColor(colText3)
		ui.Spacer(c)
		ui.Text(c, fmt.Sprintf("%d 字", words)).FontSize(10).TextColor(colText3)
		if a.docFormID == 0 {
			ui.Text(c, "未保存").FontSize(10).TextColor(colWarning)
		} else {
			ui.Text(c, "已保存 "+model.Relative(a.currentDoc.UpdatedAt)).FontSize(10).TextColor(colText3)
		}
		ui.Text(c, "Ctrl S 保存").FontSize(10).TextColor(colText4)
	})
}

func (a *App) toolButton(c *ui.Context, label string) {
	ui.Text(c, label).FontSize(10.5).TextColor(colText3).Padding(4, 6).Radius(4)
}

func (a *App) editorOwnerName() string {
	if a.currentDoc.OwnerName != "" {
		return a.currentDoc.OwnerName
	}
	switch a.docForm.OwnerType {
	case model.OwnerProject:
		if p, ok := a.projectByID(a.docForm.OwnerID); ok {
			return p.Name
		}
	case model.OwnerGroup:
		if g, ok := a.groupByID(a.docForm.OwnerID); ok {
			return g.Name
		}
	}
	return "未知归属"
}

func (a *App) editorOwnerGroup() string {
	switch a.docForm.OwnerType {
	case model.OwnerProject:
		if p, ok := a.projectByID(a.docForm.OwnerID); ok {
			return p.GroupName
		}
	case model.OwnerGroup:
		return "项目组"
	}
	return "项目组"
}
