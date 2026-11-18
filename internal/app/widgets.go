package app

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

func itoa(n int) string { return strconv.Itoa(n) }

func (a *App) card(c *ui.Context, padding float32, fn func()) ui.Element {
	t := c.Theme()
	e := ui.Column(c).
		Background(t.Surface).
		Border(1, colBorder).
		Radius(6).
		Padding(padding)
	e.Children(fn)
	return e
}

func (a *App) flatCard(c *ui.Context, padding float32, fn func()) ui.Element {
	e := ui.Column(c).
		Background(colSurface2).
		Border(1, colBorder2).
		Radius(6).
		Padding(padding)
	e.Children(fn)
	return e
}

func (a *App) sectionLabel(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(10).Bold().TextColor(colText4).LetterSpacing(0.8)
}

func (a *App) tag(c *ui.Context, label string, fg, bg ui.Color) ui.Element {
	return ui.Row(c).Height(18).Padding(0, 6).Radius(3).Background(bg).
		AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(10).TextColor(fg).SingleLine()
	})
}

func (a *App) kindTag(c *ui.Context, kind string) ui.Element {
	switch kind {
	case "requirement":
		return a.tag(c, "需求", colOrange, colOrange1)
	case "design":
		return a.tag(c, "设计", colCyan, colCyan1)
	case "plan":
		return a.tag(c, "开发计划", ui.Hex("#2F8A12"), colSuccess1)
	case "api":
		return a.tag(c, "接口", colPrimary, colPrimary1)
	case "meeting":
		return a.tag(c, "会议纪要", colText2, ui.Hex("#F4F5F7"))
	default:
		return a.tag(c, "其他", colText2, ui.Hex("#F4F5F7"))
	}
}

func (a *App) navItem(c *ui.Context, icon *ui.SVG, label string, count int, active bool, onClick func()) {
	row := ui.Row(c).Height(36).Padding(0, 8).Gap(8).Radius(4).
		Focusable().Role(ui.RoleButton).Label(label)
	if active {
		row.Background(colPrimary1)
	} else if row.Hovered() {
		row.Background(ui.Hex("#F5F7FA"))
	}
	row.OnClick(onClick)
	fg := colText2
	if active {
		fg = colPrimary
	}
	row.Children(func() {
		ui.Icon(c, icon).FontSize(16).TextColor(fg)
		labelEl := ui.Text(c, label).FontSize(12.5).TextColor(fg)
		if active {
			labelEl.Bold()
		}
		ui.Spacer(c)
		if count > 0 {
			ui.Text(c, itoa(count)).FontSize(10).TextColor(colText3)
		}
	})
}

func (a *App) ghostButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(30, 30).Radius(4).Center().
		Focusable().Role(ui.RoleButton).Label(label)
	if b.Hovered() {
		b.Background(ui.Hex("#F3F5F8"))
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(colText3)
	})
}

func (a *App) iconTextButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Button(c, "").Height(30).Padding(0, 10).Radius(4).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(13)
		ui.Text(c, label).FontSize(12).SingleLine()
	})
}

func (a *App) primaryButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.PrimaryButton(c, "").Height(30).Padding(0, 10).Radius(4).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(13).TextColor(colSurface)
		ui.Text(c, label).FontSize(12).TextColor(colSurface).SingleLine()
	})
}

func (a *App) secondaryButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Button(c, "").Height(30).Padding(0, 10).Radius(4).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(13).TextColor(colText2)
		ui.Text(c, label).FontSize(12).TextColor(colText2).SingleLine()
	})
}

func (a *App) statCard(c *ui.Context, icon *ui.SVG, label, value, meta string, tone ui.Color, toneBg ui.Color) ui.Element {
	e := a.card(c, 12, func() {
		ui.Row(c).Gap(10).Children(func() {
			ui.Box(c).Size(36, 36).Radius(8).Background(toneBg).Center().Children(func() {
				ui.Icon(c, icon).FontSize(18).TextColor(tone)
			})
			ui.Column(c).Grow(1).Gap(1).Children(func() {
				ui.Text(c, label).FontSize(11).TextColor(colText3)
				ui.Text(c, value).FontSize(21).Bold()
				ui.Text(c, meta).FontSize(10).TextColor(colText3).SingleLine()
			})
		})
	})
	return e
}

func (a *App) clickRow(c *ui.Context, height float32, fn func(), build func()) ui.Element {
	_ = fn
	e := ui.Row(c).Height(height).AlignItems(ui.Center)
	e.Children(build)
	return e
}

func (a *App) iconTile(c *ui.Context, icon *ui.SVG, size float32, fg, bg ui.Color) ui.Element {
	e := ui.Box(c).Size(size, size).Radius(size / 3.4).Background(bg).Center()
	e.Children(func() {
		ui.Icon(c, icon).FontSize(size * 0.52).TextColor(fg)
	})
	return e
}

func (a *App) versionChip(c *ui.Context, version string) ui.Element {
	if version == "" {
		version = "—"
	}
	e := ui.Row(c).Height(20).Padding(0, 6).Radius(3).
		Border(1, colBorder).Background(colSurface2).AlignItems(ui.Center)
	e.Children(func() {
		ui.Text(c, version).FontSize(10).TextColor(colText2).SingleLine()
	})
	return e
}

func (a *App) emptyState(c *ui.Context, icon *ui.SVG, title, subtitle, action string, fn func()) {
	ui.Column(c).Grow(1).Center().Gap(8).Children(func() {
		ui.Box(c).Size(44, 44).Radius(8).Background(colSurface2).Border(1, colBorder2).Center().Children(func() {
			ui.Icon(c, icon).FontSize(20).TextColor(colText3)
		})
		ui.Text(c, title).FontSize(12.5).Bold().TextColor(colText2)
		ui.Text(c, subtitle).FontSize(11).TextColor(colText3).TextAlign(ui.Center)
		if action != "" {
			ui.PrimaryButton(c, action).Height(30).OnClick(fn)
		}
	})
}
