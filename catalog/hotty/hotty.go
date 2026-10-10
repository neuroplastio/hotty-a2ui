// Package hotty is HOTTY's catalog for A2UI (NEIO-11): what a terminal
// adds to the basic catalog. HottyShortcut binds a key, HottyForm submits
// its fields, hottyFocus and hottyBlur move the keyboard, hottyScrollTo
// scrolls a HottyScrollView, hottyExpandAll and hottyCollapseAll fold a
// HottyTree, hottyToast and hottyDismissToast show and take away a toast,
// and hottyStartTimer, hottyStopTimer, hottyToggleTimer and hottyResetTimer
// work a HottyTimer or a HottyStopwatch, for the renderer and the agent
// alike. Its names start with Hotty (components) and hotty (functions), so
// that none can be one A2UI adds later.
package hotty

import (
	_ "embed"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// ID is the catalog's id.
const ID = "https://hotty.neuroplast.io/a2ui/v1/catalog"

//go:embed catalog.json
var doc []byte

// Doc is the catalog document, as an agent is given it.
func Doc() []byte { return doc }

// Catalog reads the catalog. Its functions have no implementation here:
// they are the renderer's, which gives them (Implement).
func Catalog() *a2ui.Catalog {
	c, err := a2ui.ParseCatalog(doc)
	if err != nil {
		panic(err)
	}
	return c
}

// Renderer is what the catalog's functions do, the renderer's: Focus
// moves the keyboard to a component, Blur takes it back, ScrollTo scrolls
// a HottyScrollView ("start" or "end"), FoldAll opens every branch of a
// HottyTree (open) or closes them all, Toast shows a toast (hottyToast's
// args, resolved when a component called it, as the agent sent them when
// it did), DismissToast takes one away, and Timer works a HottyTimer or a
// HottyStopwatch (do is "start", "stop", "toggle" or "reset"). s is the
// surface of the call, nil when the agent called it (callRendererFunction
// names no surface); scope is the caller's, which picks a template's
// instance.
type Renderer struct {
	Focus        func(s *a2ui.Surface, scope a2ui.Scope, id string) error
	Blur         func(s *a2ui.Surface) error
	ScrollTo     func(s *a2ui.Surface, scope a2ui.Scope, id, to string) error
	FoldAll      func(s *a2ui.Surface, scope a2ui.Scope, id string, open bool) error
	Toast        func(s *a2ui.Surface, args map[string]any) error
	DismissToast func(s *a2ui.Surface, id string) error
	Timer        func(s *a2ui.Surface, scope a2ui.Scope, id, do string) error
}

// Implement gives a catalog's functions their implementation: the
// renderer's.
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
	for name, open := range map[string]bool{"hottyExpandAll": true, "hottyCollapseAll": false} {
		if f := c.Functions[name]; f != nil && r.FoldAll != nil {
			f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) {
				id, _ := args["id"].(string)
				return nil, r.FoldAll(ctx.Surface, ctx.Scope, id, open)
			}
		}
	}
	if f := c.Functions["hottyToast"]; f != nil && r.Toast != nil {
		f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) { return nil, r.Toast(ctx.Surface, args) }
	}
	if f := c.Functions["hottyDismissToast"]; f != nil && r.DismissToast != nil {
		f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) {
			return nil, r.DismissToast(ctx.Surface, a2ui.ToString(args["id"]))
		}
	}
	for name, do := range map[string]string{"hottyStartTimer": "start", "hottyStopTimer": "stop", "hottyToggleTimer": "toggle", "hottyResetTimer": "reset"} {
		if f := c.Functions[name]; f != nil && r.Timer != nil {
			f.Impl = func(ctx *a2ui.Context, args map[string]any) (any, error) {
				id, _ := args["id"].(string)
				return nil, r.Timer(ctx.Surface, ctx.Scope, id, do)
			}
		}
	}
}
