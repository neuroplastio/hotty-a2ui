package view

import (
	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/qr"
)

// QRTooLong is a HottyQRCode's Error when its value is too long for any
// QR code at its level.
const QRTooLong = "Too long for a QR code"

// mapQRCode makes a HottyQRCode's element (profile §6.29): Value, its
// value as it is; QR, the value encoded at its level (Variant, M without
// one), or Error where no code holds it; Label, its caption.
func mapQRCode(b *Builder, n *a2ui.Node) *Element {
	level, ok := qr.ParseLevel(b.Enum(n, "errorCorrection", "M"))
	if !ok {
		level = qr.M
	}
	v := b.String(n, "value")
	e := &Element{Kind: QRCode, Value: v, Label: b.String(n, "label"), Variant: level.String()}
	if v == "" {
		return e
	}
	c, err := qr.Encode(v, level)
	if err != nil {
		e.Error = QRTooLong
		return e
	}
	e.QR = c
	return e
}
