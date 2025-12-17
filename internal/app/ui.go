package app

import (
"fmt"
"os"
"strings"

tea "github.com/charmbracelet/bubbletea"
)

type uiState int

const (
stateMenu uiState = iota
stateFolderSelect
)

type model struct {
	state          uiState
	choices        []string
	cursor         int
	selected       string
	
	folders        []string
	folderCursor   int
	selectedFolder string
	err            error
}

func initialModel() model {
	return model{
		state:   stateMenu,
		choices: []string{"Init", "Run", "Quit"},
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.selected = "quit"
			return m, tea.Quit
		case "q":
			if m.state == stateFolderSelect {
				m.state = stateMenu
				return m, nil
			}
			m.selected = "quit"
			return m, tea.Quit
		case "up", "k":
			if m.state == stateMenu {
				if m.cursor > 0 {
					m.cursor--
				}
			} else if m.state == stateFolderSelect {
				if m.folderCursor > 0 {
					m.folderCursor--
				}
			}
		case "down", "j":
			if m.state == stateMenu {
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			} else if m.state == stateFolderSelect {
				if m.folderCursor < len(m.folders)-1 {
					m.folderCursor++
				}
			}
		case "enter":
			if m.state == stateMenu {
				choice := m.choices[m.cursor]
				if choice == "Run" {
					folders, err := getFolders()
					if err != nil {
						m.err = err
						return m, tea.Quit
					}
					m.folders = folders
					m.state = stateFolderSelect
					m.folderCursor = 0
					return m, nil
				}
				m.selected = strings.ToLower(choice)
				return m, tea.Quit
			} else if m.state == stateFolderSelect {
				m.selected = "run"
				m.selectedFolder = m.folders[m.folderCursor]
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	s := "Scaffo CLI\n\n"

	if m.state == stateMenu {
		s += "Select a command:\n"
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else if m.state == stateFolderSelect {
		s += "Select source folder:\n"
		for i, folder := range m.folders {
			cursor := " "
			if m.folderCursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, folder)
		}
	}

	s += "\nUse ↑/↓ to move, Enter to select, q to quit/back."
	return s
}

func getFolders() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var folders []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			folders = append(folders, e.Name())
		}
	}
	folders = append(folders, "Other (enter path)")
	return folders, nil
}

// RunUI runs the Bubble Tea UI and returns the selected command and argument (if any)
func RunUI() (string, string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", "", err
	}
	finalModel := m.(model)
	return finalModel.selected, finalModel.selectedFolder, finalModel.err
}
