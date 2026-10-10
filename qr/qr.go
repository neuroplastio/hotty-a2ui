// Package qr encodes text as a QR code (ISO/IEC 18004) for HottyQRCode
// (profile §6.29). It is rsc.io/qr's coding (BSD-3-Clause), with the mask
// chosen by the standard's penalty rules, which rsc.io/qr leaves at 0.
package qr

import (
	"errors"
	"sync"

	"rsc.io/qr/coding"
)

// Level is how much of a code may be lost and still read: L about 7%, M
// 15%, Q 25%, H 30%.
type Level byte

// The levels.
const (
	L Level = iota
	M
	Q
	H
)

// ParseLevel is the level named by "L", "M", "Q" or "H"; ok is false for
// anything else.
func ParseLevel(s string) (Level, bool) {
	switch s {
	case "L":
		return L, true
	case "M":
		return M, true
	case "Q":
		return Q, true
	case "H":
		return H, true
	}
	return M, false
}

func (l Level) String() string { return "LMQH"[l : l+1] }

// Quiet is the light margin a code needs on each side, in modules.
const Quiet = 4

// ErrTooLong is the error for text that no code holds at its level: more
// than 2,953 bytes at L, 2,331 at M, 1,663 at Q, 1,273 at H (fewer where
// it is not all digits or capitals).
var ErrTooLong = errors.New("qr: too long for a QR code")

// Code is a QR code: Size modules a side, without its quiet zone; Level,
// the level it has, which may be above the one asked for (Encode).
type Code struct {
	Size    int
	Version int
	Level   Level
	Mask    int
	dark    []bool
}

// Dark reports whether the module at (x, y) is dark; outside the code (its
// quiet zone) it is not.
func (c *Code) Dark(x, y int) bool {
	return x >= 0 && y >= 0 && x < c.Size && y < c.Size && c.dark[y*c.Size+x]
}

// Encode encodes text at a level in the smallest version that holds it:
// as digits, else as capitals and the few marks QR's alphanumeric mode
// has, else as its UTF-8 bytes; at the highest level that version holds
// it at, which is never below the level asked for, as OpenTUI's encoder
// and Nayuki's raise it, since that costs no size; of the eight masks,
// the one the standard's penalty rules score lowest.
func Encode(text string, level Level) (*Code, error) {
	key := cacheKey{text, level}
	cache.Lock()
	c, ok := cache.codes[key]
	cache.Unlock()
	if ok {
		return c, nil
	}
	c, err := encode(text, level)
	if err != nil {
		return nil, err
	}
	cache.Lock()
	if len(cache.codes) >= cacheSize {
		clear(cache.codes)
	}
	cache.codes[key] = c
	cache.Unlock()
	return c, nil
}

// The codes encoded last, since a view encodes its codes again at every
// rebuild.
const cacheSize = 32

type cacheKey struct {
	text  string
	level Level
}

var cache = struct {
	sync.Mutex
	codes map[cacheKey]*Code
}{codes: map[cacheKey]*Code{}}

func encode(text string, level Level) (*Code, error) {
	var enc coding.Encoding
	switch {
	case coding.Num(text).Check() == nil:
		enc = coding.Num(text)
	case coding.Alpha(text).Check() == nil:
		enc = coding.Alpha(text)
	default:
		enc = coding.String(text)
	}
	l := coding.Level(level)
	v := coding.Version(coding.MinVersion)
	for enc.Bits(v) > v.DataBytes(l)*8 {
		if v++; v > coding.MaxVersion {
			return nil, ErrTooLong
		}
	}
	for l < coding.H && enc.Bits(v) <= v.DataBytes(l+1)*8 {
		l++
	}
	level = Level(l)
	var best *Code
	score := 0
	for m := range 8 {
		p, err := coding.NewPlan(v, l, coding.Mask(m))
		if err != nil {
			return nil, err
		}
		cc, err := p.Encode(enc)
		if err != nil {
			return nil, err
		}
		c := &Code{Size: cc.Size, Version: int(v), Level: level, Mask: m, dark: make([]bool, cc.Size*cc.Size)}
		for y := range cc.Size {
			for x := range cc.Size {
				c.dark[y*cc.Size+x] = cc.Black(x, y)
			}
		}
		if s := penalty(c); best == nil || s < score {
			best, score = c, s
		}
	}
	return best, nil
}

// penalty is a code's score by ISO/IEC 18004's rules for choosing a mask
// (§7.8.3): runs of five or more modules of one colour in a row or a
// column, 2×2 blocks of one colour, finder-like 1:1:3:1:1 patterns with
// four light modules beside them, and how far the dark modules are from
// half.
func penalty(c *Code) int {
	n := c.Size
	s := 0
	for _, col := range []bool{false, true} {
		at := func(i, j int) bool {
			if col {
				return c.Dark(i, j)
			}
			return c.Dark(j, i)
		}
		for i := range n {
			run := 0
			for j := range n {
				if j > 0 && at(i, j) == at(i, j-1) {
					run++
				} else {
					run = 1
				}
				switch {
				case run == 5:
					s += 3
				case run > 5:
					s++
				}
			}
			// The quiet zone is light past each end.
			for j := -Quiet; j < n+Quiet-10; j++ {
				if finderLike(func(k int) bool { return at(i, j+k) }) {
					s += 40
				}
			}
		}
	}
	dark := 0
	for y := range n {
		for x := range n {
			d := c.Dark(x, y)
			if d {
				dark++
			}
			if x+1 < n && y+1 < n && d == c.Dark(x+1, y) && d == c.Dark(x, y+1) && d == c.Dark(x+1, y+1) {
				s += 3
			}
		}
	}
	pct := dark * 100 / (n * n)
	return s + 10*(abs(pct-50)/5)
}

// finderLike reports whether the eleven modules from 0 are dark, light,
// dark three, light, dark and four light, or the same the other way round.
func finderLike(at func(int) bool) bool {
	const a, b = "10111010000", "00001011101"
	fwd, back := true, true
	for k := range 11 {
		d := at(k)
		fwd = fwd && d == (a[k] == '1')
		back = back && d == (b[k] == '1')
	}
	return fwd || back
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
