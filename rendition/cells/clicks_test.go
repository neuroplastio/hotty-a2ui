package cells

import (
	"testing"
	"time"
)

func TestWordAt(t *testing.T) {
	for _, tc := range []struct {
		v          string
		p          int
		start, end int
	}{
		{"hello, wide world", 0, 0, 5},     // a word
		{"hello, wide world", 4, 0, 5},     // its last letter
		{"hello, wide world", 5, 5, 6},     // a mark
		{"hello, wide world", 6, 6, 7},     // a space
		{"hello, wide world", 9, 7, 11},    // "wide"
		{"hello, wide world", 17, 12, 17},  // past the end: the last word
		{"a  b", 2, 1, 3},                  // a run of spaces
		{"café au lait", 3, 0, 4},         // é is one character
		{"one two\nthree four", 7, 4, 7},   // past a line's end: its last word
		{"one two\nthree four", 10, 8, 13}, // "three"
		{"one\n\ntwo", 4, 4, 4},            // an empty line: nothing
		{"", 0, 0, 0},
	} {
		if start, end := wordAt(tc.v, tc.p, false); start != tc.start || end != tc.end {
			t.Errorf("wordAt(%q, %d) = %d, %d; want %d, %d", tc.v, tc.p, start, end, tc.start, tc.end)
		}
	}
	if start, end := wordAt("hunter2 x", 2, true); start != 0 || end != 9 {
		t.Errorf("obscured: %d, %d; want the whole value", start, end)
	}
}

func TestLineAt(t *testing.T) {
	cl := clusters("one two\nthree four")
	for p, want := range map[int][2]int{0: {0, 7}, 7: {0, 7}, 8: {8, 18}, 18: {8, 18}} {
		if start, end := lineAt(cl, p); start != want[0] || end != want[1] {
			t.Errorf("lineAt(%d) = %d, %d; want %v", p, start, end, want)
		}
	}
}

func TestClickRun(t *testing.T) {
	at := time.Unix(0, 0)
	var c clickRun
	for i, tc := range []struct {
		col, row int
		after    time.Duration
		n        int
	}{
		{3, 1, 0, 1},
		{3, 1, 200 * time.Millisecond, 2},
		{3, 1, clickGap, 3},
		{3, 1, clickGap + time.Millisecond, 1}, // too late
		{4, 1, 100 * time.Millisecond, 1},      // another cell
		{4, 1, 100 * time.Millisecond, 2},
	} {
		at = at.Add(tc.after)
		if c = c.next(tc.col, tc.row, at); c.n != tc.n {
			t.Errorf("press %d: n = %d, want %d", i, c.n, tc.n)
		}
	}
}

// TestDoubleClickSelects: in a field, a second press on the same cell
// selects the word there, a third the line, and one that comes too late
// only puts the caret there; ClickCount takes a host's count as it is.
func TestDoubleClickSelects(t *testing.T) {
	c, _, _ := form(t)
	if err := c.S.Write("/name", "hello wide world"); err != nil {
		t.Fatal(err)
	}
	if err := c.S.Write("/note", "one two\nthree four"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	r := New(c)
	now := time.Unix(0, 0)
	r.Clock = func() time.Time { return now }
	r.Draw(40)
	cell := func(id string, line, p int) (col, row int) {
		for _, h := range r.hits {
			if h.id == id && h.field != nil {
				return h.field.x + p, h.field.y + line
			}
		}
		t.Fatalf("no field %s", id)
		return
	}
	sel := func(id string) [2]int {
		start, end := r.field(c.V.Find(id)).Selection()
		return [2]int{start, end}
	}
	press := func(col, row int, after time.Duration) {
		now = now.Add(after)
		if err := r.Click(col, row); err != nil {
			t.Fatal(err)
		}
		r.Draw(40)
	}

	col, row := cell("name", 0, 8) // in "wide"
	press(col, row, 0)
	press(col, row, 150*time.Millisecond)
	if got := sel("name"); got != [2]int{6, 10} || r.cursor["name"] != 10 {
		t.Errorf("double click: %v, caret %d; want [6 10], 10", got, r.cursor["name"])
	}
	press(col, row, 150*time.Millisecond)
	if got := sel("name"); got != [2]int{0, 16} {
		t.Errorf("triple click: %v, want the line [0 16]", got)
	}
	press(col, row, time.Second)
	if got := sel("name"); got[0] != got[1] || r.cursor["name"] != 8 {
		t.Errorf("a late press: %v, caret %d; want none, 8", got, r.cursor["name"])
	}

	col, row = cell("note", 1, 2) // in "three"
	press(col, row, time.Second)
	press(col, row, 100*time.Millisecond)
	if got := sel("note"); got != [2]int{8, 13} {
		t.Errorf("double click in a long text: %v, want [8 13]", got)
	}
	press(col, row, 100*time.Millisecond)
	if got := sel("note"); got != [2]int{8, 18} {
		t.Errorf("triple click in a long text: %v, want its line [8 18]", got)
	}

	col, row = cell("name", 0, 1) // in "hello"
	if err := r.ClickCount(col, row, 2); err != nil {
		t.Fatal(err)
	}
	if got := sel("name"); got != [2]int{0, 5} {
		t.Errorf("a host's double click: %v, want [0 5]", got)
	}
}
