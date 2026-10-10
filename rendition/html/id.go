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
// "~" and a lower-case letter, which no escape has, and a number for one
// of several (a Slider's notches, "~k3"); the rendition's own elements
// are "~" and a letter alone, which no element's id is. A HottyMarkdown's
// heading is its document's id, "~#" and its GitHub ID, encoded so too
// ("guide~#install"; headingID), which no other id has (profile §6.27).
const (
	partWrap   = "w" // the box around a control, its label and its error
	partLabel  = "l"
	partError  = "e"
	partOutput = "v" // a Slider's value
	partRange  = "r" // a Slider's track, buttons and value
	partLess   = "m" // a Slider's − button
	partMore   = "n" // a Slider's + button
	partNotch  = "k" // a Slider's notches, "k0" to "kN": where a drag sets it
	partTabs   = "t" // a Tabs' bar
	partPanel  = "p" // a Tabs' content
	partSubmit = "s" // a Form's hidden submit button
	partList   = "x" // a select's open list; a text field's suggestions' datalist
	partOption = "o" // a select's options in its list, "o0" to "oN"
	partRow    = "y" // a Table's rows, "y0" to "yN", by their index
	partFrame  = "f" // a Spinner's frame; a timer's time: what a tick changes
	partTitle  = "h" // a HottyList's title, or its filter while typed
	partStatus = "u" // a HottyList's status line
	partItem   = "i" // a HottyList's items, "i0" to "iN", by their index
	partDots   = "d" // a HottyList's or a HottyPaginator's page dots; a HottyPaginator's each, "d0" to "dN"
	partLines  = "j" // a HottyScrollView's lines, which a log appends to
	partHunk   = "b" // a HottyDiff's hunks, "b0" to "bN", by their index
	partNode   = "q" // a HottyTree's nodes, "q0" to "qN", by their index
	partPlot   = "a" // a HottyChart's plot
	partSeries = "g" // a HottyChart's lines' svgs, "g0" to "gN", a series each
	partPath   = "c" // the line in each, "c0" to "cN": a new point is its d's delta; a HottyQRCode's dark modules
	partTip    = "z" // a HottyKeyHints' tooltip row
	surfaceID  = "~s"
	layerID    = "~o" // the overlay an open Modal or list shows in
	backdropID = "~d"
	dialogID   = "~g"
	dismissID  = "~c" // under an open list: a click there closes it
	popoverID  = "~p" // an open list's own surface (Rendition.Popover)
	toastsID   = "~t" // the toasts' region, over the layer (toasts)
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

// headingID is the DOM id of a HottyMarkdown's heading: its document's id,
// "~#", and the heading's ID (view.Element.Anchors).
func headingID(id, anchor string) string { return partID(id, "#"+domID(anchor)) }

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
		if i+1 < len(dom) && dom[i+1] == '#' {
			// A HottyMarkdown's heading: the part is "#" and its ID.
			if b.Len() == 0 {
				return "", "", false
			}
			return b.String(), dom[i+1:], true
		}
		if i+1 < len(dom) && 'a' <= dom[i+1] && dom[i+1] <= 'z' {
			part := dom[i+1:]
			if b.Len() == 0 || strings.Trim(part[1:], "0123456789") != "" {
				return "", "", false
			}
			return b.String(), part, true
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

// DOMID is the id of an element of the view in the document: what the
// host's events and a test's clicks name.
func DOMID(id string) string { return domID(id) }

// ViewID is the element of the view a DOM id names, itself or one of its
// parts: a HottyTree's node, which has the host's focus for it.
func ViewID(dom string) (id string, ok bool) {
	id, _, ok = viewID(dom)
	return id, ok
}
