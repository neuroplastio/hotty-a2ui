// Package hotty is HOTTY's catalog for A2UI (NEIO-11): what a terminal
// adds to the basic catalog. HottyShortcut binds a key, HottyForm submits
// its fields, hottyFocus and hottyBlur move the keyboard, and
// hottyScrollTo scrolls a HottyScrollView, for the renderer and the agent
// alike. Its names start with Hotty (components) and
// hotty (functions), so that none can be one A2UI adds later.
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
// hottyFocus, hottyBlur and hottyScrollTo are the renderer's, which gives
// them (Implement).
func Catalog() *a2ui.Catalog {
	c, err := a2ui.ParseCatalog(doc)
	if err != nil {
		panic(err)
	}
	return c
}

// Renderer is what the catalog's functions do, the renderer's: Focus
// moves the keyboard to a component, Blur takes it back, ScrollTo scrolls
// a HottyScrollView ("start" or "end"). s is the surface of the call, nil
// when the agent called it (callRendererFunction names no surface); scope
// is the caller's, which picks a template's instance.
type Renderer struct {
	Focus    func(s *a2ui.Surface, scope a2ui.Scope, id string) error
	Blur     func(s *a2ui.Surface) error
	ScrollTo func(s *a2ui.Surface, scope a2ui.Scope, id, to string) error
}

// Implement gives a catalog's hottyFocus, hottyBlur and hottyScrollTo their
// implementation: the renderer's.
func Implement(c *a2ui.Catalog, r Renderer) {
	if f := c.Functions["hottyFocus"]; f != nil {
		f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) {
			id, _ := args["id"].(string)
			return nil, r.Focus(ctx.Surface, ctx.Scope, id)
		}
	}
	if f := c.Functions["hottyBlur"]; f != nil {
		f.Impl = func(ctx *a2ui.Context, _ map[string]any) (any, error) { return nil, r.Blur(ctx.Surface) }
	}
	if f := c.Functions["hottyScrollTo"]; f != nil {
		f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) {
			id, _ := args["id"].(string)
			to, _ := args["to"].(string)
			return nil, r.ScrollTo(ctx.Surface, ctx.Scope, id, to)
		}
	}
}
