package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/barchart"
)

const (
	chartWidth  = 64
	chartHeight = 20
)

// Catppuccin Mocha palette.
var (
	ctpMauve    = lipgloss.Color("#cba6f7")
	ctpTeal     = lipgloss.Color("#94e2d5")
	ctpSubtext1 = lipgloss.Color("#bac2de")
	ctpOverlay1 = lipgloss.Color("#7f849c")
	ctpOverlay0 = lipgloss.Color("#6c7086")
	ctpSurface2 = lipgloss.Color("#585b70")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ctpMauve)

	subtleStyle = lipgloss.NewStyle().
			Foreground(ctpOverlay1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ctpSurface2).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(ctpSubtext1)

	barStyle = lipgloss.NewStyle().
			Foreground(ctpTeal)

	helpStyle = lipgloss.NewStyle().
			Foreground(ctpOverlay1).
			MarginTop(1)

	emptyStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(ctpOverlay0)
)

type screen int

const (
	screenInput screen = iota
	screenGraph
)

type model struct {
	screen   screen
	textarea textarea.Model
	chart    barchart.Model
	isEmpty  bool
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type or paste text here, then press ctrl+d to analyze."
	ta.Prompt = ""
	ta.ShowLineNumbers = true
	ta.SetWidth(60)
	ta.SetHeight(12)
	ta.Focus()

	return model{
		screen:   screenInput,
		textarea: ta,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenInput:
		return m.updateInput(msg)
	case screenGraph:
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			return m.updateGraph(keyMsg)
		}
	}
	return m, nil
}

func (m model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+d":
			freq := charFrequencies(m.textarea.Value())
			m.isEmpty = len(freq) == 0
			if !m.isEmpty {
				m.chart = drawChart(freq)
			}
			m.screen = screenGraph
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) updateGraph(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	case "r":
		return initialModel(), nil
	}
	return m, nil
}

func (m model) View() tea.View {
	content := m.viewInput()
	if m.screen == screenGraph {
		content = m.viewGraph()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) viewInput() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Freaks"))
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("Type or paste text, then press ctrl+d to analyze (ctrl+c to quit)."))
	b.WriteString("\n\n")
	b.WriteString(boxStyle.Render(m.textarea.View()))
	b.WriteString("\n")
	return b.String()
}

func (m model) viewGraph() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Character Frequency"))
	b.WriteString("\n\n")
	if m.isEmpty {
		b.WriteString(emptyStyle.Render("input is empty!"))
	} else {
		b.WriteString(m.chart.View())
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q to quit | r to analyze new input"))
	return b.String()
}

func charFrequencies(s string) map[rune]float64 {
	freq := make(map[rune]float64, len(s))
	var total float64

	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}

		freq[r]++
		total++
	}

	if total == 0 {
		return freq
	}

	for r, count := range freq {
		freq[r] = count / total
	}
	return freq
}

func drawChart(freq map[rune]float64) barchart.Model {
	keys := make([]rune, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	data := make([]barchart.BarData, 0, len(keys))
	for _, k := range keys {
		data = append(data, barchart.BarData{
			Label: string(k),
			Values: []barchart.BarValue{
				{
					Name:  string(k),
					Value: freq[k],
					Style: barStyle,
				},
			},
		})
	}

	bc := barchart.New(chartWidth, chartHeight,
		barchart.WithDataSet(data),
		barchart.WithBarWidth(1),
		barchart.WithBarGap(1),
		barchart.WithStyles(barStyle, labelStyle),
	)
	bc.Draw()
	return bc
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
