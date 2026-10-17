package cells

import (
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A cycle (io_neuroplast_hotty.cycle) is a select's row, its input
// "‹ label ›", the label centred between the arrows, which stay put as the
// pick changes: muted, and in the accent while it has the keyboard, with
// no underline. The arrows and its other keys step it round (Enter too,
// as a host clicks a focused button); a click on "‹ " steps it back, and
// one anywhere else on its row on.
func TestCycle(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"env":["stage"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["env","go"]},
	 {"id":"env","component":"ChoicePicker","label":"Env","value":{"@path":"/env"},
	  "options":[{"label":"production","value":"prod"},{"label":"staging","value":"stage"},{"label":"preview","value":"prev"}],
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"cycle":true}}}},
	 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go"}}},
	 {"id":"go_t","component":"Text","text":"Go"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	s := p.Surface("s")
	c := view.NewController(s)
	r := New(c)
	env := func() any { return s.Data.Value("/env").([]any)[0] }
	// The input starts after the label, a space and the inset.
	at := gutter + len("Env") + 1 + inset
	row := func(f *Frame) string { return strings.Split(f.Plain(), "\n")[0] }

	f := r.Draw(40)
	if got := row(f); got != "  Env  ‹  staging   ›" {
		t.Fatalf("row %q", got)
	}
	arrows := func(f *Frame, want Role) {
		t.Helper()
		for _, x := range []int{at, at + 13} {
			if g := f.Cells[0][x]; g.Role != want || g.Attr&Underline != 0 {
				t.Errorf("%q at %d: %v, %v", g.Text, x, g.Role, g.Attr)
			}
		}
	}
	arrows(f, Muted)

	c.Focus("env")
	for _, step := range []struct {
		key, want, face string
	}{
		{"arrow_right", "prev", "‹  preview   ›"},
		{"arrow_right", "prod", "‹ production ›"},
		{"arrow_left", "prev", "‹  preview   ›"},
		{"enter", "prod", "‹ production ›"},
		{"space", "stage", "‹  staging   ›"},
	} {
		keys(t, r, step.key)
		f = r.Draw(40)
		if env() != step.want || !strings.HasSuffix(row(f), step.face) {
			t.Fatalf("%s: env %v, row %q", step.key, env(), row(f))
		}
	}
	arrows(f, Accent)

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	var got []any
	for _, col := range []int{at, at + 1, at + 5, at + 13, gutter} {
		must(r.Click(col, 0))
		r.Draw(40)
		got = append(got, env())
	}
	if want := []any{"prod", "prev", "prod", "stage", "prev"}; !slices.Equal(got, want) {
		t.Errorf("clicks on ‹, its space, the value, › and the label: %v, want %v", got, want)
	}
	if r.ListOpen() {
		t.Error("a cycle opened a list")
	}
}

// A Slider's label joins its run (labelRuns): in a Column of a cycle and
// a Slider, the track starts where the cycle's input does, past the run's
// widest label, a space and the inset.
func TestSliderInRun(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"env":["prod"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["env","n"]},
	 {"id":"env","component":"ChoicePicker","label":"Environment","value":{"@path":"/env"},
	  "options":[{"label":"production","value":"prod"},{"label":"staging","value":"stage"}],
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"cycle":true}}}},
	 {"id":"n","component":"Slider","label":"Replicas","min":1,"max":10,"value":6}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	rows := strings.Split(r.Draw(40).Plain(), "\n")
	at := gutter + len("Environment") + 1 + inset
	if !strings.HasPrefix(rows[0][at:], "‹") || !strings.HasPrefix(rows[1], "  Replicas    ") || strings.HasPrefix(rows[1][at:], " ") {
		t.Errorf("the track starts at %d, as the cycle's input does:\n%s", at, strings.Join(rows, "\n"))
	}
}
