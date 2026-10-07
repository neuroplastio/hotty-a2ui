package html

import (
	"strconv"
	"strings"
)

// An element's DOM id is its view id, encoded for HOTTY: a control value
// is printable ASCII without ':', ';' and '=' (SPEC §3.2), and a node's
// key has brackets and '#' ("row[/items/0]#2"). Letters, digits and
// "-_./" stay; every other byte is "~" and two upper-case hex digits.
// The parts an element draws besides itself (its label, its error) add
// "~" and a lower-case letter, which no escape has; the rendition's own
// elements are "~" and a letter alone, which no element's id is.
const (
	partWrap   = "w" // the box around a control, its label and its error
	partLabel  = "l"
	partError  = "e"
	partOutput = "v" // a Slider's value
	partRange  = "r" // a Slider's track and value
	partTabs   = "t" // a Tabs' bar
	partPanel  = "p" // a Tabs' content
	partSubmit = "s" // a Form's hidden submit button
	surfaceID  = "~s"
	layerID    = "~o" // the overlay an open Modal shows its content in
	backdropID = "~d"
	dialogID   = "~g"
)

func domID(id string) string {
	var b strings.Builder
	for i := 0; i < len(id); i++ {
		c := id[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte("-_./", c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteString("~" + strings.ToUpper(strconv.FormatUint(uint64(c)|0x100, 16)[1:]))
	}
	return b.String()
}

func partID(id, part string) string { return domID(id) + "~" + part }

// viewID reads a DOM id: the element's view id, and the part, if the id
// is one of its parts'. ok is false for an id the rendition did not make.
func viewID(dom string) (id, part string, ok bool) {
	var b strings.Builder
	for i := 0; i < len(dom); i++ {
		c := dom[i]
		if c != '~' {
			b.WriteByte(c)
			continue
		}
		if i+1 < len(dom) && 'a' <= dom[i+1] && dom[i+1] <= 'z' {
			if i+2 != len(dom) || b.Len() == 0 {
				return "", "", false
			}
			return b.String(), dom[i+1:], true
		}
		if i+2 >= len(dom) {
			return "", "", false
		}
		v, err := strconv.ParseUint(dom[i+1:i+3], 16, 8)
		if err != nil || strings.ToUpper(dom[i+1:i+3]) != dom[i+1:i+3] {
			return "", "", false
		}
		b.WriteByte(byte(v))
		i += 2
	}
	return b.String(), "", b.Len() > 0
}
