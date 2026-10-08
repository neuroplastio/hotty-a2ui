// Package theme is the colours a rendition paints with: one set of roles
// (NEIO-4's names) for the cells, and the same colours as the kit's CSS
// variables on a HOTTY host. A theme's empty colour keeps the terminal's
// (or the host's) own.
package theme

import "strings"

// Theme is a named set of colours, each "#rrggbb" or "" for the default.
type Theme struct {
	Name string
	// Bg paints the whole surface; Fg is the text.
	Bg, Fg string
	// The roles, as cells names them.
	Muted, Accent, Selection, Surface, Border string
	Success, Warning, Error, Info             string
	// Shape is how round things are on a host; cells draws no shapes.
	Shape Shape
}

// Shape is corner radii, as CSS lengths; "" keeps the kit's own (rounded,
// rendition/html/kit.css).
type Shape struct {
	Button, Chip, Card, Field string
}

// Default keeps the terminal's colours and the host's palette: the ANSI-16
// floor in cells, and the host's stylesheet on a host.
var Default = Theme{Name: "Terminal"}

// All are the themes the storybook offers, in the order F3 cycles them.
var All = []Theme{
	Default,
	// Material 3's baseline scheme: the surface, on-surface and primary
	// roles, outline and error as the scheme has them. Success, warning and
	// info are not in the scheme; they are chosen to sit with it.
	{
		Name: "Material dark",
		Bg:   "#141218", Fg: "#E6E0E9", Muted: "#CAC4D0", Accent: "#D0BCFF",
		Selection: "#4A4458", Surface: "#211F26", Border: "#938F99",
		Success: "#A8DBA0", Warning: "#F5C77E", Error: "#F2B8B5", Info: "#A8C8FF",
		Shape: material,
	},
	{
		Name: "Material light",
		Bg:   "#FFFBFE", Fg: "#1C1B1F", Muted: "#49454F", Accent: "#6750A4",
		Selection: "#E8DEF8", Surface: "#F3EDF7", Border: "#79747E",
		Success: "#386A20", Warning: "#8A5A00", Error: "#B3261E", Info: "#205A9E",
		Shape: material,
	},
	{
		Name: "Nord",
		Bg:   "#2e3440", Fg: "#d8dee9", Muted: "#7b88a1", Accent: "#88c0d0",
		Selection: "#3b4252", Surface: "#3b4252", Border: "#4c566a",
		Success: "#a3be8c", Warning: "#ebcb8b", Error: "#bf616a", Info: "#81a1c1",
	},
	{
		Name: "Dracula",
		Bg:   "#282a36", Fg: "#f8f8f2", Muted: "#6272a4", Accent: "#bd93f9",
		Selection: "#44475a", Surface: "#343746", Border: "#6272a4",
		Success: "#50fa7b", Warning: "#f1fa8c", Error: "#ff5555", Info: "#8be9fd",
	},
	{
		Name: "Gruvbox",
		Bg:   "#282828", Fg: "#ebdbb2", Muted: "#928374", Accent: "#fabd2f",
		Selection: "#3c3836", Surface: "#32302f", Border: "#665c54",
		Success: "#b8bb26", Warning: "#fe8019", Error: "#fb4934", Info: "#83a598",
	},
	{
		Name: "Solarized light",
		Bg:   "#fdf6e3", Fg: "#657b83", Muted: "#93a1a1", Accent: "#268bd2",
		Selection: "#eee8d5", Surface: "#eee8d5", Border: "#93a1a1",
		Success: "#859900", Warning: "#b58900", Error: "#dc322f", Info: "#2aa198",
	},
}

// material is Material 3's shape scale: full (pill) buttons, small (8dp)
// chips, medium (12dp) cards, extra-small (4dp) text fields.
var material = Shape{Button: "999px", Chip: "8px", Card: "12px", Field: "4px"}

// ByName is the theme with that name, ignoring case; false if none has it.
func ByName(name string) (Theme, bool) {
	for _, t := range All {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	return Theme{}, false
}

// Next is the theme after t in All, wrapping around.
func Next(t Theme) Theme {
	for i, o := range All {
		if o.Name == t.Name {
			return All[(i+1)%len(All)]
		}
	}
	return All[0]
}

// Colour is the colour of a role, by its name ("muted", "accent", …), or ""
// for the default. "fg" and "bg" are the text and the surface.
func (t Theme) Colour(role string) string {
	switch role {
	case "fg":
		return t.Fg
	case "bg":
		return t.Bg
	case "muted":
		return t.Muted
	case "accent":
		return t.Accent
	case "selection":
		return t.Selection
	case "surface":
		return t.Surface
	case "border":
		return t.Border
	case "success":
		return t.Success
	case "warning":
		return t.Warning
	case "error":
		return t.Error
	case "info":
		return t.Info
	}
	return ""
}
