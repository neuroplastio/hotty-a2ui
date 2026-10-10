package html

import (
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/qr"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// qrCode is a HottyQRCode (profile §2): its code an inline svg, a white
// square its quiet zone and all, and one black path of its dark modules,
// a rectangle a run of them across a row, whose id lets a new value be
// one attribute's delta; black on white in every theme, edges crisp. A
// screen reader has it as an image named by its value. Its label is
// centred under it (kit.css); where its value has no code, its error.
func qrCode(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-qr")
	v, _ := e.Value.(string)
	if e.Error != "" {
		return n.add(el("div", "id", partID(e.ID, partError), "class", "k-error").add(txt("✗ " + e.Error)))
	}
	if e.QR == nil {
		return n
	}
	side := strconv.Itoa(e.QR.Size + 2*qr.Quiet)
	n.add(el("svg", "class", "k-qr-code", "viewBox", "0 0 "+side+" "+side, "shape-rendering", "crispEdges",
		"role", "img", "aria-label", "QR code: "+v, "style", "--k-modules: "+side).add(
		el("rect", "width", side, "height", side, "fill", "#fff"),
		el("path", "id", partID(e.ID, partPath), "d", qrPath(e.QR), "fill", "#000")))
	if e.Label != "" {
		n.add(el("div", "id", partID(e.ID, partLabel), "class", "k-qr-label").add(texts(e.Label)...))
	}
	return n
}

// qrPath is a code's dark modules as path data, in modules from the quiet
// zone's corner: a rectangle a run of them across a row.
func qrPath(c *qr.Code) string {
	var b strings.Builder
	for y := range c.Size {
		for x := 0; x < c.Size; {
			if !c.Dark(x, y) {
				x++
				continue
			}
			run := 1
			for c.Dark(x+run, y) {
				run++
			}
			n := strconv.Itoa(run)
			b.WriteString("M" + strconv.Itoa(x+qr.Quiet) + " " + strconv.Itoa(y+qr.Quiet) + "h" + n + "v1h-" + n + "z")
			x += run
		}
	}
	return b.String()
}
