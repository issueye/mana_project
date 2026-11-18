package app

import (
	"fmt"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"

	"mana/internal/model"
)

// editorPage is the writing paper: a plain header with the title, the
// source with a line-number gutter on the left and the rendered preview on
// the right. A file document shows its viewer instead.
func (a *App) editorPage(c *ui.Context) {
	ui.Column(c).Fill().Children(func() {
		a.editorTop(c)
		if a.currentDoc.FileName != "" {
			a.filePaper(c)
		} else {
			a.editorPaper(c)
		}
	})
}

// editorTop is the header: breadcrumb, the editable title with its type,
// the ownership and save state, the link field and the actions.
func (a *App) editorTop(c *ui.Context) {
	ui.Column(c).Padding(16, 24, 10, 24).Gap(4).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "索引").FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, a.editorOwnerGroup()).FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, a.editorOwnerName()).FontSize(10).TextColor(colInk4)
			ui.Text(c, "/").FontSize(10).TextColor(colInk4)
			ui.Text(c, "文档 01").FontSize(10).FontWeight(600).TextColor(colSeal)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.TextInputBase(c, &a.docForm.Title).FontSize(22).FontWeight(700).
				Placeholder("未命名文档").Grow(1).AutoFocus()
			ui.Select(c, &a.docKindChoice,
				[]string{"需求", "设计", "开发计划", "接口", "会议纪要", "其他"}).Width(100).Height(28)
			if a.currentDoc.FileName == "" {
				ui.Segmented(c, &a.editorMode, "源码", "预览").Height(28)
			} else {
				a.iconButton(c, icons.folderOpen, "在系统中打开", func() {
					a.openAttachedFile(c, a.currentDoc)
				})
			}
			a.primaryButton(c, icons.check, "保存", func() { a.saveDoc(c) })
			if a.docFormID != 0 {
				a.ghostButton(c, icons.trash, "删除文档", func() { a.askDeleteDoc(a.docFormID) })
			}
		})
		if err := a.fieldErr["title"]; err != "" {
			ui.Text(c, err).FontSize(10).TextColor(colSeal)
		}
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "归属 "+a.editorOwnerName()).FontSize(10.5).TextColor(colInk3)
			a.metaDot(c)
			if a.currentDoc.FileName != "" {
				monoText(c, fileMeta(a.currentDoc), 10).TextColor(colInk2)
			} else if a.docFormID != 0 {
				ui.Text(c, "最后保存 "+model.Relative(a.currentDoc.UpdatedAt)).FontSize(10.5).TextColor(colInk3)
			} else {
				ui.Text(c, "新文档").FontSize(10.5).TextColor(colInk3)
			}
			a.metaDot(c)
			state := "未保存"
			stateColor := colSeal
			if a.currentDoc.FileName == "" && !a.editorDirty() && a.docFormID != 0 {
				state = "已保存"
				stateColor = colMoss
			} else if a.currentDoc.FileName != "" {
				state = ""
				stateColor = colMoss
			}
			if state != "" {
				ui.Text(c, state).FontSize(10.5).FontWeight(600).TextColor(stateColor)
			}
		})
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icons.external).FontSize(11).TextColor(colInk4)
			ui.TextInput(c, &a.docForm.Link).Width(420).Height(26).Placeholder("关联链接（可选）")
			if err := a.fieldErr["link"]; err != "" {
				ui.Text(c, err).FontSize(10).TextColor(colSeal)
			}
		})
	})
}

// editorPaper is the sheet: a thin header of format facts, then the source
// with its gutter and the preview side by side.
func (a *App) editorPaper(c *ui.Context) {
	words := utf8.RuneCountInString(a.docContent())
	ui.Column(c).Grow(1).Margin(0, 24, 12, 24).Background(colSheet).
		Border(1, colLine).Radius(2).Children(func() {
			ui.Row(c).Height(28).Padding(0, 12).Background(colSheet2).
				BorderWidth(0, 0, 1, 0).BorderColor(colLine2).AlignItems(ui.Center).Children(func() {
				monoText(c, "MARKDOWN", 9).TextColor(colInk3)
				ui.Box(c).Width(1).Height(10).Background(colLine).Margin(0, 8)
				monoText(c, "UTF-8", 9).TextColor(colInk3)
				ui.Spacer(c)
				monoText(c, fmt.Sprintf("%d 字", words), 9).TextColor(colInk3)
				ui.Box(c).Width(10)
				monoText(c, "CTRL S 保存", 9).TextColor(colInk4)
			})
			w, _ := c.Size()
			split := a.editorMode == 0 && w >= 1200
			ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
				if a.editorMode == 0 {
					a.editorSource(c)
				}
				if split {
					ui.Box(c).Width(1).FillHeight().Background(colLine2)
				}
				if a.editorMode == 1 || split {
					a.editorPreview(c)
				}
			})
		})
}

// editorSource builds the mono editor with the line-number gutter.
func (a *App) editorSource(c *ui.Context) {
	if a.docBuf == nil {
		a.docBuf = editor.NewBuffer(a.docForm.Content)
	}
	st := &a.editorState
	st.Buffer = a.docBuf
	st.Font = ui.Font{Family: fontMono, Size: 12}
	st.LineHeight = 1.6
	st.SoftWrap = true
	st.LineNumbers = editor.LineNumbersOn
	st.Placeholder = "在这里写下这篇文档…"
	st.Label = "Markdown 源码"
	t := editor.DefaultTheme(c.Theme())
	t.Background = colSheet
	t.Text = colInk
	t.LineNumber = colInk4
	t.Cursor = colSeal
	t.IndentGuide = colLine2
	t.WrapGuide = colLine2
	t.Tokens[editor.TokenHeading].Color = colSeal
	t.Tokens[editor.TokenCode].Color = colIndigo
	t.Tokens[editor.TokenLink].Color = colIndigo
	st.Theme = t
	editor.View(c, st).Grow(1)
}

func (a *App) editorPreview(c *ui.Context) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).FillWidth().Padding(18, 22, 18, 22).Children(func() {
			content := a.docContent()
			if content == "" {
				ui.Text(c, "在左侧写下 Markdown，这里会实时渲染。").FontSize(11.5).TextColor(colInk3)
				return
			}
			markdown.View(c, content, nil).FillWidth()
		})
	})
}

// docContent reads the source from the editor buffer, keeping the form in
// step with what the user typed.
func (a *App) docContent() string {
	if a.docBuf != nil {
		return a.docBuf.String()
	}
	return a.docForm.Content
}

// syncDocBuffer moves the buffer's text into the form before it is saved
// or compared.
func (a *App) syncDocBuffer() {
	if a.docBuf != nil {
		a.docForm.Content = a.docBuf.String()
	}
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
