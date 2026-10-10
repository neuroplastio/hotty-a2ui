package icons

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

// TestBasicIsCurrent: the shapes are the 59 names of the basic catalog A2UI
// is pinned at, made from the package icons/REV pins. A new A2UI with new
// names, or a new pin without make icons, fails here.
func TestBasicIsCurrent(t *testing.T) {
	b, err := os.ReadFile("../third_party/a2ui/catalogs/basic/v1/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Components map[string]struct {
			Properties map[string]struct {
				OneOf []struct {
					Enum []string `json:"enum"`
				} `json:"oneOf"`
			} `json:"properties"`
		} `json:"components"`
	}
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, o := range c.Components["Icon"].Properties["name"].OneOf {
		want = append(want, o.Enum...)
	}
	if len(want) != 59 {
		t.Fatalf("A2UI's Icon has %d names, want 59", len(want))
	}
	slices.Sort(want)
	for _, m := range []struct {
		what string
		keys []string
	}{{"basic.go", keysOf(basic)}, {"Glyphs", keysOf(Glyphs)}} {
		if !slices.Equal(m.keys, want) {
			t.Errorf("%s has %v, A2UI %v (make icons)", m.what, m.keys, want)
		}
	}
	r, err := os.ReadFile("REV")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(r)); got != rev {
		t.Errorf("REV pins %q, basic.go was made from %q: make icons", got, rev)
	}
	for n := range basic {
		ic, ok := Basic(n)
		if !ok || ic.Box != 24 {
			t.Errorf("%s: %v %v", n, ic, ok)
		}
		if _, ok := Path(ic.Path); !ok {
			t.Errorf("%s isn't path data Path accepts", n)
		}
	}
	if _, ok := Basic("rocket"); ok {
		t.Error("a name of none of the 59 has a shape")
	}
}

func keysOf(m map[string]string) []string {
	k := make([]string, 0, len(m))
	for n := range m {
		k = append(k, n)
	}
	slices.Sort(k)
	return k
}

// TestPath: an svgPath is a shape only when it is path data and nothing
// else, and at most MaxPath bytes.
func TestPath(t *testing.T) {
	for _, c := range []struct {
		d  string
		ok bool
	}{
		{"M10 20v-6h4v6h5v-8h3L12 3L2 12h3v8z", true}, // the A2UI test fixture: Material's home
		{"m1.5-2e-3 1,2 A1 1 0 0 1 3 3\nZ", true},
		{"", false},
		{`M0 0"/><script>alert(1)</script>`, false},
		{"M0 0 url(#x)", false},
		{"M0 0 L１ 1", false}, // a fullwidth digit
		{"M0 0\x00", false},
		{strings.Repeat("M0 0", MaxPath/4), true},
		{strings.Repeat("M0 0", MaxPath/4) + " ", false},
	} {
		ic, ok := Path(c.d)
		if ok != c.ok || ok && (ic.Box != 24 || ic.Path != c.d) {
			t.Errorf("Path(%.40q) = %v, %v; want ok %v", c.d, ic, ok, c.ok)
		}
	}
}

func TestGlyph(t *testing.T) {
	if Glyph("home") != "⌂" || Glyph("rocket") != Unknown || Glyph("") != Unknown {
		t.Error(Glyph("home"), Glyph("rocket"), Glyph(""))
	}
}
