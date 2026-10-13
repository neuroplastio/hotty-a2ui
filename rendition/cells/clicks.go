package cells

import (
	"time"

	"github.com/clipperhouse/uax29/v2/words"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Repeated clicks: a double click in a text control selects the word under
// it, a triple click its line (profile §3.5). A terminal's mouse reports
// carry no count, and no HOTTY host adds one to them (only a surface's
// clicks have one), so the rendition counts the presses itself (clickRun).

// clickGap is how soon after a press the next must come to be part of its
// run: the double-click time most platforms ship with.
const clickGap = 500 * time.Millisecond

// A clickRun is a run of presses on one cell, each within clickGap of the
// one before: a double click is a run of two.
type clickRun struct {
	col, row, n int
	at          time.Time
}

// next is the run a press at (col, row) at t makes: this one and the
// press, on the same cell and soon enough; else a new run of one.
func (c clickRun) next(col, row int, t time.Time) clickRun {
	if c.n > 0 && col == c.col && row == c.row && t.Sub(c.at) <= clickGap {
		return clickRun{col, row, c.n + 1, t}
	}
	return clickRun{col, row, 1, t}
}

// selectRun selects in a text control what the press at character p of
// its value selects as the n-th of its run (n ≥ 2): the word there on a
// second (wordAt), its line on a third and after (lineAt). The caret goes
// to the selection's end, and the button held drags nothing.
func (r *Rendition) selectRun(e *view.Element, p, n int) {
	v, _ := e.Value.(string)
	start, end := wordAt(v, p, e.Variant == "obscured")
	if n >= 3 {
		start, end = lineAt(clusters(v), p)
	}
	r.field(e).Select(start, end)
	r.cursor[e.ID] = end
	r.drag = ""
	r.wake()
	r.primary(e)
}

// wordAt is the word of a value at character (cluster) p, as a browser's
// field takes one on a double click: a segment of UAX #29's words, a run
// of letters and digits, of spaces, or a mark. Past a line's end it is
// the line's last; on an empty line, nothing at p. An obscured value is
// one word, as a host's password field is.
func wordAt(v string, p int, obscured bool) (start, end int) {
	cl := clusters(v)
	if obscured {
		return 0, len(cl)
	}
	if p >= len(cl) || isBreak(cl[p]) {
		if p == 0 || isBreak(cl[p-1]) {
			return p, p
		}
		p--
	}
	at := 0
	for w := words.FromString(v); w.Next(); {
		n := len(clusters(w.Value()))
		if p < at+n {
			return at, at + n
		}
		at += n
	}
	return p, p
}

// lineAt is the line of a value's characters at p, without its break: a
// triple click's selection. A one-line value is one line.
func lineAt(cl []string, p int) (start, end int) {
	start = min(p, len(cl))
	end = start
	for start > 0 && !isBreak(cl[start-1]) {
		start--
	}
	for end < len(cl) && !isBreak(cl[end]) {
		end++
	}
	return start, end
}
