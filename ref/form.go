package main

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func init() {
	register("form", "hotty/form", "huh: Input, Select and Confirm in a form", func() tea.Model {
		var (
			name, email, plan string
			agree             bool
		)
		f := huh.NewForm(huh.NewGroup(
			huh.NewNote().Title("Sign up"),
			huh.NewInput().Title("Name").Value(&name),
			huh.NewInput().Title("Email").Value(&email),
			huh.NewSelect[string]().Title("Plan").Options(
				huh.NewOption("Free", "free"),
				huh.NewOption("Pro", "pro"),
			).Value(&plan),
			huh.NewConfirm().Title("I accept the terms").Value(&agree),
		)).WithShowHelp(true)
		return screen{&form{f}}
	})

	register("textarea", "hotty/keys", "bubbles: textarea", func() tea.Model {
		ta := textarea.New()
		ta.SetValue("A short line\nA much longer line, to come back along\nshort\nAnd another long line to finish on")
		ta.SetWidth(60)
		ta.SetHeight(6)
		ta.Focus()
		return screen{&textArea{ta}}
	})
}

type form struct{ f *huh.Form }

func (m *form) Init() tea.Cmd { return m.f.Init() }

func (m *form) Update(msg tea.Msg) tea.Cmd {
	f, cmd := m.f.Update(msg)
	if f, ok := f.(*huh.Form); ok {
		m.f = f
	}
	return cmd
}

func (m *form) View() string { return m.f.View() }

type textArea struct{ ta textarea.Model }

func (m *textArea) Init() tea.Cmd { return textarea.Blink }

func (m *textArea) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return cmd
}

func (m *textArea) View() string { return "Notes\n" + m.ta.View() }
