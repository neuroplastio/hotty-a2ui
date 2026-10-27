package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// lines (io_neuroplast_hotty.lines, profile §6.40) cuts a Text to that
// many rows, the last ending in …: one is its first line unwrapped, cut
// at the room it has, so that a list row's title takes the row beside its
// badge, which stays whole. Two keep the wrapped rows. Its minimum is a
// column, so that a Row squeezes it before its neighbours.
func TestTextClamp(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	lines := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"lines":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["row","two","paras","short"]},
	 {"id":"row","component":"Row","children":["title","badge"]},
	 {"id":"title","component":"Text","text":"Write the release notes for 0.6","weight":1,` + lines("1") + `},
	 {"id":"badge","component":"Text","text":"due Friday",` + lines("1") + `},
	 {"id":"two","component":"Text","text":"one two three four five six seven eight nine ten eleven twelve",` + lines("2") + `},
	 {"id":"paras","component":"Text","text":"First\n\nSecond",` + lines("1") + `},
	 {"id":"short","component":"Text","text":"Fits",` + lines("1") + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	got := strings.Split(r.Draw(24).Plain(), "\n")
	want := []string{
		"Write the re… due Friday",
		"one two three four five",
		"six seven eight nine te…",
		"First…",
		"Fits",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
