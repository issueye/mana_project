package app

import "github.com/egoist/mygo/ui"

type iconSet struct {
	home         *ui.SVG
	folder       *ui.SVG
	folderOpen   *ui.SVG
	layers       *ui.SVG
	grid         *ui.SVG
	plus         *ui.SVG
	search       *ui.SVG
	settings     *ui.SVG
	file         *ui.SVG
	note         *ui.SVG
	edit         *ui.SVG
	trash        *ui.SVG
	more         *ui.SVG
	chevronDown  *ui.SVG
	chevronRight *ui.SVG
	arrowLeft    *ui.SVG
	arrowRight   *ui.SVG
	close        *ui.SVG
	minimize     *ui.SVG
	maximize     *ui.SVG
	restore      *ui.SVG
	clock        *ui.SVG
	database     *ui.SVG
	code         *ui.SVG
	check        *ui.SVG
	checkCircle  *ui.SVG
	calendar     *ui.SVG
	filter       *ui.SVG
	bookmark     *ui.SVG
	user         *ui.SVG
	bell         *ui.SVG
	warning      *ui.SVG
	info         *ui.SVG
	gitBranch    *ui.SVG
	terminal     *ui.SVG
	copy         *ui.SVG
	refresh      *ui.SVG
	eye          *ui.SVG
	list         *ui.SVG
	undo         *ui.SVG
	external     *ui.SVG
}

func svg(body string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`))
}

var icons = iconSet{
	home:         svg(`<path d="M4 11l8-7 8 7"/><path d="M6.5 9.5V20h11V9.5"/>`),
	folder:       svg(`<path d="M4 7.5A1.5 1.5 0 015.5 6h3.2c.4 0 .8.2 1 .5L11 8h7.5A1.5 1.5 0 0120 9.5v8A1.5 1.5 0 0118.5 19h-13A1.5 1.5 0 014 17.5z"/>`),
	folderOpen:   svg(`<path d="M4 8V6.5A1.5 1.5 0 015.5 5h3.2c.4 0 .8.2 1 .5L11 7h6.5A1.5 1.5 0 0119 8.5V10"/><path d="M4 10h15.2a1 1 0 01.96 1.28l-1.6 5.4A1.5 1.5 0 0117.1 18H5.5A1.5 1.5 0 014 16.5z"/>`),
	layers:       svg(`<path d="M12 4l8 4.5-8 4.5-8-4.5z"/><path d="M4 13l8 4.5 8-4.5"/><path d="M4 17.5l8 4.5 8-4.5"/>`),
	grid:         svg(`<rect x="4" y="4" width="7" height="7" rx="1.5"/><rect x="13" y="4" width="7" height="7" rx="1.5"/><rect x="4" y="13" width="7" height="7" rx="1.5"/><rect x="13" y="13" width="7" height="7" rx="1.5"/>`),
	plus:         svg(`<path d="M12 5v14M5 12h14"/>`),
	search:       svg(`<circle cx="11" cy="11" r="6"/><path d="M20 20l-3.6-3.6"/>`),
	settings:     svg(`<path d="M4 8h9"/><path d="M17 8h3"/><path d="M4 16h3"/><path d="M11 16h9"/><circle cx="15" cy="8" r="2"/><circle cx="9" cy="16" r="2"/>`),
	file:         svg(`<path d="M6 4h8l4 4v12H6z"/><path d="M14 4v4h4"/><path d="M9 13h6"/><path d="M9 16h4"/>`),
	note:         svg(`<path d="M5 5h14v9l-5 5H5z"/><path d="M19 14h-5v5"/>`),
	edit:         svg(`<path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 013 3L7 19l-4 1 1-4z"/>`),
	trash:        svg(`<path d="M4 7h16"/><path d="M9.5 7V5h5v2"/><path d="M6.5 7l1 13h9l1-13"/>`),
	more:         svg(`<circle cx="6" cy="12" r="1.3" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1.3" fill="currentColor" stroke="none"/><circle cx="18" cy="12" r="1.3" fill="currentColor" stroke="none"/>`),
	chevronDown:  svg(`<path d="M5 9l7 7 7-7"/>`),
	chevronRight: svg(`<path d="M9 5l7 7-7 7"/>`),
	arrowLeft:    svg(`<path d="M19 12H5"/><path d="M11 6l-6 6 6 6"/>`),
	arrowRight:   svg(`<path d="M5 12h14"/><path d="M13 6l6 6-6 6"/>`),
	close:        svg(`<path d="M6 6l12 12M18 6L6 18"/>`),
	minimize:     svg(`<path d="M5 12h14"/>`),
	maximize:     svg(`<rect x="5" y="5" width="14" height="14" rx="1.5"/>`),
	restore:      svg(`<rect x="4" y="8" width="12" height="10" rx="1.5"/><path d="M8 8V5.5A1.5 1.5 0 019.5 4H18a2 2 0 012 2v9"/>`),
	clock:        svg(`<circle cx="12" cy="12" r="8"/><path d="M12 7v5l3 2"/>`),
	database:     svg(`<ellipse cx="12" cy="6" rx="7" ry="2.8"/><path d="M5 6v12c0 1.5 3.1 2.8 7 2.8s7-1.3 7-2.8V6"/><path d="M5 12c0 1.5 3.1 2.8 7 2.8s7-1.3 7-2.8"/>`),
	code:         svg(`<path d="M9 7l-4 5 4 5"/><path d="M15 7l4 5-4 5"/>`),
	check:        svg(`<path d="M4 12.5l5 5L20 6.5"/>`),
	checkCircle:  svg(`<circle cx="12" cy="12" r="8.5"/><path d="M8 12.2l2.8 2.8L16 9.5"/>`),
	calendar:     svg(`<rect x="4" y="6" width="16" height="14" rx="2"/><path d="M8 3v5"/><path d="M16 3v5"/><path d="M4 11h16"/>`),
	filter:       svg(`<path d="M20 4H4l6 7v6l4 2v-8z"/>`),
	bookmark:     svg(`<path d="M6 5h12v15l-6-4-6 4z"/>`),
	user:         svg(`<circle cx="12" cy="9" r="3.5"/><path d="M5 20c1.2-3.4 3.9-5 7-5s5.8 1.6 7 5"/>`),
	bell:         svg(`<path d="M8 9.5a4 4 0 018 0c0 3.6 1.4 5.2 1.4 5.2H6.6S8 13.1 8 9.5z"/><path d="M10.4 17.6a1.7 1.7 0 003.2 0"/>`),
	warning:      svg(`<path d="M12 4l8.5 15h-17z"/><path d="M12 10v4"/><path d="M12 16.6v.9"/>`),
	info:         svg(`<circle cx="12" cy="12" r="8.5"/><path d="M12 11v6"/><path d="M12 7.4v.9"/>`),
	gitBranch:    svg(`<circle cx="8" cy="6" r="2.4"/><circle cx="8" cy="18" r="2.4"/><circle cx="16" cy="10.5" r="2.4"/><path d="M8 8.4v7.2"/><path d="M16 12.9v1.4c0 1.9-1.6 3.3-3.7 3.3H10.4"/>`),
	terminal:     svg(`<rect x="3" y="5" width="18" height="14"/><path d="M7.5 10.4l2.8 2.4-2.8 2.4"/><path d="M13 15h3.8"/>`),
	copy:         svg(`<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>`),
	refresh:      svg(`<path d="M20 12a8 8 0 10-2.6 5.9"/><path d="M20 6.5V12h-5.5"/>`),
	eye:          svg(`<path d="M3 12c1.8-3.6 4.9-6 9-6s7.2 2.4 9 6c-1.8 3.6-4.9 6-9 6s-7.2-2.4-9-6z"/><circle cx="12" cy="12" r="3"/>`),
	list:         svg(`<path d="M4 7h16"/><path d="M4 12h16"/><path d="M4 17h16"/>`),
	undo:         svg(`<path d="M4 10h9a5 5 0 010 10H8"/><path d="M8 6l-4 4 4 4"/>`),
	external:     svg(`<path d="M7 17L17 7"/><path d="M9 7h8v8"/>`),
}
