package cells

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

var cellsTestFS = fstest.MapFS{
	"README.md": &fstest.MapFile{Data: []byte("# Title\n"), Mode: 0644},
	"main.go":   &fstest.MapFile{Data: []byte("package main\n"), Mode: 0644},
	"src":       &fstest.MapFile{Mode: fs.ModeDir | 0755},
}

func filePickerCellsHarness(t *testing.T, extra string) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"dir":"","sel":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":"project",
	  "path":{"@path":"/dir"},"selected":{"@path":"/sel"}` + extra + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Grant("project", cellsTestFS)
	c.Rebuild()
	return New(c), c
}

func TestFilePickerCellsDraw(t *testing.T) {
	r, c := filePickerCellsHarness(t, "")
	// Unfocused draw: title row is /
	plain := r.Draw(30).Plain()
	lines := strings.Split(plain, "\n")
	if len(lines) < 4 {
		t.Fatalf("plain output too short:\n%s", plain)
	}
	// Title has "/"
	if !strings.Contains(lines[0], "/") {
		t.Errorf("line 0 = %q, want title with /", lines[0])
	}
	// Index 0 is src/
	if !strings.Contains(plain, "src/") {
		t.Errorf("expected src/ in output:\n%s", plain)
	}
	// Files have sizes
	if !strings.Contains(plain, "main.go") || !strings.Contains(plain, "B") {
		t.Errorf("expected main.go with size in output:\n%s", plain)
	}

	// Focus gives selected entry the accent
	c.Focus("root")
	focused := r.Draw(30).Plain()
	if !strings.Contains(focused, "│") {
		t.Errorf("expected selection bar │ in focused output:\n%s", focused)
	}
}

func TestFilePickerCellsClick(t *testing.T) {
	r, c := filePickerCellsHarness(t, "")
	r.Draw(40)
	// Click on entry 1: main.go row (row 3: title is row 0, row 1 is blank, row 2 is src/, row 3 is main.go)
	if err := r.Click(3, 3); err != nil {
		t.Fatal(err)
	}
	e := c.V.Find("root")
	if e.Selected != 1 {
		t.Errorf("selected after click = %d, want 1", e.Selected)
	}

	// Second click picks main.go
	if err := r.Click(3, 3); err != nil {
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/sel"); got != "main.go" {
		t.Errorf("picked file = %v, want main.go", got)
	}
}

func TestFilePickerCellsErrorAndEmpty(t *testing.T) {
	r, c := filePickerCellsHarness(t, "")
	c.Grant("project", nil)
	c.Rebuild()
	plain := r.Draw(40).Plain()
	if !strings.Contains(plain, "✗ No root named") {
		t.Errorf("expected error in output:\n%s", plain)
	}
}
