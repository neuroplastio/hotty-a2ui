package main

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// The reference for HottyPicker (KIT-PICKER): bubbles has no palette, and
// its list filtering as you type is the nearest, as soft-serve's and
// gum filter's pickers use it. The story's rows, "Group: Title" each with
// its detail; / filters.
func init() {
	register("picker", "hotty/picker", "bubbles: list, filtering, as a palette (KIT-PICKER)", func() tea.Model {
		return screen{newPicker()}
	})
}

type pickerRef struct{ l list.Model }

func newPicker() *pickerRef {
	items := []list.Item{
		item{"Pane: Split right", "Recently used"},
		item{"Tab: New tab", "Recently used"},
		item{"Pane: fish", "~/code/plexos"},
		item{"Pane: claude", "agent · plexos"},
		item{"Pane: htop", "bastion"},
		item{"Pane: Close", ""},
		item{"Pane: Zoom", ""},
		item{"Workspace: New workspace", ""},
		item{"Workspace: Rename…", ""},
		item{"Config: Get a setting", ""},
		item{"Session: Quit", ""},
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowTitle(false)
	return &pickerRef{l}
}

func (m *pickerRef) Init() tea.Cmd { return nil }

func (m *pickerRef) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.l.SetSize(ws.Width, ws.Height)
	}
	var cmd tea.Cmd
	m.l, cmd = m.l.Update(msg)
	return cmd
}

func (m *pickerRef) View() string { return m.l.View() }
