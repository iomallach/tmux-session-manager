package tsm

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type appState int

const (
	CHOOSING appState = iota
	SESSION_MANAGEMENT
	WINDOW_MANAGEMENT
	PANE_MANAGEMENT
)

type rootModel struct {
	state       appState
	activeModel tea.Model
}

func InitialRootModel() rootModel {
	return rootModel{state: CHOOSING, activeModel: nil}
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case CHOOSING:
		// TODO: implement
		return m.updateChoosingState(msg)
	case SESSION_MANAGEMENT:
		return m.activeModel.Update(msg)
	case WINDOW_MANAGEMENT:
		// TODO: implement
		return nil, nil
	case PANE_MANAGEMENT:
		// TODO: implement
		return nil, nil
	default:
		return m.activeModel.Update(msg)
	}
}

func (m rootModel) updateChoosingState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			// TODO: no need to pass the initial list, might as well keep inside the InitialModel
			m.activeModel = InitialModel(TmuxListSessions())
			m.state = SESSION_MANAGEMENT
		case "w":
			panic("not implemented")
		case "p":
			panic("not implemented")
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m rootModel) View() string {
	switch m.state {
	case CHOOSING:
		return m.viewChoosingState()
	case SESSION_MANAGEMENT:
		return m.activeModel.View()
	case WINDOW_MANAGEMENT:
		return ""
	case PANE_MANAGEMENT:
		return ""
	default:
		return m.activeModel.View()
	}
}

func (m rootModel) viewChoosingState() string {
	return rootStyle.Render(
		fmt.Sprintf(
			"%s\n%s\n%s",
			headerStyle.Render("Choose an action:"),
			helpStyle.Render("s: sessions • w: windows • p: panes"),
			helpStyle.Render("q: quit"),
		),
	)
}
