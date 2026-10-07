package cells

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/story"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/neuroplastio/hotty-a2ui/view"
)

var update = flag.Bool("update", false, "write the golden grids in testdata")

const examples = "a2ui/catalogs/basic/v1/examples/"

// example is a basic catalog example's first surface, as a controller.
func example(t *testing.T, name string) *view.Controller {
	t.Helper()
	b, err := fs.ReadFile(thirdparty.BasicExamples, examples+name+".json")
	if err != nil {
		t.Fatal(err)
	}
	var ex struct{ Messages []any }
	if err := json.Unmarshal(b, &ex); err != nil {
		t.Fatal(err)
	}
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	if err := p.Process(ex.Messages); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return view.NewController(p.Surfaces()[0])
}

// openModal opens the first Modal in the view, as a click on its trigger
// would.
func openModal(t *testing.T, c *view.Controller) {
	t.Helper()
	var trigger string
	var first func(es []*view.Element)
	first = func(es []*view.Element) {
		for _, e := range es {
			if trigger == "" && e.Focusable() {
				trigger = e.ID
			}
			first(e.Children)
		}
	}
	c.V.Walk(func(e *view.Element) bool {
		if e.Kind != view.Modal {
			return true
		}
		if trigger = ""; e.Clickable {
			trigger = e.ID
		} else {
			first(e.Children)
		}
		return false
	})
	if trigger == "" {
		t.Fatal("no Modal")
	}
	if err := c.Activate(trigger); err != nil {
		t.Fatal(err)
	}
	if c.V.Overlay == nil {
		t.Fatal("the Modal did not open")
	}
}

func TestGolden(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*testing.T, *view.Controller)
	}{
		{"00_simple-login-form", nil},
		{"00_row-layout", nil},
		{"07_task-card", nil},
		{"36_modal", openModal},
		{"34_child-list-template", nil},
	} {
		for _, cols := range []int{60, 40} {
			t.Run(fmt.Sprintf("%s/%d", tc.name, cols), func(t *testing.T) {
				c := example(t, tc.name)
				if tc.setup != nil {
					tc.setup(t, c)
				}
				got := New(c).Draw(cols).Plain() + "\n"
				path := filepath.Join("testdata", fmt.Sprintf("%s.%d.golden", tc.name, cols))
				if *update {
					if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
						t.Fatal(err)
					}
					return
				}
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("%v (go test -update writes it)", err)
				}
				if got != string(want) {
					t.Errorf("got\n%s\nwant\n%s", got, want)
				}
			})
		}
	}
}

// checkFrame checks a frame's grid: every row Cols cells, each wide cell
// followed by its second half, no row wider than Cols.
func checkFrame(t *testing.T, f *Frame) {
	t.Helper()
	if len(f.Cells) != f.Rows {
		t.Fatalf("%d rows, Rows %d", len(f.Cells), f.Rows)
	}
	for y, row := range f.Cells {
		if len(row) != f.Cols {
			t.Fatalf("row %d: %d cells, Cols %d", y, len(row), f.Cols)
		}
		for x := 0; x < len(row); x++ {
			c := row[x]
			switch c.Width {
			case 1:
				if c.Text == "" {
					t.Fatalf("row %d col %d: empty cell of width 1", y, x)
				}
			case 2:
				if x+1 >= len(row) || row[x+1].Width != 0 || row[x+1].Text != "" {
					t.Fatalf("row %d col %d: wide cell without its second half", y, x)
				}
				x++
			default:
				t.Fatalf("row %d col %d: cell %+v", y, x, c)
			}
		}
	}
	for y, l := range strings.Split(f.Plain(), "\n") {
		if Width(l) > f.Cols {
			t.Fatalf("row %d is %d wide: %q", y, Width(l), l)
		}
	}
}

func TestExamplesDraw(t *testing.T) {
	names, _ := fs.Glob(thirdparty.BasicExamples, examples+"*.json")
	if len(names) != 43 {
		t.Fatalf("%d examples", len(names))
	}
	for _, name := range names {
		name = strings.TrimSuffix(filepath.Base(name), ".json")
		for _, cols := range []int{80, 40} {
			c := example(t, name)
			f := New(c).Draw(cols)
			checkFrame(t, f)
			if f.Rows == 0 {
				t.Errorf("%s at %d: no rows", name, cols)
			}
			_ = f.ANSI(true)
			_ = f.ANSI(false)
		}
	}
}

// TestStoriesDraw draws every story's surfaces, then each again with the
// keyboard on each of its focusable elements in turn.
func TestStoriesDraw(t *testing.T) {
	for _, st := range story.All() {
		run, err := story.Start(st)
		if err != nil {
			t.Fatalf("%s: %v", st.Name, err)
		}
		for _, s := range run.Surfaces() {
			r := New(s.C)
			for _, cols := range []int{80, 40, 12} {
				checkFrame(t, r.Draw(cols))
			}
			for _, id := range s.C.V.Focusables() {
				s.C.Focus(id)
				checkFrame(t, r.Draw(40))
				if _, _, _, _, ok := r.Box(id); !ok {
					t.Errorf("%s: %s has the keyboard and no box", st.Name, id)
				}
			}
		}
	}
}
