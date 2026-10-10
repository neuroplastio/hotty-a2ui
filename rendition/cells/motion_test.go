package cells

import (
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// motion is a surface of hotty components, one under another in a Column,
// drawn with a clock that stands at tick k of every interval asked about.
func motion(t *testing.T, data, comps string) (*Rendition, *view.Controller, *time.Time) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":` + data + `}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[` + strings.ReplaceAll(comps, "HOTTY", hotty.ID) + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	now := time.Unix(0, 0)
	r.Clock = func() time.Time { return now }
	return r, c, &now
}

// A Progress is its label, then a bar filled in eighths and the
// percentage; a value past max is a full bar. A bar with a value does not
// animate.
func TestProgress(t *testing.T) {
	r, _, _ := motion(t, `{}`, `
	 {"id":"root","component":"Column","children":["a","b"]},
	 {"id":"a","component":"HottyProgress","catalogId":"HOTTY","label":"Download","value":3,"max":8},
	 {"id":"b","component":"HottyProgress","catalogId":"HOTTY","value":9,"max":8}`)
	f := r.Draw(25)
	want := "Download\n" +
		"███████▌░░░░░░░░░░░░  38%\n" +
		"\n" +
		"████████████████████ 100%"
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if d := r.Animating(); d != 0 {
		t.Errorf("bars with values animate every %v", d)
	}
	// The fill blends across the bar (info into accent, or the terminal's
	// own gradient); a done bar is success.
	for x, mix := range map[int]uint8{0: 0, 6: 255 * 6 / 19, 7: 255 * 7 / 19} {
		if c := f.Cells[1][x]; c.Role != Info || c.To != Accent || c.Mix != mix || !c.Fill {
			t.Errorf("cell %d: %v to %v by %d (fill %v), want info to accent by %d", x, c.Role, c.To, c.Mix, c.Fill, mix)
		}
	}
	if c := f.Cells[3][19]; c.Role != Success || c.Mix != 0 || c.Fill {
		t.Errorf("a done bar's cell: %v by %d (fill %v), want success alone", c.Role, c.Mix, c.Fill)
	}
	if c := f.Cells[1][10]; c.Fill {
		t.Errorf("the track is a fill: %+v", c)
	}
	th := theme.Theme{Info: "#000000", Accent: "#ffffff"}
	for mix, want := range map[uint8]string{0: "38;2;0;0;0", 51: "38;2;51;51;51", 255: "38;2;255;255;255"} {
		if got := (Cell{Role: Info, To: Accent, Mix: mix}).style(&th); got != want {
			t.Errorf("mix %d: %q, want %q", mix, got, want)
		}
	}
	if got := (Cell{Role: Info, To: Accent, Mix: 51}).style(&theme.Default); got != "36" {
		t.Errorf("mix without the theme's colours: %q, want info's ANSI 36", got)
	}
	// Once the terminal said its colours, its theme blends them too; a
	// cell that blends nothing keeps the terminal's own, by number.
	term := theme.Default
	term.Term.ANSI[6], term.Term.ANSI[12] = "#000000", "#ffffff"
	if got := (Cell{Role: Info, To: Accent, Mix: 51}).style(&term); got != "38;2;51;51;51" {
		t.Errorf("mix in the terminal's colours: %q", got)
	}
	if got := (Cell{Role: Info, To: Accent}).style(&term); got != "36" {
		t.Errorf("no mix, the terminal's colours known: %q, want info's ANSI 36", got)
	}

	// A fill goes from the terminal's accent into its bright magenta once
	// it said both (its magenta, else), from the bar's first cell; where
	// it hasn't, or the theme colours the accent, it is info into accent.
	fill := func(mix uint8) Cell { return Cell{Role: Info, To: Accent, Mix: mix, Fill: true} }
	if got := fill(51).style(&term); got != "38;2;51;51;51" {
		t.Errorf("a fill, no magenta said: %q, want info into accent", got)
	}
	term.Term.ANSI[12], term.Term.ANSI[13] = "#0000ff", "#ff00ff"
	for mix, want := range map[uint8]string{0: "38;2;0;0;255", 51: "38;2;51;0;255", 255: "38;2;255;0;255"} {
		if got := fill(mix).style(&term); got != want {
			t.Errorf("a fill by %d, the terminal's magenta: %q, want %q", mix, got, want)
		}
	}
	term.Term.ANSI[13], term.Term.ANSI[5] = "", "#ff00ff"
	if got := fill(51).style(&term); got != "38;2;51;0;255" {
		t.Errorf("a fill, magenta from 5: %q", got)
	}
	named := theme.Theme{Info: "#000000", Accent: "#ffffff", Term: term.Term}
	if got := fill(51).style(&named); got != "38;2;51;51;51" {
		t.Errorf("a fill in a theme that colours the accent: %q, want info into accent", got)
	}
}

// A Progress without a value is a quarter of its bar, sliding across with
// the clock, and asks for a frame every ProgressInterval.
func TestProgressSlides(t *testing.T) {
	r, _, now := motion(t, `{}`, `
	 {"id":"root","component":"HottyProgress","catalogId":"HOTTY"}`)
	for _, at := range []struct {
		tick int64
		bar  string
	}{
		// The segment, 5 of the bar's 20, starts at step × 25 / 50 − 5.
		{0, "░░░░░░░░░░░░░░░░░░░░"},
		{2, "█░░░░░░░░░░░░░░░░░░░"},
		{10, "█████░░░░░░░░░░░░░░░"},
		{24, "░░░░░░░█████░░░░░░░░"},
		{49, "░░░░░░░░░░░░░░░░░░░█"},
		{50, "░░░░░░░░░░░░░░░░░░░░"},
	} {
		*now = time.Unix(0, at.tick*int64(view.ProgressInterval))
		if got := r.Draw(25).Plain(); got != at.bar {
			t.Errorf("tick %d: %q, want %q", at.tick, got, at.bar)
		}
		if d := r.Animating(); d != view.ProgressInterval {
			t.Errorf("tick %d: animates every %v, want %v", at.tick, d, view.ProgressInterval)
		}
	}
}

// A Spinner is the frame the clock is at, then its label, which stays put
// while the frames change width and once the Spinner stops. The rendition
// asks for frames at its fastest Spinner's rate, and none once all stop.
func TestSpinner(t *testing.T) {
	r, c, now := motion(t, `{"busy":true}`, `
	 {"id":"root","component":"Column","children":["dots","points"]},
	 {"id":"dots","component":"HottySpinner","catalogId":"HOTTY","label":"Dot","active":{"@path":"/busy"}},
	 {"id":"points","component":"HottySpinner","catalogId":"HOTTY","spinner":"points","label":"Points","active":{"@path":"/busy"}}`)
	dot, points := view.Spinners["dot"], view.Spinners["points"]
	for _, k := range []int64{0, 1, 9} {
		*now = time.Unix(0, k*int64(dot.Interval))
		pk := now.UnixNano() / int64(points.Interval)
		want := dot.Frames[k%int64(len(dot.Frames))] + " Dot\n" +
			points.Frames[pk%int64(len(points.Frames))] + " Points"
		if got := r.Draw(20).Plain(); got != want {
			t.Errorf("at %v: got\n%s\nwant\n%s", now.Sub(time.Unix(0, 0)), got, want)
		}
		if d := r.Animating(); d != dot.Interval {
			t.Errorf("animates every %v, want dot's %v", d, dot.Interval)
		}
	}
	if err := c.S.Write("/busy", false); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if got, want := r.Draw(20).Plain(), "  Dot\n    Points"; got != want {
		t.Errorf("stopped: got\n%s\nwant\n%s", got, want)
	}
	if d := r.Animating(); d != 0 {
		t.Errorf("stopped spinners animate every %v", d)
	}
}
