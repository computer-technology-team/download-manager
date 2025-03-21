package listinput

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var (
	ErrItemNotFound = errors.New("item not found")
	docStyle        = lipgloss.NewStyle().Margin(1, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CDD6F4"))
)

type Item struct {
	id    string
	title string
	desc  string
}

func NewItem(id, title, desc string) Item {
	return Item{
		id:    id,
		title: title,
		desc:  desc,
	}
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	list.Model
	focus bool
}

func (m *Model) GetSelected() (list.Item, bool) {
	if m.Model.SelectedItem() == nil {
		return nil, false
	}
	return m.Model.SelectedItem(), true
}

func (m *Model) GetSelectedTitle() string {
	if selected, ok := m.GetSelected(); ok {
		return selected.(Item).Title()
	}
	return "None selected"
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return nil
}

func (m Model) Error() error {
	return nil
}

func (m Model) Value() string {
	return m.Model.SelectedItem().(Item).id
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (types.Input[string], tea.Cmd) {

	if !m.focus {
		return &m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		m.Model.SetWidth(msg.Width - 4)
		m.Model.SetHeight(10)
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return &m, cmd
}

func (m Model) View() string {
	if !m.focus {

		selected := m.GetSelectedTitle()
		return selectedItemStyle.Render(selected)
	}

	m.Model.SetHeight(10)

	m.Model.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89B4FA")).
		MarginLeft(2)

	if m.Model.Index() < 0 && len(m.Model.Items()) > 0 {
		m.Model.Select(0)
	}

	return docStyle.Render(m.Model.View())
}

func (m Model) ShortHelp() []key.Binding {
	return m.Model.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	return m.Model.FullHelp()
}

func (m *Model) SetValue(id string) error {
	for i, itemI := range m.Model.Items() {
		if itemI.(Item).id == id {
			m.Model.Select(i)
			return nil
		}
	}

	return ErrItemNotFound
}

func (m *Model) FindItemIdx(id string) (idx int, found bool) {
	for i, item := range m.Items() {
		if item.(Item).id == id {
			idx = i
			found = true
			return
		}
	}

	return
}

func New(title string, barItemNameSingular, barItemNamePlural string, items []list.Item) *Model {

	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#F38BA8")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#CBA6F7"))

	listModel := list.New(items, delegate, 30, 10)
	listModel.Title = title

	listModel.KeyMap.CursorDown.SetKeys("p")
	listModel.KeyMap.CursorUp.SetKeys("n")
	listModel.KeyMap.CursorDown.SetHelp("p", "prev")
	listModel.KeyMap.CursorUp.SetHelp("n", "next")

	listModel.Select(0)

	listModel.SetStatusBarItemName(barItemNameSingular, barItemNamePlural)
	listModel.SetShowStatusBar(true)
	listModel.SetShowTitle(true)
	listModel.SetShowHelp(false)

	return &Model{
		Model: listModel,
		focus: false,
	}
}
