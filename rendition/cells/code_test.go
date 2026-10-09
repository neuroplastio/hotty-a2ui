package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// coding is a surface of the components given (JSON), in a Column.
func coding(t *testing.T, comps ...string) *Rendition {
	t.Helper()
	var ids []string
	for _, c := range comps {
		ids = append(ids, `"`+strings.SplitN(strings.SplitN(c, `"id":"`, 2)[1], `"`, 2)[0]+`"`)
	}
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Column","children":[` + strings.Join(ids, ",") + `]},` + strings.Join(comps, ",") + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return New(view.NewController(p.Surface("s")))
}

func codeComp(id, props string) string {
	return `{"id":"` + id + `","component":"HottyCode","catalogId":"` + hotty.ID + `",` + props + `}`
}

// A HottyCode's rows are its lines: the number right-aligned as wide as
// the widest, in muted, and a space; the mark's sign and a space, when it
// has marks; then the code, coloured by kind. A long line wraps under the
// code, the gutter blank; without wrap it is cut with "…".
func TestCodeDraws(t *testing.T) {
	r := coding(t, codeComp("c", `"code":"x := \"hi\" // say\nreturn x + 10","language":"go","lineNumbers":true,"startLine":9,
		"marks":[{"line":10,"kind":"added"}]`))
	f := r.Draw(40)
	want := " 9   x := \"hi\" // say\n10 + return x + 10"
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, c := range []struct {
		x, y int
		role Role
		attr Attr
	}{{1, 0, Muted, 0}, {10, 0, Success, 0}, {15, 0, Muted, Italic}, {5, 1, Info, Bold}, {3, 1, Success, 0}, {16, 1, Warning, 0}, {5, 0, Fg, 0}} {
		if got := f.Cells[c.y][c.x]; got.Role != c.role || got.Attr != c.attr {
			t.Errorf("(%d,%d) %q: %v %v, want %v %v", c.x, c.y, got.Text, got.Role, got.Attr, c.role, c.attr)
		}
	}
	long := codeComp("c", `"code":"abcdefghijklmnopqrstuvwxyz","lineNumbers":true`)
	if got := coding(t, long).Draw(12).Plain(); got != "1 abcdefghij\n  klmnopqrst\n  uvwxyz" {
		t.Errorf("wrapped:\n%s", got)
	}
	cut := codeComp("c", `"code":"abcdefghijklmnopqrstuvwxyz","wrap":false`)
	if got := coding(t, cut).Draw(12).Plain(); got != "abcdefghijk…" {
		t.Errorf("cut: %q", got)
	}
}

// A marked line's row is tinted across the width toward its role, where
// the theme has colours: a highlighted one with the selection colour
// itself, the others a sixth of the way; with the terminal's, it is not.
func TestCodeTint(t *testing.T) {
	r := coding(t, codeComp("c", `"code":"a\nb\nc","marks":[{"line":1,"kind":"highlight"},{"line":3,"kind":"error"}]`))
	f := r.Draw(10)
	if c := f.Cells[0][9]; c.Back != Selection || c.BackMix != 255 {
		t.Errorf("a highlighted row's last cell: %v %d", c.Back, c.BackMix)
	}
	if c := f.Cells[1][0]; c.BackMix != 0 {
		t.Errorf("an unmarked row is tinted: %v %d", c.Back, c.BackMix)
	}
	if c := f.Cells[2][0]; c.Back != Error || c.BackMix != 42 || c.Text != "✗" || c.Role != Error {
		t.Errorf("an error row's sign: %+v", c)
	}
	th := theme.Theme{Name: "t", Bg: "#000000", Fg: "#ffffff", Selection: "#336699", Error: "#ff0000"}
	out := f.Themed(th)
	if !strings.Contains(out, "48;2;51;102;153") || !strings.Contains(out, "48;2;42;0;0") {
		t.Errorf("no tints in %q", out)
	}
	if strings.Contains(f.Themed(theme.Default), "48;") {
		t.Error("the terminal's background is tinted")
	}
}

// A HottyCode takes its gutter and its widest line, and a blank row
// follows it in a Column, as one follows a scroll view.
func TestCodeLayout(t *testing.T) {
	r := coding(t, codeComp("c", `"code":"abc\nabcdef","lineNumbers":true,"marks":[{"line":1,"kind":"warning"}]`),
		`{"id":"after","component":"Text","text":"|"}`)
	if got := r.Draw(40).Plain(); got != "1 ! abc\n2   abcdef\n\n|" {
		t.Errorf("got %q", got)
	}
	if w := codeWidth(r.c.V.Find("c")); w != 2+2+6 {
		t.Errorf("natural width %d", w)
	}
}

// A Text's fenced code block is highlighted as its fence says, two
// columns in, its plain tokens in the Text's colour.
func TestTextCodeBlock(t *testing.T) {
	r := coding(t, `{"id":"t","component":"Text","text":"Run:\n\n`+"```go"+`\nreturn nil\n`+"```"+`"}`)
	f := r.Draw(20)
	if got := f.Plain(); got != "Run:\n  return nil" {
		t.Fatalf("got %q", got)
	}
	if k, n := f.Cells[1][2], f.Cells[1][9]; k.Role != Info || k.Attr != Bold || n.Role != Warning {
		t.Errorf("return %v, nil %v", k.Role, n.Role)
	}
}
