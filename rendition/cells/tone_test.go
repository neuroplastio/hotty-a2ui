package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A tone (io_neuroplast_hotty.tone, profile §6.33) is the role a Text's or
// an Icon's glyphs take: a literal, or a path in a template's item, which
// follows the data model. A caption's muted gives way to it, and a value
// that names no tone leaves the Text its own.
func TestTone(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"rows":[
	  {"name":"alpha","tone":"success"},{"name":"beta","tone":"error"},{"name":"gamma","tone":"plum"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["rows","cap","mark"]},
	 {"id":"rows","component":"List","children":{"componentId":"row","path":"/rows"}},
	 {"id":"row","component":"Text","text":{"@path":"name"},
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"tone":{"@path":"tone"}}}}},
	 {"id":"cap","component":"Text","text":"waiting","variant":"caption",
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"tone":"warning"}}}},
	 {"id":"mark","component":"Icon","name":"check",
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"tone":"info"}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	s := p.Surface("s")
	c := view.NewController(s)
	r := New(c)
	at := func(f *Frame, word string) Role {
		t.Helper()
		for y, line := range strings.Split(f.Plain(), "\n") {
			if x := strings.Index(line, word); x >= 0 {
				return f.Cells[y][Width(line[:x])].Role
			}
		}
		t.Fatalf("%q is not drawn:\n%s", word, f.Plain())
		return Fg
	}
	f := r.Draw(30)
	for word, want := range map[string]Role{"alpha": Success, "beta": Error, "gamma": Fg, "waiting": Warning, "✓": Info} {
		if got := at(f, word); got != want {
			t.Errorf("%s: %v, want %v", word, got, want)
		}
	}

	// The item's data moves its tone.
	if err := p.ProcessJSON([]byte(`{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/rows/1",
	  "value":{"name":"beta","tone":"muted"}}}`)); err != nil {
		t.Fatal(err)
	}
	c.Rebuild() // as a program does on the processor's Changed
	if got := at(r.Draw(30), "beta"); got != Muted {
		t.Errorf("beta after its tone moved: %v, want %v", got, Muted)
	}
}
