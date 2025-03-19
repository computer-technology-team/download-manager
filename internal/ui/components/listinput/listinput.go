package listinput

import (
	"errors"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var (
	ErrItemNotFound = errors.New("item not found")
	docStyle        = lipgloss.NewStyle().Margin(1, 2)

	// Style for the selected item when not focused
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CDD6F4"))
)

type item struct {
	id          int
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list  list.Model
	focus bool
}

// GetSelected returns the currently selected item
func (m *model) GetSelected() (list.Item, bool) {
	if m.list.SelectedItem() == nil {
		return nil, false
	}
	return m.list.SelectedItem(), true
}

// GetSelectedTitle returns the title of the currently selected item
func (m *model) GetSelectedTitle() string {
	if selected, ok := m.GetSelected(); ok {
		return selected.(item).Title()
	}
	return "None selected"
}

// Blur implements types.Input.
func (m *model) Blur() {
	m.focus = false
}

// Focus implements types.Input.
func (m *model) Focus() tea.Cmd {
	m.focus = true
	return nil
}

func (m model) Error() error {
	return nil
}

func (m model) Value() string {
	return m.list.SelectedItem().(item).title
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (types.Input[string], tea.Cmd) {
	// Only process key events when focused
	if !m.focus {
		return &m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window size changes
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(10)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return &m, cmd
}

func (m model) View() string {
	if !m.focus {
		// When not focused, just show the selected item
		selected := m.GetSelectedTitle()
		return selectedItemStyle.Render(selected)
	}

	// When focused, show the list with a reasonable height
	// Setting to 10 to ensure items are visible (3 is too small)
	m.list.SetHeight(10)

	// Make sure the list is visible by setting appropriate styles
	m.list.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89B4FA")).
		MarginLeft(2)

	// Ensure the selected item is visible
	if m.list.Index() < 0 && len(m.list.Items()) > 0 {
		m.list.Select(0)
	}

	return docStyle.Render(m.list.View())
}

func (m model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m model) FullHelp() [][]key.Binding {
	return m.list.FullHelp()
}

func (m *model) SetValue(id string) error {
	for i, itemI := range m.list.Items() {
		if strconv.Itoa(itemI.(item).id) == id {
			m.list.Select(i)
			return nil
		}
	}

	return ErrItemNotFound
}

func New(title string, barItemNameSingular, barItemNamePlural string) *model {
	// Create a delegate with custom styles
	delegate := list.NewDefaultDelegate()

	// Customize the delegate styles to ensure visibility
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#F38BA8")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#CBA6F7"))

	items := []list.Item{
		item{id: 1, title: "Default", desc: "Default download queue"},
		item{id: 2, title: "High Priority", desc: "For urgent downloads"},
		item{id: 3, title: "Low Priority", desc: "For background downloads"},
		item{id: 4, title: "Media", desc: "For videos and music"},
		item{id: 5, title: "Documents", desc: "For PDFs and other documents"},
	}

	// Create the list with initial dimensions
	listModel := list.New(items, delegate, 30, 10)
	listModel.Title = title

	// Set custom keybindings
	listModel.KeyMap.CursorDown.SetKeys("p")
	listModel.KeyMap.CursorUp.SetKeys("n")
	listModel.KeyMap.CursorDown.SetHelp("p", "prev")
	listModel.KeyMap.CursorUp.SetHelp("n", "next")

	// Set initial selection
	listModel.Select(0)

	// Add help text
	listModel.SetStatusBarItemName(barItemNameSingular, barItemNamePlural)
	listModel.SetShowStatusBar(true)
	listModel.SetShowTitle(true)
	listModel.SetShowHelp(false)

	return &model{
		list:  listModel,
		focus: false,
	}
}
