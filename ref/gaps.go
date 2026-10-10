package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	btree "charm.land/bubbles/v2/tree"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// The references for the gaps (vault/knowledge/gap-analysis.md), as the
// bubbles examples show them. Each phase-1 leg writes its story with the
// same content, or changes this one to match its story.
func init() {
	register("progress", "hotty/progress", "bubbles: progress (KIT-02)", func() tea.Model {
		return screen{newProgress()}
	})
	register("spinner", "hotty/spinner", "bubbles: spinner (KIT-03)", func() tea.Model {
		return screen{newSpinners()}
	})
	register("table", "hotty/table", "bubbles: table (KIT-01)", func() tea.Model {
		return screen{newTable()}
	})
	register("list", "hotty/list", "bubbles: list and its help (KIT-04, KIT-08)", func() tea.Model {
		return screen{newList()}
	})
	register("viewport", "hotty/scroll", "bubbles: viewport (KIT-07)", func() tea.Model {
		return screen{newViewports()}
	})
	register("code", "hotty/code", "glamour: code blocks (KIT-05)", func() tea.Model {
		return screen{&codeRef{}}
	})
	register("tree", "hotty/tree", "bubbles: tree (KIT-09)", func() tea.Model {
		return screen{newTree()}
	})
	register("switch", "hotty/switch", "huh: Confirm, inline, as an on/off setting (KIT-22)", func() tea.Model {
		return screen{newSwitches()}
	})
	register("confirm", "hotty/confirm", "huh: Confirm, a yes or no question (KIT-16)", func() tea.Model {
		return screen{newConfirm()}
	})
	register("confirm-form", "hotty/confirm-form", "huh: Confirm ending a form (KIT-16)", func() tea.Model {
		return screen{newConfirmForm()}
	})
	// Bubble Tea has no charts: ntcharts' (charts.go).
	register("chart", "hotty/chart", "ntcharts: time series line chart, bar chart (KIT-10)", func() tea.Model {
		return screen{newCharts()}
	})
	register("sparkline", "hotty/sparkline", "ntcharts: sparkline (KIT-10)", func() tea.Model {
		return screen{newSparklines()}
	})
}

// progress: bars at fixed values, each under its label, and one that
// fills as time goes.
type progressRef struct {
	bars []progress.Model
	at   []float64
	live progress.Model
}

type progressTick struct{}

func newProgress() *progressRef {
	m := &progressRef{at: []float64{0, 0.25, 0.6, 1}}
	for range m.at {
		m.bars = append(m.bars, progress.New(progress.WithDefaultBlend(), progress.WithWidth(40)))
	}
	m.live = progress.New(progress.WithDefaultBlend(), progress.WithWidth(40))
	return m
}

func (m *progressRef) Init() tea.Cmd { return progressNext() }

func progressNext() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return progressTick{} })
}

func (m *progressRef) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// As wide as the window, as the story's bars are in a Column.
		for i := range m.bars {
			m.bars[i].SetWidth(msg.Width)
		}
		m.live.SetWidth(msg.Width)
	case progressTick:
		p := m.live.Percent() + 0.1
		if p > 1.0001 {
			p = 0
		}
		return tea.Batch(progressNext(), m.live.SetPercent(p))
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.live, cmd = m.live.Update(msg)
		return cmd
	}
	return nil
}

func (m *progressRef) View() string {
	var b strings.Builder
	b.WriteString("Progress\n\n")
	for i, bar := range m.bars {
		fmt.Fprintf(&b, "%s\n%s\n\n", []string{"Queued", "Download", "Install", "Done"}[i], bar.ViewAs(m.at[i]))
	}
	// bubbles has no bar without a value: this one fills as time goes.
	fmt.Fprintf(&b, "%s\n%s\n", "Indexing", m.live.View())
	return b.String()
}

// spinner: each of bubbles' frame sets, with its name.
type spinnersRef struct {
	names []string
	s     []spinner.Model
}

func newSpinners() *spinnersRef {
	m := &spinnersRef{}
	for _, sp := range []struct {
		name string
		s    spinner.Spinner
	}{
		{"Line", spinner.Line}, {"Dot", spinner.Dot}, {"MiniDot", spinner.MiniDot},
		{"Jump", spinner.Jump}, {"Pulse", spinner.Pulse}, {"Points", spinner.Points},
		{"Globe", spinner.Globe}, {"Moon", spinner.Moon}, {"Monkey", spinner.Monkey},
		{"Meter", spinner.Meter}, {"Hamburger", spinner.Hamburger}, {"Ellipsis", spinner.Ellipsis},
	} {
		m.names = append(m.names, sp.name)
		m.s = append(m.s, spinner.New(spinner.WithSpinner(sp.s),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205")))))
	}
	return m
}

func (m *spinnersRef) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, s := range m.s {
		cmds = append(cmds, s.Tick)
	}
	return tea.Batch(cmds...)
}

func (m *spinnersRef) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.s {
		var cmd tea.Cmd
		m.s[i], cmd = m.s[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m *spinnersRef) View() string {
	var b strings.Builder
	b.WriteString("Spinners\n\n")
	for i, s := range m.s {
		fmt.Fprintf(&b, "%s %s\n", s.View(), m.names[i])
	}
	return b.String()
}

// table: bubbles' table example, the world's largest cities.
type tableRef struct{ t table.Model }

func newTable() *tableRef {
	columns := []table.Column{
		{Title: "Rank", Width: 4},
		{Title: "City", Width: 10},
		{Title: "Country", Width: 10},
		{Title: "Population", Width: 10},
	}
	rows := []table.Row{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
		{"4", "Dhaka", "Bangladesh", "22,478,116"},
		{"5", "São Paulo", "Brazil", "22,429,800"},
		{"6", "Mexico City", "Mexico", "22,085,140"},
		{"7", "Cairo", "Egypt", "21,750,020"},
		{"8", "Beijing", "China", "21,333,332"},
		{"9", "Mumbai", "India", "20,961,472"},
		{"10", "Osaka", "Japan", "19,059,856"},
		{"11", "Chongqing", "China", "16,874,740"},
		{"12", "Karachi", "Pakistan", "16,839,950"},
	}
	// bubbles v2 draws no rows until the table has a width: its columns
	// and a column of padding each side of each.
	t := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true), table.WithHeight(7), table.WithWidth(4+10+10+10+4*2))
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)
	return &tableRef{t}
}

var tableBox = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))

func (m *tableRef) Init() tea.Cmd { return nil }

func (m *tableRef) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.t, cmd = m.t.Update(msg)
	return cmd
}

func (m *tableRef) View() string { return tableBox.Render(m.t.View()) + "\n" + m.t.HelpView() }

// list: bubbles' default list, items with descriptions, a filter.
type item struct{ title, desc string }

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type listRef struct{ l list.Model }

func newList() *listRef {
	items := []list.Item{
		item{"Raspberry Pi’s", "I have ’em all over my house"},
		item{"Nutella", "It's good on toast"},
		item{"Bitter melon", "It cools you down"},
		item{"Nice socks", "And by that I mean socks without holes"},
		item{"Eight hours of sleep", "I had this once"},
		item{"Cats", "Usually"},
		item{"Plantasia, the album", "My plants love it too"},
		item{"Pour over coffee", "It takes forever to make though"},
		item{"VR", "Virtual reality...what is there to say?"},
		item{"Noguchi Lamps", "Such pleasing organic forms"},
		item{"Linux", "Pretty much the best OS"},
		item{"Business school", "Just kidding"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "My Fave Things"
	return &listRef{l}
}

func (m *listRef) Init() tea.Cmd { return nil }

func (m *listRef) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.l.SetSize(ws.Width, ws.Height)
	}
	var cmd tea.Cmd
	m.l, cmd = m.l.Update(msg)
	return cmd
}

func (m *listRef) View() string { return m.l.View() }

// tree: bubbles' tree with the story's files, rendition and cells open,
// as the story starts. bubbles' tree has one root, so the repository is
// it; the story's roots are its children.
type treeRef struct{ t btree.Model }

func newTree() *treeRef {
	dir := func(name string, kids ...any) *btree.Node { return btree.Root(name).Child(kids...).Close() }
	cells := dir("cells", "cells.go", "keys.go", "tree.go").Open()
	root := btree.Root("hotty-a2ui").Child(
		dir("a2ui", "data.go", "processor.go", "surface.go"),
		dir("catalog", dir("basic", "catalog.json"), dir("hotty", "catalog.go", "catalog.json")),
		dir("docs", "profile.md", "storybook.md"),
		dir("rendition", cells, dir("html", "html.go", "kit.css", "tree.go"), dir("text", "text.go")).Open(),
		dir("view", "act.go", "tree.go", "view.go"),
		"go.mod", "Makefile", "README.md")
	return &treeRef{btree.New(root, 60, 16)}
}

func (m *treeRef) Init() tea.Cmd { return nil }

func (m *treeRef) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.t.SetSize(ws.Width, ws.Height)
	}
	var cmd tea.Cmd
	m.t, cmd = m.t.Update(msg)
	return cmd
}

func (m *treeRef) View() string { return m.t.View() }

// switch: Bubble Tea has no on/off switch. Its nearest is huh's Confirm, a
// boolean field toggled with ←/→ (h/l), y and n: here inline, a row each,
// "On" and "Off" its buttons, with the story's settings and values, and its
// account form, whose two-factor Confirm must be on. huh has no disabled
// field, so the story's disabled switches are plain ones here.
func newSwitches() *form {
	v := struct {
		airplane, wifi, bluetooth, hotspot, location, dark bool
		email                                              string
		twoFactor, digest                                  bool
	}{wifi: true, location: true, dark: true, email: "ada@example.com", digest: true}
	setting := func(title string, value *bool) *huh.Confirm {
		return huh.NewConfirm().Title(title + " ").Affirmative("On").Negative("Off").Inline(true).Value(value)
	}
	return &form{huh.NewForm(huh.NewGroup(
		huh.NewNote().Title("Settings"),
		setting("Airplane mode", &v.airplane),
		setting("Wi-Fi", &v.wifi),
		setting("Bluetooth", &v.bluetooth),
		setting("Personal hotspot", &v.hotspot),
		setting("Location services", &v.location),
		setting("Dark appearance", &v.dark),
		huh.NewNote().Title("Account"),
		huh.NewInput().Title("Email").Value(&v.email),
		setting("Two-factor sign-in", &v.twoFactor).Validate(func(on bool) error {
			if !on {
				return errors.New("Your organisation requires it")
			}
			return nil
		}),
		setting("Weekly digest by email", &v.digest),
	)).WithShowHelp(true)}
}

// confirm: huh's Confirm on its own, as huh.NewConfirm().Run() asks a
// question: the story's, with No picked at the start, as a false value
// shows it. y and n answer, ←/→ (h/l) switch, Enter takes the one picked.
func newConfirm() *form {
	var del bool
	return &form{huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Delete 3 files?").
			Description("notes.md, todo.md and draft.md go to the trash.").
			Affirmative("Yes").Negative("No").Value(&del),
	)).WithShowHelp(true)}
}

// confirm-form: a form that ends in huh's Confirm, whose answer submits
// it, with the story's fields: the repository's name and description,
// then the question, Yes picked at the start.
func newConfirmForm() *form {
	v := struct {
		name, desc string
		create     bool
	}{name: "hotty-kit", create: true}
	return &form{huh.NewForm(huh.NewGroup(
		huh.NewNote().Title("New repository"),
		huh.NewInput().Title("Name").Value(&v.name),
		huh.NewInput().Title("Description").Value(&v.desc),
		huh.NewConfirm().Title("Create the repository?").Affirmative("Yes").Negative("No").Value(&v.create),
	)).WithShowHelp(true)}
}

// viewport: two of bubbles' viewports, the story's build log at its end
// and its document (the Text's Markdown, as is: bubbles has no renderer
// for it), soft-wrapped, and a help line. Tab moves the keys between them.
// The content is read from the story, so ref runs from the repository's
// root, as scripts/ref-shot.sh runs it.
type viewportsRef struct {
	vp    [2]viewport.Model
	focus int
	help  help.Model
}

func newViewports() *viewportsRef {
	log, doc := scrollStory()
	m := &viewportsRef{help: help.New()}
	m.vp[0] = viewport.New(viewport.WithHeight(8))
	m.vp[0].SetContentLines(log)
	m.vp[1] = viewport.New(viewport.WithHeight(10))
	m.vp[1].SoftWrap = true
	m.vp[1].SetContent(doc)
	return m
}

// scrollStory is the story hotty/scroll's log and its document's text.
func scrollStory() (log []string, doc string) {
	b, err := os.ReadFile("story/stories/hotty/scroll.json")
	if err != nil {
		return []string{"ref viewport runs from the repository's root: " + err.Error()}, ""
	}
	var s struct {
		Messages []struct {
			CreateSurface *struct {
				DataModel struct{ Log []string }
			}
			UpdateComponents *struct {
				Components []struct{ ID, Text string }
			}
		}
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return []string{err.Error()}, ""
	}
	for _, msg := range s.Messages {
		if msg.CreateSurface != nil {
			log = msg.CreateSurface.DataModel.Log
		}
		if msg.UpdateComponents != nil {
			for _, c := range msg.UpdateComponents.Components {
				if c.ID == "doc_text" {
					doc = c.Text
				}
			}
		}
	}
	return log, doc
}

func (m *viewportsRef) Init() tea.Cmd { return nil }

func (m *viewportsRef) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for i := range m.vp {
			m.vp[i].SetWidth(msg.Width)
		}
		m.vp[0].GotoBottom()
		m.help.SetWidth(msg.Width)
		return nil
	case tea.KeyPressMsg:
		if msg.String() == "tab" {
			m.focus = 1 - m.focus
			return nil
		}
	}
	var cmd tea.Cmd
	m.vp[m.focus], cmd = m.vp[m.focus].Update(msg)
	return cmd
}

func (m *viewportsRef) View() string {
	km := m.vp[m.focus].KeyMap
	bold := lipgloss.NewStyle().Bold(true)
	return bold.Render("Build log") + "\n" + m.vp[0].View() + "\n\n" +
		bold.Render("A document") + "\n" + m.vp[1].View() + "\n\n" +
		m.help.ShortHelpView([]key.Binding{km.Up, km.Down, km.PageDown, km.PageUp})
}

// code: the story's code as glamour renders Markdown's fenced code
// blocks, which is how code looks in a Charm app (bubbles has no code
// component): chroma's colours, a block's margin, no line numbers and no
// marks. The story is read as viewport's is, from the repository's root.
type codeRef struct{ out string }

func (m *codeRef) Init() tea.Cmd { return nil }

func (m *codeRef) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		r, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(ws.Width))
		if err != nil {
			m.out = err.Error()
			return nil
		}
		if m.out, err = r.Render(codeMarkdown()); err != nil {
			m.out = err.Error()
		}
	}
	return nil
}

func (m *codeRef) View() string { return m.out }

// codeMarkdown is the story hotty/code as Markdown: each HottyCode a
// fenced block under its title, then the Text.
func codeMarkdown() string {
	b, err := os.ReadFile("story/stories/hotty/code.json")
	if err != nil {
		return "ref code runs from the repository's root: " + err.Error()
	}
	var s struct {
		Messages []struct {
			UpdateComponents *struct {
				Components []struct{ ID, Text, Code, Language string }
			}
		}
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err.Error()
	}
	byID := map[string]struct{ text, code, lang string }{}
	for _, msg := range s.Messages {
		if msg.UpdateComponents != nil {
			for _, c := range msg.UpdateComponents.Components {
				byID[c.ID] = struct{ text, code, lang string }{c.Text, c.Code, c.Language}
			}
		}
	}
	// glamour knows a language by its name, not a file's.
	lang := map[string]string{"fetch.py": "python"}
	var out strings.Builder
	for _, id := range []string{"go_title", "go", "py_title", "py", "prose"} {
		c := byID[id]
		if c.code != "" {
			l := c.lang
			if v, ok := lang[l]; ok {
				l = v
			}
			fmt.Fprintf(&out, "```%s\n%s```\n\n", l, c.code)
			continue
		}
		out.WriteString(c.text + "\n\n")
	}
	return out.String()
}
