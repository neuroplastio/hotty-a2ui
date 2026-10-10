package qr

import (
	"strings"
	"testing"

	"rsc.io/qr/coding"
)

// A code is the smallest version that holds the text at its level, 17 +
// 4·version modules a side, at the highest level that version holds it
// at: digits and capitals take fewer bits than bytes; the most a code
// holds is 2,953 bytes at L and 1,273 at H.
func TestEncodeSizes(t *testing.T) {
	for _, tc := range []struct {
		text     string
		level    Level
		version  int
		raisedTo Level
	}{
		{"12345", M, 1, H},
		{strings.Repeat("7", 41), L, 1, L},
		{strings.Repeat("7", 42), L, 2, Q},
		{"HELLO WORLD HELLO WO", M, 1, M},
		{"hello world hello wo", M, 2, Q},
		{"https://hotty.neuroplast.io", M, 3, Q},
		{"https://hotty.neuroplast.io", L, 2, L},
		{strings.Repeat("x", 2953), L, 40, L},
		{strings.Repeat("x", 1273), H, 40, H},
	} {
		c, err := Encode(tc.text, tc.level)
		if err != nil {
			t.Fatalf("%.20q at %v: %v", tc.text, tc.level, err)
		}
		if c.Version != tc.version || c.Size != 17+4*tc.version || c.Level != tc.raisedTo {
			t.Errorf("%.20q at %v: version %d, %d a side, at %v", tc.text, tc.level, c.Version, c.Size, c.Level)
		}
	}
	for _, tc := range []struct {
		n     int
		level Level
	}{{2954, L}, {2332, M}, {1664, Q}, {1274, H}} {
		if _, err := Encode(strings.Repeat("x", tc.n), tc.level); err != ErrTooLong {
			t.Errorf("%d bytes at %v: %v", tc.n, tc.level, err)
		}
	}
}

// A code has its finder patterns in three corners, a dark square in a
// light ring in a dark ring, and its timing patterns between them; the
// module outside it is light.
func TestEncodePatterns(t *testing.T) {
	c, err := Encode("https://hotty.neuroplast.io", Q)
	if err != nil {
		t.Fatal(err)
	}
	finder := []string{"#######", "#.....#", "#.###.#", "#.###.#", "#.###.#", "#.....#", "#######"}
	for _, at := range [][2]int{{0, 0}, {c.Size - 7, 0}, {0, c.Size - 7}} {
		for y, row := range finder {
			for x, m := range row {
				if c.Dark(at[0]+x, at[1]+y) != (m == '#') {
					t.Fatalf("finder at %v: module %d,%d", at, x, y)
				}
			}
		}
	}
	for i := 8; i < c.Size-8; i++ {
		if c.Dark(i, 6) != (i%2 == 0) || c.Dark(6, i) != (i%2 == 0) {
			t.Fatalf("timing at %d", i)
		}
	}
	if c.Dark(-1, 0) || c.Dark(0, c.Size) {
		t.Error("dark outside the code")
	}
}

// Of the eight masks, the code has the one the penalty rules score lowest,
// and the same text encodes the same code again.
func TestEncodeMask(t *testing.T) {
	for _, text := range []string{"12345", "https://hotty.neuroplast.io", strings.Repeat("The quick brown fox. ", 20)} {
		c, err := Encode(text, M)
		if err != nil {
			t.Fatal(err)
		}
		best := penalty(c)
		for m := range 8 {
			p, err := coding.NewPlan(coding.Version(c.Version), coding.Level(c.Level), coding.Mask(m))
			if err != nil {
				t.Fatal(err)
			}
			enc := coding.Encoding(coding.String(text))
			if coding.Num(text).Check() == nil {
				enc = coding.Num(text)
			}
			cc, err := p.Encode(enc)
			if err != nil {
				t.Fatal(err)
			}
			o := &Code{Size: cc.Size, dark: make([]bool, cc.Size*cc.Size)}
			for y := range cc.Size {
				for x := range cc.Size {
					o.dark[y*cc.Size+x] = cc.Black(x, y)
				}
			}
			if s := penalty(o); s < best {
				t.Errorf("%.20q: mask %d scores %d, below the chosen %d's %d", text, m, s, c.Mask, best)
			}
		}
		if again, _ := Encode(text, M); again.Mask != c.Mask || again.Version != c.Version {
			t.Errorf("%.20q again: %+v", text, again)
		}
	}
}

// The levels by their letters.
func TestParseLevel(t *testing.T) {
	for _, s := range []string{"L", "M", "Q", "H"} {
		if l, ok := ParseLevel(s); !ok || l.String() != s {
			t.Errorf("%s: %v %v", s, l, ok)
		}
	}
	if _, ok := ParseLevel("m"); ok {
		t.Error("m is a level")
	}
}

// penalty scores a pattern 1:1:3:1:1 with four light modules on either
// side, the quiet zone's included, at 40; runs of five and more at 3 and
// one more a module; 2×2 blocks at 3; a code all dark at 100 for its
// balance.
func TestPenalty(t *testing.T) {
	full := &Code{Size: 5, dark: make([]bool, 25)}
	for i := range full.dark {
		full.dark[i] = true
	}
	// Rows and columns: 10 runs of 5 at 3; 16 blocks at 3; balance 10·10.
	if got := penalty(full); got != 10*3+16*3+100 {
		t.Errorf("all dark: %d", got)
	}
}
