package html

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

var htmlTestFS = fstest.MapFS{
	"README.md": &fstest.MapFile{Data: []byte("# Project\n"), Mode: 0644},
	"main.go":   &fstest.MapFile{Data: []byte("package main\n"), Mode: 0644},
	"docs":      &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"docs/a.md": &fstest.MapFile{Data: []byte("a\n"), Mode: 0644},
}

func TestFilePickerOnHost(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"dir":"","sel":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["fp"]},
 {"id":"fp","component":"HottyFilePicker",`+h+`,"root":"project","path":{"@path":"/dir"},"selected":{"@path":"/sel"},"height":4,
  "onPick":{"event":{"name":"pick","context":{"file":{"@path":"/sel"}}}},
  "onChange":{"event":{"name":"change","context":{"dir":{"@path":"/dir"}}}}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	r.C.Grant("project", htmlTestFS)
	r.C.Rebuild()
	x.update(r)

	s := x.h.Surface(r.name)
	for _, want := range [][3]string{
		{"fp", "class", "k-filepicker"},
		{"fp", "tabindex", "0"},
		{"fp", "role", "listbox"},
		{"fp", "aria-label", "/"},
		{"fp", "data-keys", fileKeys},
		{partID("fp", partTitle), "class", "k-file-title"},
		{partID("fp", partEntry+"0"), "class", "k-file-item k-sel"},
		{partID("fp", partEntry+"0"), "role", "option"},
		{partID("fp", partEntry+"0"), "aria-selected", "true"},
		{partID("fp", partEntry+"1"), "class", "k-file-item"},
		{partID("fp", partEntry+"1"), "aria-selected", "false"},
	} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}

	// Text content has directory and files
	txt := s.TextOf("fp")
	if !strings.Contains(txt, "docs/") || !strings.Contains(txt, "main.go") {
		t.Errorf("text of fp = %q", txt)
	}

	// Click entry 1 (main.go): selects it
	must(t, x.h.Click(r.name, partID("fp", partEntry+"1")))
	x.pump()
	if r.C.St.Focus != "fp" || !r.C.St.Keyboard {
		t.Fatalf("focus %q (%v), want fp", r.C.St.Focus, r.C.St.Keyboard)
	}
	if got, _ := s.Attr(partID("fp", partEntry+"1"), "aria-selected"); got != "true" {
		t.Errorf("entry 1 aria-selected = %q, want true", got)
	}

	// Second click picks main.go
	must(t, x.h.Click(r.name, partID("fp", partEntry+"1")))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "main.go" {
		t.Errorf("/sel = %v, want main.go", got)
	}

	// Click entry 0 (docs): first selects it, second opens it
	must(t, x.h.Click(r.name, partID("fp", partEntry+"0")))
	x.pump()
	must(t, x.h.Click(r.name, partID("fp", partEntry+"0")))
	x.pump()
	x.update(r)
	if got := r.C.S.Data.Value("/dir"); got != "docs" {
		t.Fatalf("/dir = %v, want docs", got)
	}
	if got, _ := s.Attr(partID("fp", partTitle), "class"); !strings.Contains(got, "k-file-up") {
		t.Errorf("title class in subdir = %q, want k-file-up", got)
	}

	// Click title to go back up
	must(t, x.h.Click(r.name, partID("fp", partTitle)))
	x.pump()
	x.update(r)
	if got := r.C.S.Data.Value("/dir"); got != "" {
		t.Errorf("/dir after back = %v, want empty", got)
	}
	if got, _ := s.Attr(partID("fp", partTitle), "class"); got != "k-file-title" {
		t.Errorf("title class at root = %q, want k-file-title", got)
	}
	if got := r.C.V.Find("fp").Selected; got != 0 {
		t.Errorf("selected after back = %d, want 0", got)
	}
	if got, _ := s.Attr(partID("fp", partEntry+"0"), "aria-selected"); got != "true" {
		t.Errorf("entry 0 aria-selected after back = %q, want true", got)
	}
}

func TestFilePickerOnHostScrollbar(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"dir":"","sel":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["fp"]},
 {"id":"fp","component":"HottyFilePicker",`+h+`,"root":"project","height":2}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	r.C.Grant("project", htmlTestFS)
	r.C.Rebuild()
	x.update(r)

	doc := r.Doc()
	if !strings.Contains(doc, "k-file-scrollbar") || !strings.Contains(doc, "k-file-thumb") {
		t.Errorf("expected k-file-scrollbar and k-file-thumb in doc:\n%s", doc)
	}
}
