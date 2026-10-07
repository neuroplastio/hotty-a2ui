package cells

import (
	"slices"

	"github.com/clipperhouse/uax29/v2/graphemes"
)

//go:generate go run ./internal/gen

// A table is a set of code points, as sorted ranges.
type table [][2]rune

func (t table) has(r rune) bool {
	_, found := slices.BinarySearchFunc(t, r, func(x [2]rune, r rune) int {
		switch {
		case x[1] < r:
			return -1
		case x[0] > r:
			return 1
		}
		return 0
	})
	return found
}

const (
	zwj  = '\u200d'
	vs16 = '\ufe0f'
)

// runeWidth is a code point's width (profile §3.1): 0 for marks, format
// controls and variation selectors; 2 for East Asian Wide and Fullwidth,
// and pictographs with emoji presentation; else 1.
func runeWidth(r rune) int {
	switch {
	case r >= 0x20 && r < 0x7f:
		return 1
	case zero.has(r):
		return 0
	case wide.has(r):
		return 2
	}
	return 1
}

// clusterWidth is a grapheme cluster's width: its first non-zero code
// point's, or 2 when it holds VS16, is an emoji ZWJ sequence or is a flag;
// 0 when every code point is of width 0.
func clusterWidth(g string) int {
	first, w, n := rune(-1), 0, 0
	ri, zwjSeq := 0, false
	for _, r := range g {
		n++
		if first < 0 {
			first = r
		}
		switch {
		case r == vs16:
			return 2
		case r == zwj:
			zwjSeq = true
		case r >= 0x1F1E6 && r <= 0x1F1FF:
			ri++
		}
		if w == 0 {
			w = runeWidth(r)
		}
	}
	if zwjSeq && pictographic.has(first) || ri == 2 && n == 2 {
		return 2
	}
	return w
}

// Width is a string's width in cells: the sum of its grapheme clusters'.
func Width(s string) int {
	w := 0
	for g := graphemes.FromString(s); g.Next(); {
		w += clusterWidth(g.Value())
	}
	return w
}

// clusters splits s into grapheme clusters.
func clusters(s string) []string {
	var out []string
	for g := graphemes.FromString(s); g.Next(); {
		out = append(out, g.Value())
	}
	return out
}
