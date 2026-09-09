package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const maxBarWidth = 40

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Width(6)

	countStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(4).
			Align(lipgloss.Right)

	barStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	emptyStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("241"))
)

type screen int

const (
	screenInput screen = iota
	screenGraph
)

type model struct {
	screen   screen
	textarea textarea.Model
	graph    string
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type or paste text here, then press ctrl+d to analyze."
	ta.Prompt = ""
	ta.ShowLineNumbers = true
	ta.SetWidth(60)
	ta.SetHeight(12)
	ta.Focus()

	return model{screen: screenInput, textarea: ta}
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
			m.graph = createGraph(charFrequencies(m.textarea.Value()))
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
	var v tea.View
	v.AltScreen = true

	if m.screen == screenGraph {
		v.SetContent(m.viewGraph())
	} else {
		v.SetContent(m.viewInput())
	}
	return v
}

func (m model) viewInput() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Character Frequency Grapher"))
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
	b.WriteString(m.graph)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q to quit  ·  r to analyze new input"))
	return b.String()
}

func charFrequencies(s string) map[rune]int {
	freq := make(map[rune]int, len(s))
	for _, r := range s {
		freq[r]++
	}
	return freq
}

func createGraph(data map[rune]int) string {
	if len(data) == 0 {
		return emptyStyle.Render("input is empty!")
	}

	maxCount := 0
	for _, n := range data {
		maxCount = max(maxCount, n)
	}

	keys := make([]rune, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	rows := make([]string, 0, len(data))
	for _, k := range keys {
		n := data[k]
		barLen := int(math.Round(float64(n) / float64(maxCount) * maxBarWidth))
		bar := barStyle.Render(strings.Repeat("█", barLen))

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			labelStyle.Render(displayChar(k)),
			countStyle.Render(fmt.Sprintf("%d", n)),
			"  ▌"+bar,
		)
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func displayChar(r rune) string {
	switch r {
	case '\n':
		return `'\n'`
	case '\t':
		return `'\t'`
	case ' ':
		return "' '"
	default:
		return string(r)
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
