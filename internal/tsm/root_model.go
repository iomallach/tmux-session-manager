package tsm

import (
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
	activeModel tea.Model
}

func InitialRootModel() rootModel {
	return rootModel{activeModel: InitialSessionModel()}
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.activeModel.Update(msg)
}

func (m rootModel) View() string {
	return m.activeModel.View()
}
