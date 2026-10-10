package view

import "github.com/neuroplastio/hotty-a2ui/a2ui"

// The sizes of a HottyBigText (its Element's Variant): in cells, letters
// 3, 4 and 5 rows tall (profile §6.28).
const (
	BigSmall  = "small"
	BigMedium = "medium"
	BigLarge  = "large"
)

// mapBigText makes a HottyBigText's element (profile §6.28): Label, its
// text as it is; Variant, its size (medium without one); Align, where its
// lines go across its box (start without one).
func mapBigText(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: BigText, Label: b.String(n, "text"), Variant: BigMedium, Align: "start"}
	switch s := b.Enum(n, "size", BigMedium); s {
	case BigSmall, BigLarge:
		e.Variant = s
	}
	switch a := b.Enum(n, "align", "start"); a {
	case "center", "end":
		e.Align = a
	}
	return e
}
