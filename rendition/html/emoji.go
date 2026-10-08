package html

import (
	"strings"

	"github.com/clipperhouse/uax29/v2/graphemes"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
)

// Emoji go in a span of class k-emoji, which names a colour emoji font. A
// host's font fallback may otherwise draw them from a text font that has
// the code point, in outline: Blitz does, VS16 or not. Only the emoji go
// in the span, so the spaces between them keep the text's font.

// texts is s as nodes: its text, each emoji (cells.Emoji) in a span.
func texts(s string) []*node {
	var out []*node
	var run strings.Builder
	for g := graphemes.FromString(s); g.Next(); {
		v := g.Value()
		if !cells.Emoji(v) {
			run.WriteString(v)
			continue
		}
		if run.Len() > 0 {
			out = append(out, txt(run.String()))
			run.Reset()
		}
		out = append(out, el("span", "class", "k-emoji").add(txt(v)))
	}
	if run.Len() > 0 || len(out) == 0 {
		out = append(out, txt(run.String()))
	}
	return out
}

// emojiHTML is HTML with the emoji in its text in spans, as texts makes
// them: what is between tags is text, and an emoji is never escaped.
func emojiHTML(h string) string {
	var b strings.Builder
	for len(h) > 0 {
		lt := strings.IndexByte(h, '<')
		if lt < 0 {
			lt = len(h)
		}
		b.WriteString(emojiSpans(h[:lt]))
		h = h[lt:]
		if gt := strings.IndexByte(h, '>'); gt >= 0 {
			b.WriteString(h[:gt+1])
			h = h[gt+1:]
		} else {
			b.WriteString(h)
			h = ""
		}
	}
	return b.String()
}

func emojiSpans(text string) string {
	if !hasEmoji(text) {
		return text
	}
	var b strings.Builder
	for g := graphemes.FromString(text); g.Next(); {
		if v := g.Value(); cells.Emoji(v) {
			b.WriteString(`<span class="k-emoji">` + v + `</span>`)
		} else {
			b.WriteString(v)
		}
	}
	return b.String()
}

// hasEmoji is a quick test: an emoji is never ASCII.
func hasEmoji(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			for g := graphemes.FromString(s[i:]); g.Next(); {
				if cells.Emoji(g.Value()) {
					return true
				}
			}
			return false
		}
	}
	return false
}
