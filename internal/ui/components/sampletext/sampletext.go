package sampletext

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model string

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) View() string {
	return string(m)
}

func New(s string) tea.Model {
	return model(s)
}
