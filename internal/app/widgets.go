package app

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
)

func itoa(n int) string { return strconv.Itoa(n) }

// monoText renders text in the monospace stack: paths, versions, numbers,
// counts and shortcuts.
func monoText(c *ui.Context, text string, size float32) ui.Element {
	return ui.Text(c, text).Font(fontMono).FontSize(size)
}

// microLabel is the 9.5px letterspaced caption that heads sections and
// panels; the caller upper-cases where the design wants it.
func microLabel(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(9.5).TextColor(colInk4).LetterSpacing(1.5)
}

// sectionTitle is the 13/20 600 heading of a section or table.
func sectionTitle(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(13).FontWeight(600)
}

func (a *App) card(c *ui.Context, padding float32, fn func()) ui.Element {
	e := ui.Column(c).
		Background(colSheet).
		Border(1, colLine).
		Radius(2).
		Padding(padding)
	e.Children(fn)
	return e
}

func (a *App) flatCard(c *ui.Context, padding float32, fn func()) ui.Element {
	e := ui.Column(c).
		Background(colSheet2).
		Border(1, colLine2).
		Radius(2).
		Padding(padding)
	e.Children(fn)
	return e
}

// kindColor maps a document kind to one of the low-saturation hues; only
// small dots use it.
func kindColor(kind string) ui.Color {
	switch kind {
	case "requirement":
		return colAmber
	case "design":
		return colIndigo
	case "plan":
		return colMoss
	case "api":
		return colSeal
	case "meeting":
		return colInk2
	default:
		return colInk4
	}
}

// kindDot is the small square that marks a document's type.
func kindDot(c *ui.Context, kind string) {
	ui.Box(c).Size(6, 6).Radius(1).Background(kindColor(kind))
}

// kindText is the plain type label used in table columns.
func kindText(c *ui.Context, kind string) {
	ui.Text(c, model.KindLabel(kind)).FontSize(10.5).TextColor(colInk3)
}

// kindTag is dot plus label, for document strips and lists.
func (a *App) kindTag(c *ui.Context, kind string) ui.Element {
	return ui.Row(c).AlignItems(ui.Center).Gap(5).Children(func() {
		kindDot(c, kind)
		ui.Text(c, model.KindLabel(kind)).FontSize(10).TextColor(colInk3).SingleLine()
	})
}

// versionChip is a monospace version in a hairline chip.
func (a *App) versionChip(c *ui.Context, version string) ui.Element {
	if version == "" {
		version = "—"
	}
	e := ui.Row(c).Height(20).Padding(0, 7).Radius(2).
		Border(1, colLine).Background(colSheet2).AlignItems(ui.Center)
	e.Children(func() {
		monoText(c, version, 10).TextColor(colInk2).SingleLine()
	})
	return e
}

// tagsText renders tags as words with middle dots — no colored pills.
func tagsText(c *ui.Context, tags []string, max int) {
	if len(tags) == 0 {
		ui.Text(c, "—").FontSize(10.5).TextColor(colInk4)
		return
	}
	joined := make([]string, 0, len(tags))
	for i, tag := range tags {
		if max > 0 && i >= max {
			break
		}
		joined = append(joined, tag)
	}
	text := strings.Join(joined, " · ")
	if max > 0 && len(tags) > max {
		text += " +" + itoa(len(tags)-max)
	}
	ui.Text(c, text).FontSize(10.5).TextColor(colInk2).SingleLine()
}

func (a *App) ghostButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(30, 30).Radius(2).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Hovered() {
		b.Background(colSheet2)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(colInk3)
	})
}

// iconButton is an icon-only action button: a bordered square whose meaning
// the accessible label and the tooltip tell, with no text inside.
func (a *App) iconButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(30, 30).Radius(2).Border(1, colLine).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Hovered() || b.Pressed() {
		b.Background(colSheet2)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(colInk2)
	})
}

// primaryIconButton is iconButton in the seal color, for the main action.
func (a *App) primaryIconButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	t := c.Theme()
	b := ui.Box(c).Size(30, 30).Radius(2).Background(t.Accent).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Pressed() {
		b.Background(t.AccentPressed)
	} else if b.Hovered() {
		b.Background(t.AccentHover)
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(t.AccentText)
	})
}

// dangerIconButton is iconButton for a destructive action: the icon and its
// ground turn seal while the pointer is on it.
func (a *App) dangerIconButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Box(c).Size(30, 30).Radius(2).Border(1, colLine).Center().
		Focusable().Role(ui.RoleButton).Label(label).Tooltip(label)
	fg := colInk3
	if b.Hovered() || b.Pressed() {
		b.Background(colSealSoft)
		fg = colSeal
	}
	b.OnClick(fn)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15).TextColor(fg)
	})
}

func (a *App) iconTextButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Button(c, "").Height(30).Padding(0, 10).Radius(2).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(13)
		}
		ui.Text(c, label).FontSize(12).SingleLine()
	})
}

func (a *App) primaryButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	t := c.Theme()
	b := ui.PrimaryButton(c, "").Height(30).Padding(0, 11).Radius(2).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(13).TextColor(t.AccentText)
		}
		ui.Text(c, label).FontSize(12).TextColor(t.AccentText).SingleLine()
	})
}

func (a *App) secondaryButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Button(c, "").Height(30).Padding(0, 11).Radius(2).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(13).TextColor(colInk2)
		}
		ui.Text(c, label).FontSize(12).TextColor(colInk2).SingleLine()
	})
}

// borderedButton is the quiet icon-plus-text action in a page header.
func (a *App) borderedButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Row(c).Height(30).Padding(0, 11).Radius(2).Gap(6).
		Border(1, colLine).Background(colSheet).AlignItems(ui.Center).
		Focusable().Role(ui.RoleButton).Label(label)
	if b.Hovered() || b.Pressed() {
		b.Background(colSheet2)
	}
	b.OnClick(fn)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(13).TextColor(colInk2)
		}
		ui.Text(c, label).FontSize(11.5).TextColor(colInk2).SingleLine()
	})
}

// dangerButton is the bordered button whose text and border turn seal.
func (a *App) dangerButton(c *ui.Context, icon *ui.SVG, label string, fn func()) {
	b := ui.Button(c, "").Height(30).Padding(0, 11).Radius(2).
		Focusable().Role(ui.RoleButton).Label(label)
	border, fg := colLine, colSeal
	if b.Hovered() || b.Pressed() {
		b.Background(colSealSoft)
		border = colSeal
	}
	b.Border(1, border)
	b.OnClick(fn)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(13).TextColor(fg)
		}
		ui.Text(c, label).FontSize(12).TextColor(fg).SingleLine()
	})
}

// metaItem is an icon plus caption used in page headers.
func (a *App) metaItem(c *ui.Context, icon *ui.SVG, text string) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, icon).FontSize(11).TextColor(colInk4)
		ui.Text(c, text).FontSize(10.5).TextColor(colInk3)
	})
}

// ledgerCell is one of the four bookkeeping stats on the home page: a
// micro label, a monospace number with its unit and a caption underneath.
func (a *App) ledgerCell(c *ui.Context, icon *ui.SVG, label, unit, value, meta string, first bool) {
	cell := ui.Column(c).Grow(1).Gap(3)
	if first {
		cell.Padding(0, 24, 4, 0)
	} else {
		cell.Padding(0, 24, 4, 24).BorderWidth(0, 0, 0, 1).BorderColor(colLine2)
	}
	cell.Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icon).FontSize(12).TextColor(colInk4)
			microLabel(c, label)
		})
		ui.Row(c).Gap(4).AlignItems(ui.Baseline).Children(func() {
			monoText(c, value, 30).FontWeight(600).TextColor(colInk)
			ui.Text(c, unit).FontSize(10.5).TextColor(colInk3)
		})
		ui.Text(c, meta).FontSize(10).TextColor(colInk3).SingleLine()
	})
}

// linkMore is the small "more" link that heads list sections.
func (a *App) linkMore(c *ui.Context, label string, fn func()) {
	b := ui.Button(c, "").Height(24).Padding(0, 4).Radius(2).
		Focusable().Role(ui.RoleButton).Label(label)
	b.OnClick(fn)
	b.Children(func() {
		ui.Text(c, label).FontSize(11).TextColor(colInk3)
		ui.Icon(c, icons.arrowRight).FontSize(11).TextColor(colInk4)
	})
}

// dashedButton is the quiet bordered action at panel bottoms.
func (a *App) dashedButton(c *ui.Context, label string, fn func()) {
	button := ui.Row(c).Height(32).Radius(2).Center().Gap(6).
		Border(1, colLine).BorderStyle(ui.BorderDashed).
		Focusable().Role(ui.RoleButton).Label(label)
	if button.Hovered() {
		button.Border(1, colInk4)
		button.Background(colSheet)
	}
	button.OnClick(fn)
	button.Children(func() {
		ui.Icon(c, icons.plus).FontSize(13).TextColor(colInk3)
		ui.Text(c, label).FontSize(11.5).TextColor(colInk3)
	})
}

// hairline is a full-width 1px rule.
func hairline(c *ui.Context) {
	ui.Box(c).Height(1).Background(colLine2)
}

// emptyState is the blank index card: a dashed record with the first
// actions on it. One or two action pairs may follow (label, fn).
func (a *App) emptyState(c *ui.Context, title, subtitle string, actions ...[2]any) {
	ui.Column(c).Grow(1).Center().Gap(10).Children(func() {
		cardEl := ui.Row(c).Width(300).Height(130).Background(colSheet).
			Border(1, colLine).Radius(2)
		cardEl.Children(func() {
			ui.Box(c).Width(3).FillHeight().Background(colSeal)
			ui.Column(c).Grow(1).FillHeight().Padding(16, 18).Justify(ui.Center).Gap(7).Children(func() {
				microLabel(c, "BLANK RECORD")
				ui.Box(c).Width(150).Height(1).Background(colLine2).Margin(4, 0)
				ui.Box(c).Width(100).Height(1).Background(colLine2)
				ui.Box(c).Width(190).Height(1).Background(colLine2)
			})
		})
		ui.Text(c, title).FontSize(13).FontWeight(600)
		ui.Text(c, subtitle).FontSize(11).TextColor(colInk3).TextAlign(ui.Center)
		if len(actions) > 0 {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				for i, act := range actions {
					label, _ := act[0].(string)
					fn, _ := act[1].(func())
					if label == "" || fn == nil {
						continue
					}
					if i == 0 {
						a.primaryButton(c, icons.plus, label, fn)
					} else {
						a.secondaryButton(c, nil, label, fn)
					}
				}
			})
		}
	})
}
