package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// code is a HottyCode on a host (profile §2): a box in the terminal's
// font, filled as a Text's code block is, a row for each line: its number
// (aria-hidden: the code is what a screen reader reads), its mark's sign,
// then its code, a span for each token but plain ones, coloured by the
// kit's sheet (view.TokenClass). A marked line is tinted; lines that do
// not wrap make the box scroll sideways, which the host does itself
// (SPEC §5.3).
func (m *markup) code(e *view.Element) *node {
	class := "k-code"
	if !e.Wrap {
		class += " k-nowrap"
	}
	n := el("div", "id", domID(e.ID), "class", class)
	if e.Lang != "" {
		n.set("data-lang", e.Lang)
	}
	digits := 0
	if e.Numbers {
		digits = max(len(strconv.Itoa(e.FirstLine)), len(strconv.Itoa(e.FirstLine+len(e.Code)-1)))
		n.set("style", "--k-ln: "+strconv.Itoa(digits)+"ch")
	}
	signs := map[view.Mark]string{view.MarkHighlight: "▎", view.MarkAdded: "+", view.MarkRemoved: "-", view.MarkError: "✗", view.MarkWarning: "!"}
	for i, l := range e.Code {
		num := e.FirstLine + i
		row := el("div", "class", "k-cl")
		mark, marked := e.Marks[num]
		if marked {
			row.set("class", "k-cl k-m-"+string(mark))
		}
		if digits > 0 {
			row.add(el("span", "class", "k-ln", "aria-hidden", "true").add(txt(strconv.Itoa(num))))
		}
		if len(e.Marks) > 0 {
			sign := el("span", "class", "k-sign")
			if marked {
				sign.set("aria-label", string(mark)).add(txt(signs[mark]))
			}
			row.add(sign)
		}
		src := el("span", "class", "k-src")
		for _, t := range l {
			if c := view.TokenClass(t.Kind); c != "" {
				src.add(el("span", "class", c).add(texts(t.Text)...))
				continue
			}
			src.add(texts(t.Text)...)
		}
		n.add(row.add(src))
	}
	return n
}
