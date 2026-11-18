package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openSettings shows the data sheet and refreshes the database size off
// the frame.
func (a *App) openSettings() {
	a.showSettings = true
	a.dbSize = 0
	if a.dbPath == "" {
		return
	}
	if size, err := os.Stat(a.dbPath); err == nil {
		a.dbSize = size.Size()
	}
}

// chooseBackupFolder runs the directory picker, then writes a timestamped
// copy of the database into it.
func (a *App) chooseBackupFolder() {
	if a.win == nil || a.svc == nil {
		return
	}
	go func() {
		dirs, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      a.win,
			Title:       "选择备份位置",
			ButtonLabel: "备份到这里",
			Directory:   true,
		})
		if err != nil {
			a.win.Update(func() { a.notice = "打开目录选择器失败：" + err.Error() })
			return
		}
		if len(dirs) == 0 {
			return
		}
		dest := filepath.Join(dirs[0], "mana-backup-"+time.Now().Format("20060102-150405")+".db")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.svc.Backup(ctx, dest); err != nil {
			a.win.Update(func() { a.notice = err.Error() })
			return
		}
		a.win.Update(func() { a.notice = "已备份到 " + dest })
	}()
}

// revealDataFolder opens the folder holding mana.db in the file manager.
func (a *App) revealDataFolder() {
	if a.dbPath == "" {
		return
	}
	dir := filepath.Dir(a.dbPath)
	go func() {
		if err := mygo.Shell.OpenPath(dir); err != nil && a.win != nil {
			a.win.Update(func() { a.notice = "打开目录失败：" + err.Error() })
		}
	}()
}

// settingsDialog answers where the data lives and how to take it along.
func (a *App) settingsDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showSettings, func(backdrop, panel ui.Element) {
		a.dialogShell(c, backdrop, panel, 560,
			func() {
				a.dialogTop(c, "SETTINGS · 设置", "设置",
					"Mana 的数据都在本机；备份是一份可直接拷走的 SQLite 文件。",
					func() { a.showSettings = false })
			},
			func() {
				ui.Column(c).FillWidth().Padding(4, 20, 14, 20).Gap(0).Children(func() {
					microLabel(c, "数据 · DATA")
					ui.Box(c).Height(8)
					a.settingsRow(c, "数据文件", func() {
						path := a.dbPath
						if path == "" {
							path = "启动中…"
						}
						monoText(c, path, 9.5).TextColor(colInk2).Grow(1).SingleLine()
					}, "在文件管理器中显示", a.revealDataFolder)
					a.settingsRow(c, "数据库大小", func() {
						monoText(c, byteSize(a.dbSize), 10).TextColor(colInk2)
					}, "", nil)
					a.settingsRow(c, "备份", func() {
						ui.Text(c, "生成带时间戳的一致性副本，不影响当前使用。").FontSize(10).TextColor(colInk3)
					}, "立即备份", a.chooseBackupFolder)
					ui.Box(c).Height(10)
					hairline(c)
					ui.Box(c).Height(10)
					microLabel(c, "关于 · ABOUT")
					ui.Box(c).Height(8)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Size(18, 18).Radius(2).Background(colSeal).Center().Children(func() {
							monoText(c, "M", 10).FontWeight(700).TextColor(colPaper)
						})
						ui.Text(c, "Mana v1.0.0").FontSize(11.5).FontWeight(600)
						ui.Text(c, "Go + MyGo 原生 UI · 本地 SQLite · MIT").FontSize(10).TextColor(colInk3)
					})
				})
				a.dialogFoot(c, "ESC 关闭", "关闭", "完成", false, func() { a.showSettings = false })
			})
	})
}

// settingsRow is one label/value/action line of the settings sheet.
func (a *App) settingsRow(c *ui.Context, label string, value func(), action string, fn func()) {
	ui.Row(c).MinHeight(34).Padding(6, 0).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Width(72).Children(func() {
			ui.Text(c, label).FontSize(10.5).TextColor(colInk3)
		})
		ui.Column(c).Grow(1).Children(value)
		if action != "" && fn != nil {
			a.secondaryButton(c, nil, action, fn)
		}
	})
}

// byteSize renders a file size for the settings sheet and file viewers.
func byteSize(n int64) string {
	switch {
	case n <= 0:
		return "—"
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
