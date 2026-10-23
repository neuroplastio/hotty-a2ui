package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// capTint is how far a cap's fill is from the background toward the
// text: lighter than a Button's, so that a line of caps reads as keys and
// not as buttons.
const capTint = 36

var (
	capStyle = style{back: Fg, backMix: capTint, backShade: true}
	capPlus  = style{attr: Faint, back: Fg, backMix: capTint, backShade: true}
	capSep   = style{role: Muted}
)

// caps are a hint's keys (view.Controller.Caps, KIT-08h) as cells draw
// them: each chord a cap, a column of room either side of its labels and
// its background tinted toward the text, else the terminal's bright black
// (as a Button's fill); its labels joined by its faint separator ("+",
// say) unless they sit tight (⌘⇧P); a muted ">" between a sequence's chords and "/" between
// alternatives. Where on is set (a selected row's reverse), the caps are
// drawn in it, untinted, as tint and reverse don't mix.
func (r *Rendition) caps(keys string, on *style) []glyph {
	return capsOf(r.c.Caps(keys, r.c.Keycaps.Glyphs(false)), on)
}

// keyCaps are a HottyKeys' caps, each chord read with the Platform.
func (r *Rendition) keyCaps(e *view.Element) []glyph {
	return r.caps(view.KeysText(e.Label, r.keyText), nil)
}

// capsOf draws chords as caps (Rendition.caps).
func capsOf(alts [][]view.Chord, on *style) []glyph {
	body, plus, sep := capStyle, capPlus, capSep
	if on != nil {
		body, plus, sep = *on, *on, *on
	}
	var out []glyph
	for i, seq := range alts {
		if i > 0 {
			out = concat(out, line("/", sep))
		}
		for j, ch := range seq {
			if j > 0 {
				out = concat(out, line(">", sep))
			}
			out = concat(out, line(" ", body))
			for k, l := range ch.Labels {
				if ch.Joined(k) {
					out = concat(out, line(ch.Sep, plus))
				}
				out = concat(out, line(l, body))
			}
			out = concat(out, line(" ", body))
		}
	}
	return out
}
