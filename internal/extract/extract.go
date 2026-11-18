// Package extract pulls readable text out of uploaded documents with the
// standard library plus a pure-Go PDF reader: OOXML files unzip into XML,
// PDFs go through ledongthuc/pdf, everything else is read as text.
package extract

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	pdflib "github.com/ledongthuc/pdf"
)

// MaxText caps the preview text kept in the database.
const MaxText = 20000

// Kind classifies an attachment for the viewer.
type Kind string

const (
	KindPDF      Kind = "pdf"
	KindWord     Kind = "word"
	KindExcel    Kind = "excel"
	KindSlides   Kind = "ppt"
	KindImage    Kind = "image"
	KindMarkdown Kind = "markdown"
	KindText     Kind = "text"
	KindOther    Kind = "other"
)

// KindFor maps a file extension to a viewer kind.
func KindFor(name string) Kind {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")) {
	case "pdf":
		return KindPDF
	case "doc", "docx", "odt", "rtf":
		return KindWord
	case "xls", "xlsx", "ods":
		return KindExcel
	case "ppt", "pptx", "odp":
		return KindSlides
	case "png", "jpg", "jpeg", "gif", "webp", "bmp", "svg":
		return KindImage
	case "md", "markdown":
		return KindMarkdown
	case "txt", "log", "json", "yaml", "yml", "toml", "ini", "go", "js", "ts", "py", "sql":
		return KindText
	default:
		return KindOther
	}
}

// KindLabel is the Chinese chip label for a kind.
func (k Kind) Label() string {
	switch k {
	case KindPDF:
		return "PDF"
	case KindWord:
		return "Word"
	case KindExcel:
		return "Excel"
	case KindSlides:
		return "PPT"
	case KindImage:
		return "图片"
	case KindMarkdown:
		return "Markdown"
	case KindText:
		return "文本"
	default:
		return "文件"
	}
}

// NeedsSheets reports whether the kind previews as a spreadsheet.
func (k Kind) NeedsSheets() bool { return k == KindExcel }

// Text extracts a readable plain-text preview of the file. Every format
// degrades gracefully: an error means the viewer falls back to the file
// card and the system opener.
func Text(path string) (string, error) {
	switch KindFor(path) {
	case KindWord:
		return docxText(path)
	case KindSlides:
		return pptxText(path)
	case KindPDF:
		return pdfText(path)
	default:
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return capText(string(b)), nil
	}
}

// Sheet is one worksheet of a spreadsheet: its name and trimmed cells.
type Sheet struct {
	Name string
	Rows [][]string
}

// Sheets parses an xlsx file into per-sheet cell matrices.
func Sheets(path string) ([]Sheet, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	shared, err := readSharedStrings(&zr.Reader)
	if err != nil {
		return nil, err
	}
	sheets, err := readWorkbook(&zr.Reader, shared)
	if err != nil {
		return nil, err
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("extract: no worksheets in %s", filepath.Base(path))
	}
	return sheets, nil
}

// capText trims the preview to MaxText runes.
func capText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > MaxText {
		return string(runes[:MaxText]) + "\n\n…（预览已截断）"
	}
	return s
}

// docxText walks word/document.xml, one line per paragraph.
func docxText(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	f := findFile(&zr.Reader, "word/document.xml")
	if f == nil {
		return "", fmt.Errorf("extract: not a docx: %s", filepath.Base(path))
	}
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var sb strings.Builder
	para := strings.Builder{}
	inText := false
	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				para.Reset()
			case "t":
				inText = true
			case "tab":
				para.WriteByte('\t')
			case "br":
				para.WriteByte('\n')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				if line := strings.Trim(para.String(), "\n"); strings.TrimSpace(line) != "" {
					if sb.Len() > 0 {
						sb.WriteByte('\n')
					}
					sb.WriteString(line)
				}
			}
		case xml.CharData:
			if inText {
				para.Write(t)
			}
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("extract: no text found in %s", filepath.Base(path))
	}
	return capText(sb.String()), nil
}

// pptxText collects the text runs of every slide, in page order.
func pptxText(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	type slideFile struct {
		n    int
		name string
	}
	var slides []slideFile
	for _, f := range zr.Reader.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			mid := strings.TrimSuffix(strings.TrimPrefix(f.Name, "ppt/slides/slide"), ".xml")
			if n, err := strconv.Atoi(mid); err == nil {
				slides = append(slides, slideFile{n, f.Name})
			}
		}
	}
	if len(slides) == 0 {
		return "", fmt.Errorf("extract: not a pptx: %s", filepath.Base(path))
	}
	sort.Slice(slides, func(i, j int) bool { return slides[i].n < slides[j].n })

	var sb strings.Builder
	for i, s := range slides {
		f := findFile(&zr.Reader, s.name)
		if f == nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		var para strings.Builder
		var page strings.Builder
		inText := false
		dec := xml.NewDecoder(rc)
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "p" {
					para.Reset()
				} else if t.Name.Local == "t" {
					inText = true
				}
			case xml.EndElement:
				if t.Name.Local == "t" {
					inText = false
				} else if t.Name.Local == "p" {
					if line := strings.TrimSpace(para.String()); line != "" {
						page.WriteString(line)
						page.WriteByte('\n')
					}
				}
			case xml.CharData:
				if inText {
					para.Write(t)
				}
			}
		}
		rc.Close()
		if page.Len() == 0 {
			continue
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("—— 第 %d 页 ——\n", s.n))
		sb.WriteString(strings.TrimRight(page.String(), "\n"))
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("extract: no text found in %s", filepath.Base(path))
	}
	return capText(sb.String()), nil
}

// pdfText pulls the text of every page through the pure-Go reader.
func pdfText(path string) (string, error) {
	f, r, err := pdflib.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("—— 第 %d 页 ——\n", i))
		sb.WriteString(text)
		if sb.Len() > MaxText*2 {
			break
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("extract: no readable text in %s（可能是扫描件或加密 PDF）", filepath.Base(path))
	}
	return capText(sb.String()), nil
}

// readSharedStrings loads xl/sharedStrings.xml into a slice.
func readSharedStrings(zr *zip.Reader) ([]string, error) {
	f := findFile(zr, "xl/sharedStrings.xml")
	if f == nil {
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var out []string
	var cur strings.Builder
	inSI, inT := false, false
	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				cur.Reset()
				inSI = true
			case "t":
				inT = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inT = false
			case "si":
				inSI = false
				out = append(out, cur.String())
			}
		case xml.CharData:
			if inSI && inT {
				cur.Write(t)
			}
		}
	}
	return out, nil
}

// readWorkbook maps sheet names to their worksheet files through the
// workbook's relationships, then parses the cells of each sheet.
func readWorkbook(zr *zip.Reader, shared []string) ([]Sheet, error) {
	rels := map[string]string{}
	if f := findFile(zr, "xl/_rels/workbook.xml.rels"); f != nil {
		rc, err := f.Open()
		if err == nil {
			dec := xml.NewDecoder(rc)
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if t, ok := tok.(xml.StartElement); ok && t.Name.Local == "Relationship" {
					id, target := "", ""
					for _, a := range t.Attr {
						switch a.Name.Local {
						case "Id":
							id = a.Value
						case "Target":
							target = a.Value
						}
					}
					if id != "" && target != "" {
						rels[id] = strings.TrimPrefix(target, "/xl/")
					}
				}
			}
			rc.Close()
		}
	}

	var names []string
	var targets []string
	if f := findFile(zr, "xl/workbook.xml"); f != nil {
		rc, err := f.Open()
		if err == nil {
			dec := xml.NewDecoder(rc)
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if t, ok := tok.(xml.StartElement); ok && t.Name.Local == "sheet" {
					name, id := "", ""
					for _, a := range t.Attr {
						switch a.Name.Local {
						case "name":
							name = a.Value
						case "id":
							id = a.Value
						}
					}
					target := rels[id]
					if target == "" {
						continue
					}
					names = append(names, name)
					targets = append(targets, "xl/"+target)
				}
			}
			rc.Close()
		}
	}

	var sheets []Sheet
	for i, target := range targets {
		f := findFile(zr, target)
		if f == nil {
			continue
		}
		rows, err := readSheet(f, shared)
		if err != nil {
			continue
		}
		name := fmt.Sprintf("Sheet %d", i+1)
		if i < len(names) && names[i] != "" {
			name = names[i]
		}
		sheets = append(sheets, Sheet{Name: name, Rows: rows})
	}
	return sheets, nil
}

// readSheet parses one worksheet into a trimmed cell matrix.
func readSheet(f *zip.File, shared []string) ([][]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var rows [][]string
	var row []string
	cellType := ""
	var cell strings.Builder
	inValue := false
	flushCell := func() {
		v := strings.TrimSpace(cell.String())
		if cellType == "s" {
			if idx, err := strconv.Atoi(v); err == nil && idx >= 0 && idx < len(shared) {
				v = shared[idx]
			}
		}
		row = append(row, v)
		cell.Reset()
		inValue = false
		cellType = ""
	}
	flushRow := func() {
		if len(row) > 0 {
			rows = append(rows, row)
		}
		row = nil
	}
	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				row = nil
			case "c":
				cellType = ""
				for _, a := range t.Attr {
					if a.Name.Local == "t" {
						cellType = a.Value
					}
				}
			case "v", "t":
				inValue = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "c":
				flushCell()
			case "row":
				flushRow()
			case "v", "t":
				inValue = false
			}
		case xml.CharData:
			if inValue {
				cell.Write(t)
			}
		}
	}
	flushRow()
	return rows, nil
}

func findFile(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}
