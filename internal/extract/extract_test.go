package extract

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDocxText(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	doc := `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>需求背景</w:t></w:r></w:p>
    <w:p><w:r><w:t>Mana </w:t></w:r><w:r><w:t>是本地优先的</w:t></w:r><w:r><w:t>项目管理器</w:t></w:r></w:p>
    <w:p><w:r><w:tab/><w:t>缩进的一段</w:t></w:r></w:p>
  </w:body>
</w:document>`
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte(doc))
	zw.Close()

	text, err := Text(writeTemp(t, "需求说明.docx", buf.String()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"需求背景", "Mana 是本地优先的项目管理器", "\t缩进的一段"} {
		if !strings.Contains(text, want) {
			t.Fatalf("docx text missing %q in %q", want, text)
		}
	}
}

func TestPptxTextKeepsPageOrder(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	slide := func(text string) string {
		return `<?xml version="1.0"?><p:sld xmlns:p="pkg" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:txBody><a:p><a:r><a:t>` +
			text + `</a:t></a:r></a:p></p:txBody></p:sld>`
	}
	w, _ := zw.Create("ppt/slides/slide2.xml")
	w.Write([]byte(slide("第二页内容")))
	w, _ = zw.Create("ppt/slides/slide1.xml")
	w.Write([]byte(slide("第一页标题")))
	zw.Close()

	text, err := Text(writeTemp(t, "发布材料.pptx", buf.String()))
	if err != nil {
		t.Fatal(err)
	}
	i1 := strings.Index(text, "第一页标题")
	i2 := strings.Index(text, "第二页内容")
	if i1 < 0 || i2 < 0 || i1 > i2 {
		t.Fatalf("pptx pages out of order: %q", text)
	}
	if !strings.Contains(text, "—— 第 1 页 ——") || !strings.Contains(text, "—— 第 2 页 ——") {
		t.Fatalf("pptx page markers missing: %q", text)
	}
}

func TestXlsxSheets(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, body string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	add("xl/workbook.xml", `<?xml version="1.0"?>
	<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	          xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
	  <sheets><sheet name="预算" sheetId="1" r:id="rId1"/></sheets>
	</workbook>`)
	add("xl/_rels/workbook.xml.rels", `<?xml version="1.0"?>
	<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
	  <Relationship Id="rId1" Target="worksheets/sheet1.xml"/>
	</Relationships>`)
	add("xl/sharedStrings.xml", `<?xml version="1.0"?>
	<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
	  <si><t>模块</t></si><si><t>v1.0</t></si><si><t>支持</t></si>
	</sst>`)
	add("xl/worksheets/sheet1.xml", `<?xml version="1.0"?>
	<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
	  <sheetData>
	    <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c><c r="C1" t="s"><v>2</v></c></row>
	    <row r="2"><c r="A2" t="s"><v>0</v></c><c r="B2"><v>42</v></c><c r="C2" t="inlineStr"><is><t>是</t></is></c></row>
	  </sheetData>
	</worksheet>`)
	zw.Close()

	path := writeTemp(t, "预算表.xlsx", buf.String())
	sheets, err := Sheets(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 1 || sheets[0].Name != "预算" {
		t.Fatalf("sheets = %#v", sheets)
	}
	rows := sheets[0].Rows
	if len(rows) != 2 {
		t.Fatalf("rows = %#v", rows)
	}
	if rows[0][0] != "模块" || rows[0][1] != "v1.0" || rows[1][2] != "是" {
		t.Fatalf("cells = %#v", rows)
	}
}

// minimalPDF assembles a one-page PDF with an uncompressed content
// stream, offsets and all, so the reader can walk it.
func minimalPDF(t *testing.T, text string) string {
	t.Helper()
	var buf bytes.Buffer
	offsets := make([]int, 6)
	write := func(s string) { buf.WriteString(s) }
	mark := func(i int) { offsets[i] = buf.Len() }

	write("%PDF-1.4\n")
	mark(1)
	write("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	mark(2)
	write("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	mark(3)
	write("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>\nendobj\n")
	mark(4)
	stream := "BT /F1 24 Tf 72 720 Td (" + text + ") Tj ET"
	write("4 0 obj\n<< /Length " + itoa(len(stream)) + " >>\nstream\n" + stream + "\nendstream\nendobj\n")
	mark(5)
	write("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	xref := buf.Len()
	write("xref\n0 6\n")
	write("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		write(pad10(offsets[i]) + " 00000 n \n")
	}
	write("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n" + itoa(xref) + "\n%%EOF\n")

	return writeTemp(t, "样本.pdf", buf.String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func pad10(n int) string {
	s := itoa(n)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}

func TestPDFText(t *testing.T) {
	text, err := Text(minimalPDF(t, "Hello Mana"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Hello Mana") {
		t.Fatalf("pdf text = %q", text)
	}
	if !strings.Contains(text, "—— 第 1 页 ——") {
		t.Fatalf("pdf page marker missing: %q", text)
	}
}

func TestKindFor(t *testing.T) {
	cases := map[string]Kind{
		"报告.pdf":       KindPDF,
		"说明.DOCX":      KindWord,
		"预算.xlsx":      KindExcel,
		"发布.pptx":      KindSlides,
		"截图.png":       KindImage,
		"README.md":    KindMarkdown,
		"notes.txt":    KindText,
		"firmware.bin": KindOther,
	}
	for name, want := range cases {
		if got := KindFor(name); got != want {
			t.Errorf("KindFor(%q) = %v, want %v", name, got, want)
		}
	}
}
