package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/ui/components/sampletext"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type sampleView struct {
	tea.Model
}

func (s sampleView) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

func (s sampleView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithHelp("some key", "some help")),
	}
}

func (s sampleView) Update(_ tea.Msg) (types.View, tea.Cmd) {
	return s, nil
}

func NewSampleView(s string) types.View {
	return sampleView{
		Model: sampletext.New(s),
	}
}
