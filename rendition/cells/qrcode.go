package cells

import (
	"github.com/neuroplastio/hotty-a2ui/qr"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyQRCode in cells (profile §6.29): its modules a column wide and
// half a row tall, two to a cell in half blocks, so that they are about
// square, in ink on paper (Cell.Paper) with its quiet zone, as OpenTUI's
// QR code draws it; then its label, centred under it. A box narrower than
// the code has its value as text instead, since a code cut does not scan.

// qrSide is the columns a code takes, its quiet zone's included; 0
// without one.
func qrSide(e *view.Element) int {
	if e.QR == nil {
		return 0
	}
	return e.QR.Size + 2*qr.Quiet
}

// qrWidth is a HottyQRCode's natural width: its code's or its label's,
// whichever is wider; its error's where its value has no code.
func qrWidth(e *view.Element) int {
	if e.Error != "" {
		return Width("✗ " + e.Error)
	}
	if e.QR == nil {
		return 0
	}
	return max(qrSide(e), Width(e.Label))
}

// qrMinimum is the narrowest a HottyQRCode takes: its code's width, which
// does not shrink.
func qrMinimum(e *view.Element) int {
	if e.Error != "" {
		return longestWord(line("✗ "+e.Error, style{}))
	}
	return qrSide(e)
}

// qrFits reports whether a HottyQRCode's code fits in w columns.
func qrFits(e *view.Element, w int) bool { return e.QR != nil && qrSide(e) <= w }

// qrLines are what a HottyQRCode draws under its code, or, where the code
// does not fit, instead of it: its label, then without the code its value,
// each wrapped.
func qrLines(e *view.Element, w int) [][]glyph {
	var out [][]glyph
	if e.QR != nil && e.Label != "" {
		out = wrap(line(e.Label, style{}), w, nil, nil)
	}
	if e.QR != nil && !qrFits(e, w) {
		v, _ := e.Value.(string)
		out = append(out, wrap(line(v, style{}), w, nil, nil)...)
	}
	return out
}

// qrHeight is a HottyQRCode's rows at width w: its code's, half its
// modules' rows rounded up, and its lines'; none without a value.
func qrHeight(e *view.Element, w int) int {
	h := len(qrLines(e, w))
	if qrFits(e, w) {
		h += (qrSide(e) + 1) / 2
	}
	return h
}

// paintQRCode paints a HottyQRCode in its box from (x, y), w wide: its
// code, quiet zone and all, at the box's start, and its lines under it,
// each centred on the code's width where it is narrower.
func (l *layout) paintQRCode(cv *canvas, e *view.Element, x, y, w int) {
	if e.Error != "" {
		for i, ln := range errorLines(e, w) {
			cv.write(x, y+i, w, ln)
		}
		return
	}
	row := y
	side := 0
	if qrFits(e, w) {
		side = qrSide(e)
		ink := style{paper: true}
		for top := -qr.Quiet; top < e.QR.Size+qr.Quiet; top += 2 {
			for i := range side {
				mx := i - qr.Quiet
				g := halfBlock(e.QR.Dark(mx, top), e.QR.Dark(mx, top+1))
				cv.set(x+i, row, glyph{text: g, width: 1, style: ink})
			}
			row++
		}
	}
	for _, ln := range qrLines(e, w) {
		dx := 0
		if lw := width(ln); lw < side {
			dx = (side - lw) / 2
		}
		cv.write(x+dx, row, w-dx, ln)
		row++
	}
}
