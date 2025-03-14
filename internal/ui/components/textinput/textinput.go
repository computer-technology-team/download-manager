package textinput

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type Model struct {
	*textinput.Model
}

func (m Model) Update(msg tea.Msg) (types.Input[string], tea.Cmd) {
	tmpModel, cmd := m.Model.Update(msg)
	m.Model = &tmpModel
	return &m, cmd
}

func (m Model) Error() error {
	return m.Err
}

func (m Model) Value() string {
	return m.Model.Value()
}

func New() Model {
	return Model{
		Model: lo.ToPtr(textinput.New()),
	}
}
