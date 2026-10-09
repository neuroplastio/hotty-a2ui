package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
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
	register("list", "hotty/list", "bubbles: list (KIT-04)", func() tea.Model {
		return screen{newList()}
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
