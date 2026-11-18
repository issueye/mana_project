package app

import "github.com/egoist/mygo/ui"

// The Index palette: warm paper, ink text and one cinnabar accent. Names
// follow design/README-v2.md §3.1; the legacy aliases below keep the rest
// of the package readable.
var (
	colPaper    = ui.Hex("#F0EBE0") // window canvas
	colSheet    = ui.Hex("#FCFAF5") // main content face
	colSheet2   = ui.Hex("#F6F1E6") // context panel, table heads, status bar
	colInk      = ui.Hex("#1B1913") // titles and body
	colInk2     = ui.Hex("#4A4436") // secondary text
	colInk3     = ui.Hex("#7B7464") // captions, times
	colInk4     = ui.Hex("#A39B87") // faint text, off icons
	colLine     = ui.Hex("#D6CFBE") // hairline
	colLine2    = ui.Hex("#E7E1D2") // fainter hairline
	colSeal     = ui.Hex("#AE4229") // selection, primary and danger actions
	colSealSoft = ui.Hex("#EFDFD7") // selected ground, danger ground
	colMoss     = ui.Hex("#4E6444")
	colAmber    = ui.Hex("#8C6227")
	colIndigo   = ui.Hex("#33586B")

	// On-ink colors for the dark window bar and icon rail.
	colBarText  = ui.Hex("#F0EBE0")
	colBarMuted = ui.Hex("#A39B87")
	colBarHover = ui.Hex("#2B2620")

	// Legacy aliases.
	colBg           = colPaper
	colSurface      = colSheet
	colSurface2     = colSheet2
	colBorder       = colLine
	colBorderStrong = ui.Hex("#C8BFA9")
	colBorder2      = colLine2
	colText         = colInk
	colText2        = colInk2
	colText3        = colInk3
	colText4        = colInk4
	colPrimary      = colSeal
	colPrimary1     = colSealSoft
	colPrimary2     = ui.Hex("#E2C9BB")
	colSuccess      = colMoss
	colSuccess1     = ui.Hex("#E7E9DD")
	colWarning      = colAmber
	colWarn1        = ui.Hex("#F1E7D3")
	colDanger       = colSeal
	colDanger1      = colSealSoft
)

// fontMono names the monospace stack for paths, versions, numbers and
// shortcuts: JetBrains Mono where installed, then the machine's choice.
const fontMono = "JetBrains Mono, Cascadia Mono, Consolas, monospace"

func newTheme() ui.Theme {
	t := *ui.LightTheme()
	t.Background = colPaper
	t.Surface = colSheet
	t.SurfaceHover = colSheet2
	t.SurfacePressed = colSealSoft
	t.Border = colLine
	t.Text = colInk
	t.TextMuted = colInk3
	t.Accent = colSeal
	t.AccentHover = ui.Hex("#C14F30")
	t.AccentPressed = ui.Hex("#93381F")
	t.AccentText = colSheet
	t.InactiveAccent = colSeal
	t.Danger = colSeal
	t.Warning = colAmber
	t.Success = colMoss
	t.Selection = colSealSoft
	t.Focus = colSeal
	t.Inverse = colInk
	t.InverseText = colPaper
	t.Radius = 2
	t.Spacing = 4
	t.FontSize = 12.5
	t.ScrollbarWidth = 5
	return t
}
