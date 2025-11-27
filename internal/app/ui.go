package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	cursor   int
	selected int
}

func initialModel() model {
	return model{
		choices:  []string{"Init", "Analyze", "Build Template", "Generate", "Run", "Quit"},
		cursor:   0,
		selected: -1,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.selected = len(m.choices) - 1
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Scaffo CLI - Select a command:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nUse ↑/↓ to move, Enter to select, q to quit."
	return s
}

// RunUI runs the Bubble Tea UI and returns the selected command as a string
func RunUI() (string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	selected := m.(model).selected
	switch selected {
	case 0:
		return "init", nil
	case 1:
		return "analyze", nil
	case 2:
		return "build-template", nil
	case 3:
		return "generate", nil
	case 4:
		return "run", nil
	default:
		return "quit", nil
	}
}
