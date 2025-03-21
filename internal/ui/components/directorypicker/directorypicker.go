package directorypicker

import (
	"errors"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var ErrDirectoryNotSelected error = errors.New("you must select a directory")

type Model struct {
	fpModel     *filepicker.Model
	focused     bool
	selectedDir string
	err         error
	blurStyle   lipgloss.Style
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m Model) Error() error {
	return m.err
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.fpModel.KeyMap.Down, m.fpModel.KeyMap.Up, m.fpModel.KeyMap.Select,
		m.fpModel.KeyMap.Back, m.fpModel.KeyMap.Open}
}

func (m Model) Value() string {
	return m.selectedDir
}

func (m Model) View() string {
	if m.focused {
		return m.fpModel.View()
	}
	return m.blurStyle.Render(lo.Ternary(m.selectedDir != "", m.selectedDir, "No Selection"))
}

func (m *Model) Update(msg tea.Msg) (types.Input[string], tea.Cmd) {

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.fpModel.Height = v.Height
	case tea.KeyMsg:
		slog.Debug("directory finder update", "message", v.String(), "model_select_keys", m.fpModel.KeyMap.Select.Keys())
	}

	fpModel, cmd := m.fpModel.Update(msg)
	m.fpModel = &fpModel

	if didSelect, path := m.fpModel.DidSelectFile(msg); didSelect {
		m.selectedDir = path
		m.err = nil
		slog.Debug("selected in directory picker")
	}

	if didSelect, path := m.fpModel.DidSelectDisabledFile(msg); didSelect {

		m.err = errors.New(path + " is not valid.")
		m.selectedDir = ""
	}

	return m, cmd
}

func New() types.Input[string] {
	fpModel := filepicker.New()
	fpModel.FileAllowed = false
	fpModel.DirAllowed = true
	fpModel.AutoHeight = false
	fpModel.Height = 10
	fpModel.KeyMap.Select = key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "select"))
	fpModel.KeyMap.Open.SetKeys(append(fpModel.KeyMap.Open.Keys(), "ctrl+o")...)
	fpModel.CurrentDirectory = getDirectory()

	return &Model{fpModel: &fpModel, err: ErrDirectoryNotSelected, selectedDir: ""}
}

func (m Model) Init() tea.Cmd {
	slog.Debug("directory picker init")
	return m.fpModel.Init()
}

func (m *Model) SetValue(path string) error {
	m.selectedDir = path
	m.err = nil
	return nil
}

func getDirectory() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("could not get user home directory", "error", err)
	}

	return dir
}
