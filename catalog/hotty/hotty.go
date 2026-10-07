// Package hotty is HOTTY's catalog for A2UI (NEIO-11): what a terminal
// adds to the basic catalog. Shortcut binds a key, Form submits its
// fields, and focus and blur move the keyboard, for the renderer and the
// agent alike.
package hotty

import (
	_ "embed"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// ID is the catalog's id.
const ID = "https://neuroplast.io/hotty/a2ui/v1/catalog.json"

//go:embed catalog.json
var doc []byte

// Doc is the catalog document, as an agent is given it.
func Doc() []byte { return doc }

// Catalog reads the catalog. Its functions have no implementation here:
// focus and blur are the renderer's, which gives them (Implement).
func Catalog() *a2ui.Catalog {
	c, err := a2ui.ParseCatalog(doc)
	if err != nil {
		panic(err)
	}
	return c
}
