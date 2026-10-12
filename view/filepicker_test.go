package view_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

var testFS = fstest.MapFS{
	"README.md":        &fstest.MapFile{Data: []byte("# Test\n"), Mode: 0644},
	"go.mod":           &fstest.MapFile{Data: []byte("module test\n"), Mode: 0644},
	"main.go":          &fstest.MapFile{Data: []byte("package main\n"), Mode: 0644},
	"src":              &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/app.go":       &fstest.MapFile{Data: []byte("package src\n"), Mode: 0644},
	"src/view":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/view/view.go": &fstest.MapFile{Data: []byte("package view\n"), Mode: 0644},
	"docs":             &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"docs/intro.md":    &fstest.MapFile{Data: []byte("intro\n"), Mode: 0644},
}

func filePickerHarness(t *testing.T, extra string) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"dir":"","sel":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":"project",
	  "path":{"@path":"/dir"},"selected":{"@path":"/sel"},"height":3` + extra + `,
	  "onPick":{"event":{"name":"pick","context":{"file":{"@path":"/sel"}}}},
	  "onChange":{"event":{"name":"change","context":{"dir":{"@path":"/dir"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Grant("project", testFS)
	c.Rebuild()
	return c, actions
}

func TestFilePickerRoots(t *testing.T) {
	c, _ := filePickerHarness(t, "")
	e := c.V.Find("root")
	if e == nil || e.Kind != view.FilePicker {
		t.Fatalf("element %v", e)
	}
	if e.Error != "" {
		t.Errorf("unexpected error: %s", e.Error)
	}
	if len(e.Files) != 5 { // docs, src, README.md, go.mod, main.go
		t.Errorf("got %d files, want 5", len(e.Files))
	}
	// Directories first, sorted
	if !e.Files[0].Dir || e.Files[0].Name != "docs" {
		t.Errorf("files[0] = %+v, want docs dir", e.Files[0])
	}
	if !e.Files[1].Dir || e.Files[1].Name != "src" {
		t.Errorf("files[1] = %+v, want src dir", e.Files[1])
	}
	if e.Files[2].Dir || e.Files[2].Name != "go.mod" {
		t.Errorf("files[2] = %+v, want go.mod file", e.Files[2])
	}

	// Ungrant root
	c.Grant("project", nil)
	c.Rebuild()
	e = c.V.Find("root")
	if e.Error != `No root named "project".` {
		t.Errorf("got error %q, want No root named project", e.Error)
	}
}

func TestFilePickerMissingOrEmptyRoot(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":""}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("root")
	if e == nil {
		t.Fatal("no element root")
	}
	if e.Error != "A file picker needs a root." {
		t.Errorf("got error %q, want A file picker needs a root.", e.Error)
	}
}

func TestFilePickerFilterAndEmpty(t *testing.T) {
	c, _ := filePickerHarness(t, `,"filter":"view","emptyText":"Nothing here."`)
	e := c.V.Find("root")
	// In root, no files match "view"
	if len(e.Files) != 0 {
		t.Errorf("got %d files, want 0", len(e.Files))
	}
	if e.Placeholder != "Nothing here." {
		t.Errorf("placeholder = %q, want Nothing here.", e.Placeholder)
	}
}

func TestFilePickerNavigation(t *testing.T) {
	c, actions := filePickerHarness(t, "")
	e := c.V.Find("root")
	if e.Selected != 0 {
		t.Errorf("initial selected = %d, want 0", e.Selected)
	}

	// ArrowDown moves selection
	ok, err := c.FileKey("root", "ArrowDown")
	if !ok || err != nil {
		t.Fatalf("FileKey: %v, %v", ok, err)
	}
	e = c.V.Find("root")
	if e.Selected != 1 { // src
		t.Errorf("selected = %d, want 1", e.Selected)
	}

	// Enter opens directory "src"
	ok, err = c.FileKey("root", "Enter")
	if !ok || err != nil {
		t.Fatalf("FileKey: %v, %v", ok, err)
	}
	if got := c.S.Data.Value("/dir"); got != "src" {
		t.Errorf("/dir = %v, want src", got)
	}
	if len(*actions) != 1 || (*actions)[0].Name != "change" {
		t.Errorf("actions = %+v, want [change]", *actions)
	}

	// In "src": dirs view, files app.go, util.go? Wait, testFS has src/app.go, src/view/view.go
	e = c.V.Find("root")
	if e.FileDir != "src" {
		t.Errorf("fileDir = %q, want src", e.FileDir)
	}
	if len(e.Files) != 2 { // view (dir), app.go (file)
		t.Errorf("files in src: %d (%+v), want 2", len(e.Files), e.Files)
	}

	// Escape goes up to root
	ok, err = c.FileKey("root", "Escape")
	if !ok || err != nil {
		t.Fatalf("FileKey: %v, %v", ok, err)
	}
	if got := c.S.Data.Value("/dir"); got != "" {
		t.Errorf("/dir = %v, want empty", got)
	}

	// Escape at root does nothing
	ok, err = c.FileKey("root", "Escape")
	if !ok || err != nil {
		t.Fatalf("FileKey: %v, %v", ok, err)
	}
	if got := c.S.Data.Value("/dir"); got != "" {
		t.Errorf("/dir = %v, want empty", got)
	}

	// Pick a file: End goes to last file (main.go), Enter picks it
	c.FileKey("root", "End")
	e = c.V.Find("root")
	if e.Selected != 4 {
		t.Errorf("selected after End: %d, want 4", e.Selected)
	}
	c.FileKey("root", "Enter")
	if got := c.S.Data.Value("/sel"); got != "README.md" {
		t.Errorf("/sel = %v, want README.md", got)
	}
	lastAction := (*actions)[len(*actions)-1]
	if lastAction.Name != "pick" || lastAction.Context["file"] != "README.md" {
		t.Errorf("last action = %+v, want pick with README.md", lastAction)
	}
}

func TestFilePickerClick(t *testing.T) {
	c, actions := filePickerHarness(t, "")
	// ClickFile on entry 1 (src, not selected) selects it
	if err := c.ClickFile("root", 1); err != nil {
		t.Fatal(err)
	}
	e := c.V.Find("root")
	if e.Selected != 1 {
		t.Errorf("selected = %d, want 1", e.Selected)
	}
	if len(*actions) != 0 {
		t.Errorf("unexpected action on select click")
	}

	// Second click on entry 1 enters directory
	if err := c.ClickFile("root", 1); err != nil {
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/dir"); got != "src" {
		t.Errorf("/dir = %v, want src", got)
	}
}

func TestFilePickerScrollAndFocusRestore(t *testing.T) {
	fsys := fstest.MapFS{
		"a_dir":        &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"b_dir":        &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"c_dir":        &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"d_dir":        &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"e_dir":        &fstest.MapFile{Mode: fs.ModeDir | 0755},
		"e_dir/sub.go": &fstest.MapFile{Data: []byte("package sub\n"), Mode: 0644},
		"file1.txt":    &fstest.MapFile{Data: []byte("1\n"), Mode: 0644},
		"file2.txt":    &fstest.MapFile{Data: []byte("2\n"), Mode: 0644},
	}
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"dir":"","sel":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyFilePicker","catalogId":"` + hotty.ID + `","root":"project",
	  "path":{"@path":"/dir"},"selected":{"@path":"/sel"},"height":4}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Grant("project", fsys)
	c.Rebuild()

	// Initially at root, page 0: entries 0..3 shown
	e := c.V.Find("root")
	if e.Top != 0 {
		t.Fatalf("initial top = %d, want 0", e.Top)
	}

	// Select entry 4 (e_dir), which is on page 1 (top 4)
	if err := c.ClickFile("root", 4); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.Selected != 4 || e.Top != 4 {
		t.Fatalf("after select: selected = %d, top = %d, want 4, 4", e.Selected, e.Top)
	}

	// Enter e_dir (second click)
	if err := c.ClickFile("root", 4); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.FileDir != "e_dir" {
		t.Fatalf("fileDir = %q, want e_dir", e.FileDir)
	}

	// In e_dir, select sub.go (entry 0)
	if err := c.ClickFile("root", 0); err != nil {
		t.Fatal(err)
	}

	// Now go back up (FileUp)
	if err := c.FileUp("root"); err != nil {
		t.Fatal(err)
	}
	e = c.V.Find("root")
	if e.FileDir != "" {
		t.Errorf("fileDir = %q, want empty", e.FileDir)
	}
	// Focus must be restored to e_dir (index 4)
	if e.Selected != 4 {
		t.Errorf("selected after back = %d, want 4 (e_dir)", e.Selected)
	}
	// Scrolling must be restored to page 1 (top 4)
	if e.Top != 4 {
		t.Errorf("top after back = %d, want 4", e.Top)
	}
}
