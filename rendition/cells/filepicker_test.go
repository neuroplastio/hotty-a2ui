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

	// Click entry 0 (src/): first selects it, second opens it
	if err := r.Click(3, 2); err != nil {
		t.Fatal(err)
	}
	if err := r.Click(3, 2); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.FileDir != "src" {
		t.Fatalf("fileDir = %q, want src", e.FileDir)
	}
	plain := r.Draw(40).Plain()
	if !strings.Contains(plain, "◂ /src") {
		t.Errorf("expected ◂ /src in title:\n%s", plain)
	}
	if got := r.Pointer(3, 0); got != "pointer" {
		t.Errorf("pointer over title in subdir = %q, want pointer", got)
	}

	// Click title row to go back up
	if err := r.Click(3, 0); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.FileDir != "" {
		t.Errorf("fileDir after back = %q, want empty", e.FileDir)
	}
	r.Draw(40)
	if got := r.Pointer(3, 0); got != "" {
		t.Errorf("pointer over title at root = %q, want empty", got)
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

func TestFilePickerCellsScrollAndWheel(t *testing.T) {
	fsys := fstest.MapFS{
		"d0":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"d1":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"d2":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"d3":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"d3/sub.txt": &fstest.MapFile{Data: []byte("hi\n"), Mode: 0644},
		"f4.txt":     &fstest.MapFile{Data: []byte("f4\n"), Mode: 0644},
		"f5.txt":     &fstest.MapFile{Data: []byte("f5\n"), Mode: 0644},
	}
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"dir":"","sel":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":"project",
	  "path":{"@path":"/dir"},"selected":{"@path":"/sel"},"height":3}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Grant("project", fsys)
	c.Rebuild()
	r := New(c)

	// Page 0: d0, d1, d2
	f := r.Draw(40).Plain()
	if !strings.Contains(f, "d0/") || !strings.Contains(f, "d2/") || strings.Contains(f, "d3/") {
		t.Fatalf("page 0 output unexpected:\n%s", f)
	}

	// Wheel down over blank separator row (y=1) scrolls down
	if !r.Wheel(5, 1, 0, 1) {
		t.Error("wheel down over separator was not handled")
	}
	f = r.Draw(40).Plain()
	if !strings.Contains(f, "d3/") || strings.Contains(f, "d0/") {
		t.Fatalf("wheel down did not scroll to d3:\n%s", f)
	}

	// Wheel up over the file picker scrolls back to d0
	if !r.Wheel(5, 3, 0, -1) {
		t.Error("wheel up was not handled")
	}
	f = r.Draw(40).Plain()
	if !strings.Contains(f, "d0/") {
		t.Fatalf("wheel up did not scroll to d0:\n%s", f)
	}

	// Re-select d3 and open it
	c.ClickFile("root", 3) // select
	c.ClickFile("root", 3) // open
	e := c.V.Find("root")
	if e.FileDir != "d3" {
		t.Fatalf("fileDir = %q, want d3", e.FileDir)
	}

	// Inside d3, click title to go back up
	r.Draw(40)
	if err := r.Click(3, 0); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.FileDir != "" || e.Selected != 3 || e.Top != 1 {
		t.Errorf("after back: dir %q, selected %d, top %d; want empty, 3, 1", e.FileDir, e.Selected, e.Top)
	}
	f = r.Draw(40).Plain()
	if !strings.Contains(f, "d3/") {
		t.Errorf("drawing after back did not restore view showing d3:\n%s", f)
	}
	if !strings.Contains(f, "┃") {
		t.Errorf("expected scrollbar thumb in output:\n%s", f)
	}
}

// TestFilePickerCellsModalEscape: in a Modal, Escape in a picker in a
// subdirectory goes up, and at its root closes the Modal.
func TestFilePickerCellsModalEscape(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"dir":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["m"]},
	 {"id":"m","component":"Modal","trigger":"open","content":"fp"},
	 {"id":"open_t","component":"Text","text":"Open"},
	 {"id":"open","component":"Button","child":"open_t","action":{"event":{"name":"o"}}},
	 {"id":"fp","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":"project","path":{"@path":"/dir"}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Grant("project", cellsTestFS)
	c.Rebuild()
	r := New(c)
	openModal(t, c)
	c.Focus("fp")
	r.Draw(40)
	for _, key := range []string{"enter", "escape"} { // into src/, and out
		if ok, err := r.Key(key); !ok || err != nil {
			t.Fatalf("%s: %v, %v", key, ok, err)
		}
	}
	if got := c.S.Data.Value("/dir"); got != "" || c.V.Overlay == nil {
		t.Fatalf("after Enter, Escape: /dir %v, Modal open %v; want the root, open", got, c.V.Overlay != nil)
	}
	if ok, err := r.Key("escape"); !ok || err != nil || c.V.Overlay != nil {
		t.Errorf("Escape at the root left the Modal open (%v, %v)", ok, err)
	}
}
