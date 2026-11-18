package app

import "github.com/egoist/mygo/ui"

var (
	colBg           = ui.Hex("#F4F5F7")
	colSurface      = ui.Hex("#FFFFFF")
	colSurface2     = ui.Hex("#FAFAFB")
	colBorder       = ui.Hex("#E3E5E8")
	colBorderStrong = ui.Hex("#D0D3D8")
	colBorder2      = ui.Hex("#EEF0F2")
	colText         = ui.Hex("#1D2129")
	colText2        = ui.Hex("#4E5969")
	colText3        = ui.Hex("#86909C")
	colText4        = ui.Hex("#A9AEB8")
	colPrimary      = ui.Hex("#7C5CFF")
	colPrimary1     = ui.Hex("#F3F0FF")
	colPrimary2     = ui.Hex("#DED6FE")
	colSuccess      = ui.Hex("#00B42A")
	colSuccess1     = ui.Hex("#E8FFEA")
	colWarning      = ui.Hex("#FF7D00")
	colWarn1        = ui.Hex("#FFF7E8")
	colDanger       = ui.Hex("#F53F3F")
	colDanger1      = ui.Hex("#FFECE8")
	colPurple       = ui.Hex("#7C5CFF")
	colPurple1      = ui.Hex("#F3F0FF")
	colCyan         = ui.Hex("#0FC6C2")
	colCyan1        = ui.Hex("#E8FFFB")
	colOrange       = ui.Hex("#F77234")
	colOrange1      = ui.Hex("#FFF2E8")
)

func newTheme() ui.Theme {
	t := *ui.LightTheme()
	t.Background = colBg
	t.Surface = colSurface
	t.SurfaceHover = ui.Hex("#F8F6FF")
	t.SurfacePressed = ui.Hex("#EFEAFF")
	t.Border = colBorder
	t.Text = colText
	t.TextMuted = colText3
	t.Accent = colPrimary
	t.AccentHover = ui.Hex("#9478FF")
	t.AccentPressed = ui.Hex("#6847E8")
	t.AccentText = colSurface
	t.Danger = colDanger
	t.Warning = colWarning
	t.Success = colSuccess
	t.Focus = colPrimary
	t.Selection = colPrimary1
	t.Radius = 4
	t.Spacing = 4
	t.FontSize = 13
	t.ScrollbarWidth = 5
	return t
}
