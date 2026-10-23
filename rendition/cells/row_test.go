package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A Row's weighted children share the columns its others leave as a
// host's flex row shares them (profile §3.3): from nothing, by weight,
// whatever they hold, none narrower than its minimum.
func TestRowWeights(t *testing.T) {
	for _, tc := range []struct {
		name  string
		comps string
		w     int
		want  map[string][2]int // id: x, width
	}{{
		// Equal weights, unequal texts: halves, as on a host, where
		// cells once gave each its own width and halved only the spare.
		name: "halves",
		comps: `{"id":"root","component":"Row","children":["a","b"]},
		 {"id":"a","component":"Text","text":"A","weight":1},
		 {"id":"b","component":"Text","text":"Bbbbbbbbbb","weight":1}`,
		w:    31,
		want: map[string][2]int{"a": {0, 15}, "b": {16, 15}},
	}, {
		// An unweighted child keeps its natural width; the weighted
		// share the rest.
		name: "beside a control",
		comps: `{"id":"root","component":"Row","children":["a","b","c"]},
		 {"id":"a","component":"Text","text":"Label","weight":2},
		 {"id":"b","component":"Text","text":"Other label","weight":1},
		 {"id":"c","component":"Text","text":"value"}`,
		w:    37,
		want: map[string][2]int{"a": {0, 20}, "b": {21, 10}, "c": {32, 5}},
	}, {
		// A part narrower than a child's minimum, its longest word, is
		// held there, and the others share what remains.
		name: "held at its minimum",
		comps: `{"id":"root","component":"Row","children":["a","b"]},
		 {"id":"a","component":"Text","text":"one two","weight":3},
		 {"id":"b","component":"Text","text":"unbreakable1","weight":1}`,
		w:    31,
		want: map[string][2]int{"a": {0, 18}, "b": {19, 12}},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
			msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
			{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[` + tc.comps + `]}}]`
			if err := p.ProcessJSON([]byte(msgs)); err != nil {
				t.Fatal(err)
			}
			r := New(view.NewController(p.Surface("s")))
			f := r.Draw(tc.w)
			for id, want := range tc.want {
				x, _, w, _, ok := r.Box(id)
				if !ok || x != want[0] || w != want[1] {
					t.Errorf("%s: x %d, width %d (%v), want %v\n%s", id, x, w, ok, want, f.Plain())
				}
			}
		})
	}
}
