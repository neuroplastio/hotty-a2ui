package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/barchart"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/NimbleMarkets/ntcharts/v2/sparkline"
)

// The references for KIT-10's charts (gaps.go registers them). Bubble Tea
// has none; ntcharts (NimbleMarkets/ntcharts, the chart library for
// Bubble Tea) is what a Go TUI draws them with: its time series line
// chart in braille, its bar chart and its sparkline, with the stories'
// data and their colours (the kit's series are info, then warning: ANSI
// cyan and yellow). v2.2.0, the newest that needs no newer modules than
// the other references use.

var (
	refInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	refWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	refMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	refBold    = lipgloss.NewStyle().Bold(true)
)

// chartsRef: hotty/chart's charts as ntcharts draws them. The line chart
// is a timeserieslinechart, two data sets in braille, its axis 0 to 75 as
// the kit fits it; ntcharts has no legend, so one is written under it.
// ntcharts' bars stack a bar's values, so a day's api and web deploys are
// one bar of two segments where the kit stands them side by side; its bars
// rise from 0 only, so the balance's months below 0 have none.
type chartsRef struct{ view string }

func newCharts() *chartsRef {
	cpu := []float64{12, 18, 15, 22, 30, 28, 35, 31, 40, 38, 45, 42, 50, 47, 44, 52, 58, 55, 61, 57}
	mem := []float64{40, 41, 41, 43, 44, 44, 46, 45, 47, 48, 48, 49, 50, 50, 51, 52, 52, 53, 54, 54}
	t0 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	minutes := func(_ int, v float64) string {
		return time.UnixMilli(int64(math.Round(v * 1e3))).UTC().Format("15:04")
	}
	load := timeserieslinechart.New(56, 10,
		timeserieslinechart.WithTimeRange(t0, t0.Add(19*time.Minute)),
		timeserieslinechart.WithYRange(0, 75),
		timeserieslinechart.WithXYSteps(3, 3),
		timeserieslinechart.WithXLabelFormatter(minutes),
		timeserieslinechart.WithDataSetStyle("cpu", refInfo),
		timeserieslinechart.WithDataSetStyle("mem", refWarning))
	for i := range cpu {
		at := t0.Add(time.Duration(i) * time.Minute)
		load.PushDataSet("cpu", timeserieslinechart.TimePoint{Time: at, Value: cpu[i]})
		load.PushDataSet("mem", timeserieslinechart.TimePoint{Time: at, Value: mem[i]})
	}
	load.DrawBrailleAll()

	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	api := []float64{12, 19, 7, 23, 16, 4, 9}
	web := []float64{8, 11, 9, 14, 20, 2, 5}
	deploys := barchart.New(56, 10)
	for i, d := range days {
		deploys.Push(barchart.BarData{Label: d, Values: []barchart.BarValue{
			{Name: "api", Value: api[i], Style: refInfo},
			{Name: "web", Value: web[i], Style: refWarning},
		}})
	}
	deploys.Draw()

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	balance := []float64{420, 310, -180, -60, 250, 530}
	bal := barchart.New(56, 8)
	for i, m := range months {
		bal.Push(barchart.BarData{Label: m, Values: []barchart.BarValue{{Name: m, Value: balance[i], Style: refInfo}}})
	}
	bal.Draw()

	key := func(glyph string) string {
		return refInfo.Render(glyph) + refMuted.Render(" cpu") + "   " + refWarning.Render(glyph) + refMuted.Render(" mem")
	}
	var b strings.Builder
	for _, line := range []string{
		refBold.Render("Chart"),
		refBold.Render("Load") + ", the last 20 minutes (%)",
		load.View(),
		key("━"),
		refBold.Render("Deploys") + " a day",
		deploys.View(),
		strings.NewReplacer("cpu", "api", "mem", "web").Replace(key("■")),
		refBold.Render("Balance") + " by month ($)",
		bal.View(),
	} {
		b.WriteString(line + "\n")
	}
	return &chartsRef{view: b.String()}
}

func (m *chartsRef) Init() tea.Cmd              { return nil }
func (m *chartsRef) Update(msg tea.Msg) tea.Cmd { return nil }
func (m *chartsRef) View() string               { return m.view }

// sparklinesRef: hotty/sparkline's sparklines as ntcharts draws them: CPU
// by host, 30 columns each from 0 to 100 between the host's name and its
// latest value; api three rows tall; the temperatures, which ntcharts
// draws from 0, so those below it have none; the pings, ntcharts having
// no gap, its lost ones 0.
type sparklinesRef struct{ view string }

func newSparklines() *sparklinesRef {
	hosts := []struct {
		name string
		vals []float64
	}{
		{"api", []float64{21, 25, 24, 30, 28, 35, 41, 38, 36, 44, 47, 43, 39, 42, 48, 52, 49, 55, 51, 58}},
		{"web", []float64{64, 61, 66, 70, 68, 59, 55, 57, 62, 60, 58, 54, 50, 53, 49, 47, 51, 46, 44, 42}},
		{"db", []float64{8, 9, 8, 12, 30, 61, 74, 52, 33, 20, 14, 11, 10, 9, 9, 8, 10, 9, 8, 9}},
	}
	spark := func(w, h int, vals []float64, opts ...sparkline.Option) string {
		s := sparkline.New(w, h, append([]sparkline.Option{sparkline.WithStyle(refInfo)}, opts...)...)
		s.PushAll(vals)
		s.Draw()
		return s.View()
	}
	var b strings.Builder
	b.WriteString(refBold.Render("Sparkline") + "\n")
	b.WriteString(refBold.Render("CPU") + " by host, a value a minute (%)\n")
	for _, h := range hosts {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(4).Render(h.name),
			spark(30, 1, h.vals, sparkline.WithMaxValue(100), sparkline.WithNoAutoMaxValue()),
			lipgloss.NewStyle().Width(3).Align(lipgloss.Right).Render(strconv.FormatFloat(h.vals[len(h.vals)-1], 'f', -1, 64))) + "\n")
	}
	b.WriteString(refBold.Render("api") + ", three rows tall\n")
	b.WriteString(spark(len(hosts[0].vals), 3, hosts[0].vals) + "\n")
	temps := []float64{3.1, 2.4, 0.8, -1.5, -3.2, -2.1, 0.4, 2.9, 5.6, 7.2, 8.1, 6.4}
	b.WriteString(refBold.Render("Temperature") + " by month (°C): from the lowest, below 0\n")
	b.WriteString(spark(len(temps), 2, temps) + "\n")
	pings := []float64{5, 7, 6, 9, 0, 0, 8, 10, 12, 11, 0, 14, 13, 15}
	b.WriteString(refBold.Render("Pings") + ", three lost\n")
	b.WriteString(spark(len(pings), 1, pings) + "\n")
	return &sparklinesRef{view: b.String()}
}

func (m *sparklinesRef) Init() tea.Cmd              { return nil }
func (m *sparklinesRef) Update(msg tea.Msg) tea.Cmd { return nil }
func (m *sparklinesRef) View() string               { return m.view }
