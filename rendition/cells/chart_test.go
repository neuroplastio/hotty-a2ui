package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// chartOf is a rendition of a surface whose root is the given hotty
// component, a JSON object without its id and catalogId, over a data
// model.
func chartOf(t *testing.T, data, comp string) (*Rendition, *a2ui.Processor) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":` + data + `}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","catalogId":"` + hotty.ID + `",` + comp + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return New(view.NewController(p.Surface("s"))), p
}

// A line chart is its ticks' labels, the axis, braille dots a series in
// its colour (info, then warning), a gap where a value is missing, a few
// points' labels and the legend. Where the plot is too short to label
// every tick, the zero, the top and the bottom come first.
func TestLineChartDraws(t *testing.T) {
	r, p := chartOf(t, `{"b":[8,null,4,2,0]}`, `"component":"HottyChart","series":[{"label":"a","values":[0,2,4,6,8]},{"label":"b","values":{"@path":"/b"}}],"labels":["mo","tu","we","th","fr"],"height":4`)
	want := "10 ┤\n" +
		"   │⠁                ⢀⣀⣀⡠⠤⠔⠒⠊⠉\n" +
		"   │        ⣀⣀⠤⠤⠒⠒⠮⠭⣉⣁\n" +
		" 0 ┤⠤⠤⠒⠒⠒⠉⠉⠉          ⠉⠉⠉⠒⠒⠒⠤⠤\n" +
		"   └──────────────────────────\n" +
		"    mo          we          fr\n" +
		"    ━ a   ━ b"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, c := range []struct {
		row, col int
		role     Role
	}{
		{0, 0, Muted}, {0, 3, Border}, {4, 3, Border}, {5, 4, Muted},
		{3, 4, Info}, {1, 4, Warning}, {6, 4, Info}, {6, 6, Muted}, {6, 10, Warning},
	} {
		if got := f.Cells[c.row][c.col]; got.Role != c.role {
			t.Errorf("row %d, column %d (%q): %v, want %v", c.row, c.col, got.Text, got.Role, c.role)
		}
	}
	// The agent writes b's missing value: the gap closes.
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/b/1","value":6}}]`)); err != nil {
		t.Fatal(err)
	}
	r.c.Rebuild()
	if got := r.Draw(30).Cells[1][4:8]; got[0].Text == "⠁" && got[1].Text == " " {
		t.Errorf("the gap stays: %q", r.Draw(30).Plain())
	}
}

// Bars rise from the axis's 0, which may fall inside a row: eighths of a
// row from a cell's bottom, halves (▀) where a bar ends in the middle of
// one, a bar below 0 hanging down; each label under its bars.
func TestBarChartDraws(t *testing.T) {
	r, _ := chartOf(t, `{}`, `"component":"HottyChart","kind":"bar","values":[3,-1,2],"labels":["a","b","c"],"height":4`)
	want := " 4 ┤▃▃▃▃▃▃▃▃\n" +
		"   │████████          ▅▅▅▅▅▅▅▅\n" +
		" 0 ┤▀▀▀▀▀▀▀▀ ▃▃▃▃▃▃▃▃ ▀▀▀▀▀▀▀▀\n" +
		"-2 ┤         ▀▀▀▀▀▀▀▀\n" +
		"   └──────────────────────────\n" +
		"       a        b        c"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if bar, label := f.Cells[1][4], f.Cells[5][7]; bar.Role != Info || label.Role != Muted {
		t.Errorf("bar %v, label %v", bar.Role, label.Role)
	}
}

// Several series' bars stand side by side in a group a point, each in its
// series' colour; where the groups do not fit, the last points do.
func TestGroupedBarsDraw(t *testing.T) {
	r, _ := chartOf(t, `{}`, `"component":"HottyChart","kind":"bar","series":[{"label":"x","values":[1,2,3,4,5,6,7,8]},{"label":"y","values":[8,7,6,5,4,3,2,1]}],"height":2`)
	want := "10 ┤ ▅   ▃   ▂          ▂   ▃   ▅\n" +
		" 0 ┤▂█  ▃█  ▅█  ▆█  █▆  █▅  █▃  █▂\n" +
		"   └────────────────────────────────────\n" +
		"    ■ x   ■ y"
	f := r.Draw(40)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if x, y, kx, ky := f.Cells[1][4], f.Cells[1][5], f.Cells[3][4], f.Cells[3][10]; x.Role != Info || y.Role != Warning || kx.Role != Info || ky.Role != Warning {
		t.Errorf("x %v, y %v, legend %v %v", x.Role, y.Role, kx.Role, ky.Role)
	}
	want = "10 ┤   ▂  ▃  ▅\n" +
		" 0 ┤█▆ █▅ █▃ █▂\n" +
		"   └────────────\n" +
		"    ■ x   ■ y"
	if got := r.Draw(16).Plain(); got != want {
		t.Errorf("16 wide: got\n%s\nwant\n%s", got, want)
	}
}

// A chart with no values keeps its axes and says so; it draws as wide as
// its row lets it.
func TestEmptyChartDraws(t *testing.T) {
	r, _ := chartOf(t, `{}`, `"component":"HottyChart","values":{"@path":"/none"},"height":3`)
	want := "  │\n" +
		"  │          No data\n" +
		"  │\n" +
		"  └───────────────────────────"
	if got := r.Draw(30).Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

// A sparkline is a column a value in eighths of a row, in info, the
// lowest an eighth and a missing value blank; in a window, the last
// values, the newest at its right edge.
func TestSparklineDraws(t *testing.T) {
	for _, c := range []struct{ comp, want string }{
		{`"values":[0,1,2,3,4,null,8]`, "▁▁▂▃▄ █"},
		{`"values":[1,2,3,4,5,6,7,8],"height":2,"window":5`, " ▂▄▆█\n█████"},
		{`"values":[-2,0,2],"window":5`, "  ▁▄█"},
		{`"values":[],"window":5`, ""},
	} {
		r, _ := chartOf(t, `{}`, `"component":"HottySparkline",`+c.comp)
		f := r.Draw(30)
		if got := f.Plain(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.comp, got, c.want)
			continue
		}
		for _, row := range f.Cells {
			for _, cell := range row {
				if cell.Text != " " && cell.Text != "" && cell.Role != Info {
					t.Errorf("%s: %q in %v", c.comp, cell.Text, cell.Role)
				}
			}
		}
	}
}
