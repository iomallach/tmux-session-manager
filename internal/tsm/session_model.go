package tsm

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/op/redlog/pkg/catppuccin"
)

var (
	// TODO: tidy up styles
	catppuccinStyle = catppuccin.Macchiato
	rootStyle       = lipgloss.NewStyle().Background(catppuccinStyle.Base()).Width(40).Align(lipgloss.Center)
	helpStyle       = lipgloss.NewStyle().Foreground(catppuccinStyle.Green()).Background(catppuccinStyle.Base()).Align(lipgloss.Left).Italic(true).Faint(true)
	headerStyle     = lipgloss.NewStyle().Background(catppuccinStyle.Base()).Bold(true).PaddingTop(1).PaddingBottom(1).Width(30).Align(lipgloss.Center)
	selectedStyle   = lipgloss.NewStyle().Foreground(catppuccinStyle.Mauve()).Background(catppuccinStyle.Base())
	listStyle       = lipgloss.NewStyle().Background(catppuccinStyle.Base()).Width(30).Align(lipgloss.Left).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
)

type (
	State int
	Input int
)

const (
	MANAGE_STATE State = iota
	CREATE_STATE
	RENAME_STATE
)

const (
	NEW_SESSION_INPUT Input = iota
	RENAME_SESSION_INPUT
)

type model struct {
	choices         []string
	cursor          int
	state           State
	inputs          []textinput.Model
	focused         Input
	filtering       bool
	filtering_input textinput.Model
	tmux            Tmuxer
}

func createSessionInputBuble(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 20
	input.Width = 20
	input.Prompt = "➤"
	input.Validate = func(string) error { return nil }
	return input
}

func createFilteringInputBubble() textinput.Model {
	filtering_input := textinput.New()
	filtering_input.Placeholder = "type to filter"
	filtering_input.Width = 30
	filtering_input.Prompt = "➤"
	filtering_input.Validate = func(string) error { return nil }
	return filtering_input
}

func InitialSessionModel(tmux Tmuxer) model {
	var inputs []textinput.Model = make([]textinput.Model, 2)
	inputs[NEW_SESSION_INPUT] = createSessionInputBuble("New session name")
	inputs[RENAME_SESSION_INPUT] = createSessionInputBuble("Rename session")
	filtering_input := createFilteringInputBubble()

	choices := tmux.TmuxListSessions()

	return model{
		choices:         choices,
		state:           MANAGE_STATE,
		inputs:          inputs,
		filtering:       false,
		filtering_input: filtering_input,
		tmux:            tmux,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case MANAGE_STATE:
		return m.updateManageState(msg)
	case CREATE_STATE, RENAME_STATE:
		return m.updateInputState(msg)
	}

	return m, nil
}

func (m model) updateManageState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			m.filtering_input.Focus()
			switch msg.String() {
			case "enter":
				value := m.filtering_input.Value()
				var filtered_choices []string
				for _, choice := range m.choices {
					if strings.HasPrefix(choice, value) {
						filtered_choices = append(filtered_choices, choice)
					}
				}
				m.choices = filtered_choices
				m.filtering = false
				m.filtering_input.Reset()
				return m, cmd
			case "esc":
				m.filtering = false
				m.filtering_input.Reset()
				return m, cmd
			}
			m.filtering_input, cmd = m.filtering_input.Update(msg)
		} else {
			switch msg.String() {
			case "ctrl+p", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "ctrl+n", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "d":
				err := m.tmux.TmuxKillSession(m.choices[m.cursor])
				if err == nil {
					m.choices = append(m.choices[:m.cursor], m.choices[m.cursor+1:]...)
				}
			case "enter":
				err := m.tmux.TmuxSwitchSession(m.choices[m.cursor])
				if err == nil {
					return m, tea.Quit
				}
			case "c":
				m.state = CREATE_STATE
				m.focused = NEW_SESSION_INPUT
			case "r":
				m.state = RENAME_STATE
				m.focused = RENAME_SESSION_INPUT
			case "/":
				m.filtering = true
			case "esc":
				m.choices = m.tmux.TmuxListSessions()
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
	}

	return m, cmd
}

func (m model) updateInputState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.inputs[m.focused].Focus()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			sessionName := m.inputs[m.focused].Value()
			switch m.focused {
			case NEW_SESSION_INPUT:
				err := m.tmux.TmuxCreateSession(sessionName)
				if err == nil {
					m.state = MANAGE_STATE
					m.choices = m.tmux.TmuxListSessions()
				}
			case RENAME_SESSION_INPUT:
				err := m.tmux.TmuxRenameSession(m.choices[m.cursor], sessionName)
				if err == nil {
					m.state = MANAGE_STATE
					m.choices = m.tmux.TmuxListSessions()
				}
			}
		}
	}

	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m model) viewManageState() string {
	choices := list.New()
	for i, choice := range m.choices {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
			choices.Item(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, choice)))
		} else {
			choices.Item(fmt.Sprintf("%s %s", cursor, choice))
		}
	}

	// TODO: explore if this can be used instead of the manual cursor
	choices = choices.Enumerator(blankEnumerator)

	if m.filtering {
		return rootStyle.Render(
			fmt.Sprintf(
				"%s\n%s\n%s\n%s",
				headerStyle.Render("Sessions:"),
				listStyle.Render(choices.String()),
				m.filtering_input.View(),
				helpStyle.Render("enter: confirm • esc: cancel"),
			),
		)
	} else {
		return rootStyle.Render(
			fmt.Sprintf(
				"%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
				headerStyle.Render("Sessions:"),
				listStyle.Render(choices.String()),
				helpStyle.Render("q, ctrl+c: exit • d: kill session"),
				helpStyle.Render("c: create session • r: rename session"),
				helpStyle.Render("enter: switch session"),
				helpStyle.Render("k, ctrl+p: up • j, ctrl+n: down"),
				helpStyle.Render("/: search"),
				helpStyle.Render("esc: reset search"),
			),
		)
	}
}

func (m model) viewInputState() string {
	var actionString string
	switch m.focused {
	case NEW_SESSION_INPUT:
		actionString = "Create session:"
	case RENAME_SESSION_INPUT:
		actionString = "Rename session:"
	}

	return rootStyle.Render(
		fmt.Sprintf(
			"%s\n%s",
			headerStyle.Render(actionString),
			m.inputs[m.focused].View(),
		),
	)
}

func (m model) View() string {
	switch m.state {
	case MANAGE_STATE:
		return m.viewManageState()
	case CREATE_STATE, RENAME_STATE:
		return m.viewInputState()
	default:
		return m.viewManageState()
	}
}

func blankEnumerator(l list.Items, i int) string {
	return ""
}
