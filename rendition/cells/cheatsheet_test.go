package cells

import (
	"fmt"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
	hottygo "github.com/neuroplastio/hotty-go"
)

// sheet is a cheat sheet: a Column of Rows, each a HottyKeys and a Text,
// weighted alike.
func sheet(t *testing.T, rows ...[2]string) (*Rendition, *view.Controller) {
	t.Helper()
	var ids, comps []string
	for i, r := range rows {
		id := fmt.Sprint("r", i)
		ids = append(ids, `"`+id+`"`)
		comps = append(comps, `{"id":"`+id+`","component":"Row","children":["`+id+`k","`+id+`t"]}`,
			`{"id":"`+id+`k","component":"HottyKeys","catalogId":"`+hotty.ID+`","keys":"`+r[0]+`","weight":1}`,
			`{"id":"`+id+`t","component":"Text","text":"`+r[1]+`","weight":1}`)
	}
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Column","children":[` + strings.Join(ids, ",") + `]},` + strings.Join(comps, ",") + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c
}

// A HottyKeys is its caps, as a hint's: a cap a chord, ">" between a
// sequence's, "/" between alternatives, read with the Platform. Rows
// weighted alike line up, the descriptions in one column.
func TestKeysDraw(t *testing.T) {
	r, c := sheet(t, [2]string{"ctrl+b v", "Split right"}, [2]string{"ctrl+b w t", "New tab"},
		[2]string{"mod+shift+p", "Palette"}, [2]string{"arrow_up/k", "Up"})
	f := r.Draw(40)
	want := []string{
		fmt.Sprintf("%-21s%s", " Ctrl+B > V ", "Split right"),
		fmt.Sprintf("%-21s%s", " Ctrl+B > W > T ", "New tab"),
		fmt.Sprintf("%-21s%s", " Ctrl+Shift+P ", "Palette"),
		fmt.Sprintf("%-21s%s", " ↑ / K ", "Up"),
	}
	if got := f.Plain(); got != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", got, strings.Join(want, "\n"))
	}
	row := f.Cells[0]
	for i, want := range map[int]style{0: capStyle, 1: capStyle, 5: capPlus, 8: capSep, 10: capStyle} {
		if got := row[i]; got.Role != want.role || got.Attr != want.attr || got.BackMix != want.backMix {
			t.Errorf("column %d (%q): %v %v %d, want %v %v %d", i, got.Text, got.Role, got.Attr, got.BackMix, want.role, want.attr, want.backMix)
		}
	}
	c.Platform = hottygo.Platform{Mod: "super"}
	if got := strings.Split(r.Draw(40).Plain(), "\n")[2]; got != fmt.Sprintf("%-21s%s", " Command-Shift-P ", "Palette") {
		t.Errorf("on a Mac: %q", got)
	}
}

// A HottyKeys is as wide as its caps, which it does not cut: narrower than
// them, the Text beside it gives up its columns.
func TestKeysWidth(t *testing.T) {
	r, _ := sheet(t, [2]string{"ctrl+b w t", "New tab"})
	if got := r.Draw(24).Plain(); got != " Ctrl+B > W > T  New tab" {
		t.Errorf("at 24 columns: %q", got)
	}
}
