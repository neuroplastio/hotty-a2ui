package main

import (
	"encoding/json"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	register("suggest", "hotty/suggest", "bubbles: textinput's suggestions (KIT-11)", func() tea.Model {
		return screen{newSuggest()}
	})
}

// suggestRef: the story's three fields as bubbles' text inputs with
// suggestions (ShowSuggestions, SetSuggestions), each under its title:
// Command with "git c" typed and the keyboard, Country empty, City with
// what none of its suggestions matches. Tab takes the suggestion, ↑ and ↓
// (Control+p, Control+n) move through them, as bubbles binds them; bubbles
// draws only the ghost text, and no list. Tab with nothing to take moves
// on, as the kit's does. The content is read from the story, so ref runs
// from the repository's root, as scripts/ref-shot.sh runs it.
type suggestRef struct {
	titles []string
	in     []textinput.Model
	focus  int
	help   help.Model
}

func newSuggest() *suggestRef {
	var st struct {
		Messages []struct {
			CreateSurface struct {
				DataModel struct {
					Command, Country, City      string
					Commands, Countries, Cities []string
				}
			}
		}
	}
	if b, err := os.ReadFile("story/stories/hotty/suggest.json"); err == nil {
		_ = json.Unmarshal(b, &st)
	}
	d := st.Messages[0].CreateSurface.DataModel
	m := &suggestRef{titles: []string{"Command", "Country", "City"}, help: help.New()}
	for i, f := range []struct {
		value, placeholder string
		options            []string
	}{
		{d.Command, "git …", d.Commands},
		{d.Country, "Start typing a country", d.Countries},
		{d.City, "", d.Cities},
	} {
		in := textinput.New()
		in.Prompt = "> "
		in.Placeholder = f.placeholder
		// Without a width, bubbles shows a placeholder's first character.
		in.SetWidth(40)
		in.ShowSuggestions = true
		in.SetValue(f.value)
		in.SetSuggestions(f.options)
		if i == 0 {
			in.Focus()
		}
		m.in = append(m.in, in)
	}
	return m
}

func (m *suggestRef) Init() tea.Cmd { return nil }

func (m *suggestRef) Update(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		in := &m.in[m.focus]
		take := len(in.MatchedSuggestions()) > 0 && in.CurrentSuggestion() != in.Value()
		switch k.String() {
		case "tab":
			if !take {
				m.move(1)
				return nil
			}
		case "shift+tab":
			m.move(-1)
			return nil
		}
	}
	var cmd tea.Cmd
	m.in[m.focus], cmd = m.in[m.focus].Update(msg)
	return cmd
}

// move gives the keyboard to the field by, round from either end.
func (m *suggestRef) move(by int) {
	m.in[m.focus].Blur()
	m.focus = (m.focus + by + len(m.in)) % len(m.in)
	m.in[m.focus].Focus()
}

func (m *suggestRef) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Run a command") + "\n\n")
	for i, in := range m.in {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.titles[i]) + "\n" + in.View() + "\n\n")
	}
	b.WriteString(m.help.ShortHelpView([]key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "suggestions")),
	}))
	return b.String()
}
