package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"mana/internal/extract"
	"mana/internal/model"
)

// uploadFilters is what the file picker offers for document uploads.
var uploadFilters = []mygo.FileFilter{
	{Name: "所有支持的文档", Extensions: []string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "md", "txt", "png", "jpg", "jpeg", "gif", "webp", "bmp"}},
	{Name: "PDF", Extensions: []string{"pdf"}},
	{Name: "Word", Extensions: []string{"doc", "docx"}},
	{Name: "Excel", Extensions: []string{"xls", "xlsx"}},
	{Name: "PPT", Extensions: []string{"ppt", "pptx"}},
	{Name: "图片", Extensions: []string{"png", "jpg", "jpeg", "gif", "webp", "bmp"}},
	{Name: "Markdown / 文本", Extensions: []string{"md", "txt"}},
	{Name: "所有文件", Extensions: []string{"*"}},
}

// uploadDocument opens the file picker and, on a choice, copies the file
// into the index off the frame.
func (a *App) uploadDocument(ownerType string, ownerID int64) {
	if a.win == nil || a.svc == nil {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      a.win,
			Title:       "上传文档",
			ButtonLabel: "上传",
			Filters:     uploadFilters,
		})
		if err != nil {
			a.win.Update(func() { a.notice = "打开文件选择器失败：" + err.Error() })
			return
		}
		if len(paths) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		d, err := a.svc.SaveDocumentFile(ctx, ownerType, ownerID, paths[0])
		if err != nil {
			a.win.Update(func() { a.notice = err.Error() })
			return
		}
		a.win.Update(func() {
			a.reload()
			a.notice = "已上传 " + d.FileName
		})
	}()
}

// viewerFile is the file document the editor is showing, with its viewer
// state: extracted sheets, the decoded image, the loaded text. The chosen
// sheet lives in App.viewerSheet by name.
type viewerFile struct {
	docID     int64
	kind      extract.Kind
	path      string
	text      string
	textReady bool
	textErr   string
	sheets    []extract.Sheet
	image     ui.ImageSource
	imageErr  string
}

// beginView resets the viewer for a document and starts its background
// load: extraction for documents, decoding for pictures.
func (a *App) beginView(d model.Document) {
	v := viewerFile{docID: d.ID, kind: extract.KindFor(d.FileName)}
	a.viewer = v
	a.viewerSheet = ""
	if d.FileName == "" {
		return
	}
	path, err := a.svc.DocumentFilePath(d)
	if err != nil {
		v.textErr = err.Error()
		v.textReady = true
		a.viewer = v
		return
	}
	v.path = path
	if v.kind == extract.KindImage {
		if info, statErr := os.Stat(path); statErr != nil || info.Size() > 40<<20 {
			v.imageErr = "图片过大或不可读，请用系统程序打开"
			v.textReady = true
			a.viewer = v
			return
		}
		if a.win == nil {
			// Headless tests: decode inline.
			a.decodeViewerImage(v, path)
			return
		}
		go func() {
			data, readErr := os.ReadFile(path)
			var src ui.ImageSource
			var decErr error
			if readErr == nil {
				src, decErr = ui.DecodeImage(data)
			} else {
				decErr = readErr
			}
			a.win.Update(func() {
				if a.viewer.docID != d.ID {
					return
				}
				if decErr != nil {
					a.viewer.imageErr = "图片解码失败：" + decErr.Error()
				} else {
					a.viewer.image = src
				}
			})
		}()
		return
	}
	if a.win == nil {
		a.loadViewerText(v, path)
		return
	}
	go func() {
		a.loadViewerText(v, path)
		a.win.Update(func() {
			if a.viewer.docID == d.ID {
				a.viewer = v
			}
		})
	}()
}

// loadViewerText extracts the preview for the document kinds that read
// as text, in the calling goroutine; the caller hands the result back to
// the UI thread.
func (a *App) loadViewerText(v viewerFile, path string) {
	if v.kind.NeedsSheets() {
		sheets, err := extract.Sheets(path)
		if err != nil {
			v.textErr = err.Error()
		} else {
			v.sheets = sheets
		}
		v.textReady = true
		a.viewer = v
		return
	}
	text, err := extract.Text(path)
	if err != nil {
		v.textErr = err.Error()
	} else {
		v.text = text
	}
	v.textReady = true
	a.viewer = v
}

func (a *App) decodeViewerImage(v viewerFile, path string) {
	data, err := os.ReadFile(path)
	if err == nil {
		var src ui.ImageSource
		src, err = ui.DecodeImage(data)
		if err == nil {
			v.image = src
		}
	}
	if err != nil {
		v.imageErr = "图片解码失败：" + err.Error()
	}
	v.textReady = true
	a.viewer = v
}

// fileKindLabel is the chip text for a document's attachment.
func fileKindLabel(d model.Document) string {
	if d.FileName == "" {
		return ""
	}
	return extract.KindFor(d.FileName).Label()
}

// renderFileView draws the attachment body inside the paper: a picture,
// a sheet grid, the markdown render or the extracted reading view.
func (a *App) renderFileView(c *ui.Context, d model.Document) {
	v := a.viewer
	if v.docID != d.ID {
		a.beginView(d)
		v = a.viewer
	}
	ui.Column(c).Grow(1).Padding(18, 22, 18, 22).Children(func() {
		switch {
		case v.kind == extract.KindImage && v.image != nil:
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Image(c, v.image)
			})
		case v.kind.NeedsSheets() && len(v.sheets) > 0:
			a.renderSheetGrid(c)
		case v.kind == extract.KindMarkdown && v.text != "":
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).FillWidth().Children(func() {
					markdown.View(c, v.text, nil).FillWidth()
				})
			})
		case v.text != "":
			a.renderReadingView(c, v.text)
		case !v.textReady:
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Text(c, "正在读取文件…").FontSize(11.5).TextColor(colInk3)
			})
		default:
			// Nothing extractable: the file card with a hint.
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, icons.file).FontSize(28).TextColor(colInk4)
					ui.Text(c, d.FileName).FontSize(12.5).FontWeight(600)
					ui.Text(c, fileMeta(d)).FontSize(10.5).TextColor(colInk3)
					note := "该文件没有可预览的内容，请用系统程序打开。"
					if v.textErr != "" {
						note = "预览提取失败：" + v.textErr
					}
					ui.Text(c, note).FontSize(10).TextColor(colInk4).TextAlign(ui.Center)
				})
			})
		}
	})
}

// renderReadingView shows extracted text with its page markers.
func (a *App) renderReadingView(c *ui.Context, text string) {
	lines := strings.Split(text, "\n")
	shown := lines
	if len(shown) > 4000 {
		shown = shown[:4000]
	}
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).FillWidth().Gap(2).Children(func() {
			for _, line := range shown {
				line := line
				trimmed := strings.TrimSpace(line)
				switch {
				case strings.HasPrefix(trimmed, "—— 第"):
					ui.Text(c, trimmed).FontSize(9.5).FontWeight(600).TextColor(colInk4).
						Padding(10, 0, 2, 0)
				case trimmed == "":
					ui.Box(c).Height(5)
				default:
					ui.Text(c, line).FontSize(12).TextColor(colInk)
				}
			}
			if len(shown) < len(lines) {
				ui.Text(c, "…（其余内容请用系统程序打开查看）").FontSize(10).TextColor(colInk4).
					Padding(8, 0, 0, 0)
			}
		})
	})
}

// renderSheetGrid draws the chosen worksheet as a table.
func (a *App) renderSheetGrid(c *ui.Context) {
	v := a.viewer
	if a.viewerSheet == "" && len(v.sheets) > 0 {
		a.viewerSheet = v.sheets[0].Name
	}
	sheetIdx := 0
	for i, s := range v.sheets {
		if s.Name == a.viewerSheet {
			sheetIdx = i
			break
		}
	}
	sheet := v.sheets[sheetIdx]
	ui.Row(c).Height(30).Gap(8).AlignItems(ui.Center).Children(func() {
		names := make([]string, len(v.sheets))
		for i, s := range v.sheets {
			names[i] = s.Name
		}
		ui.Select(c, &a.viewerSheet, names).Width(180).Height(28)
		ui.Spacer(c)
		monoText(c, fmt.Sprintf("%d 行", len(sheet.Rows)), 9.5).TextColor(colInk4)
	})
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Gap(0).Children(func() {
			limit := len(sheet.Rows)
			if limit > 200 {
				limit = 200
			}
			for i := 0; i < limit; i++ {
				row := sheet.Rows[i]
				head := i == 0
				r := ui.Row(c).MinHeight(26).Gap(0)
				if head {
					r.Background(colSheet2)
				} else if i%2 == 1 {
					r.Background(colPaper)
				}
				r.Children(func() {
					for j, cell := range row {
						if j >= 8 {
							break
						}
						ui.Box(c).Width(110).Padding(4, 8).Children(func() {
							ui.Text(c, cell).FontSize(10.5).FontWeight(map[bool]int{true: 600, false: 400}[head]).
								TextColor(colInk).SingleLine()
						})
					}
				})
			}
			if len(sheet.Rows) > 200 {
				ui.Text(c, fmt.Sprintf("仅显示前 200 行（共 %d 行），完整数据请用系统程序打开。", len(sheet.Rows))).
					FontSize(9.5).TextColor(colInk4).Padding(8, 0)
			}
		})
	})
}

// filePaper is the sheet that hosts a file document: a header of file
// facts, then the viewer body.
func (a *App) filePaper(c *ui.Context) {
	d := a.currentDoc
	ui.Column(c).Grow(1).Margin(0, 24, 12, 24).Background(colSheet).
		Border(1, colLine).Radius(2).Children(func() {
			ui.Row(c).Height(28).Padding(0, 12).Background(colSheet2).
				BorderWidth(0, 0, 1, 0).BorderColor(colLine2).AlignItems(ui.Center).Children(func() {
				monoText(c, strings.ToUpper(fileKindLabel(d)), 9).TextColor(colInk3)
				ui.Box(c).Width(1).Height(10).Background(colLine).Margin(0, 8)
				monoText(c, d.FileName, 9).TextColor(colInk3).SingleLine()
				ui.Spacer(c)
				monoText(c, byteSize(d.FileSize), 9).TextColor(colInk3)
				ui.Box(c).Width(10)
				monoText(c, "用原程序打开请点右上角按钮", 9).TextColor(colInk4)
			})
			a.renderFileView(c, d)
		})
}

// fileMeta is the one-line size/type caption under a file title.
func fileMeta(d model.Document) string {
	return fmt.Sprintf("%s · %s", d.FileName, byteSize(d.FileSize))
}

// openAttachedFile hands the stored copy to the system's default program.
func (a *App) openAttachedFile(c *ui.Context, d model.Document) {
	if d.FileName == "" {
		return
	}
	path, err := a.svc.DocumentFilePath(d)
	if err != nil {
		c.Toast("附件缺失，无法打开")
		return
	}
	go func() {
		if err := mygo.Shell.OpenPath(path); err != nil && a.win != nil {
			a.win.Update(func() { a.notice = "打开文件失败：" + err.Error() })
		}
	}()
}
