package main

import (
	"context"
	"flag"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mana/internal/app"
)

func main() {
	demo := flag.Bool("demo", false, "seed sample data when the database is empty")
	flag.Parse()
	application := app.New()
	mygo.App.OnQuit(application.Close)
	mygo.App.WhenReady(func() {
		if err := application.Start(); err != nil {
			mygo.Dialog.Error("Mana 启动失败", err.Error())
			return
		}
		if *demo {
			if err := application.SeedDemo(context.Background()); err != nil {
				mygo.Dialog.Error("示例数据写入失败", err.Error())
				return
			}
			if err := application.Load(); err != nil {
				mygo.Dialog.Error("读取数据失败", err.Error())
				return
			}
		}
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:             "Mana · 开发项目管理",
			Width:             1440,
			Height:            900,
			MinWidth:          1120,
			MinHeight:         720,
			StateKey:          "mana-main",
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			TitleBarHeight:    36,
			BackgroundColor:   "#F5F7FA",
			Content:           ui.View(application.View),
		})
		application.AttachWindow(win)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
