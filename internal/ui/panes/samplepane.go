package panes

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/ui/components/sampletext"
)

type samplePane struct {
	tea.Model
}

func (s samplePane) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

// ShortHelp implements types.Pane.
func (s samplePane) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithHelp("some key", "some help")),
	}
}

func NewSamplePane(s string) Pane {
	return samplePane{
		Model: sampletext.New(s),
	}
}
