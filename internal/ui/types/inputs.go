package types

import tea "github.com/charmbracelet/bubbletea"

type Input[T comparable] interface {
	Update(tea.Msg) (Input[T], tea.Cmd)
	View() string
	Focus() tea.Cmd
	Blur()
	Error() error
	Value() T
}

type AddDownloadMsg struct {
	URL       string
	QueueName string
	FileName  string
}
