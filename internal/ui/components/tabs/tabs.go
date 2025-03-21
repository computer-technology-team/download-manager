package tabs

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var fKeyPattern = regexp.MustCompile(`f(\d+)`)

type Tab struct {
	Name string
	View types.View
}

type TabManager interface {
	GetActiveTab() Tab
}

type keymap struct {
	fkeys      []key.Binding
	helpToggle key.Binding
	quit       key.Binding
	nextTab    key.Binding
	prevTab    key.Binding
}

type model struct {
	tabs      []Tab
	activeTab int
	width     int

	help help.Model

	keymap keymap
}

func (m model) Init() tea.Cmd {
	return tea.Batch(lo.Map(m.tabs, func(t Tab, _ int) tea.Cmd {
		return t.View.Init()
	})...)
}

func (m model) Update(msg tea.Msg) (types.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width

		newMsg := tea.WindowSizeMsg{Width: msg.Width - 2, Height: msg.Height - 4}

		cmds := make([]tea.Cmd, len(m.tabs))
		for i, t := range m.tabs {
			m.tabs[i].View, cmds[i] = t.View.Update(newMsg)
		}

		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.nextTab):
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case key.Matches(msg, m.keymap.prevTab):
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case key.Matches(msg, m.keymap.helpToggle):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
		if submatches := fKeyPattern.FindStringSubmatch(msg.String()); len(submatches) >= 2 {
			selectedTab, err := strconv.Atoi(submatches[1])
			if err != nil {
				slog.LogAttrs(context.Background(), slog.LevelError, "could not parse F-key selected tab",
					slog.Any("error", err))
				return m, nil
			}
			if selectedTab > len(m.tabs) {
				return m, nil
			}
			m.activeTab = selectedTab - 1
			return m, nil
		}

		var childCmd tea.Cmd
		m.tabs[m.activeTab].View, childCmd = m.GetActiveTab().View.Update(msg)

		return m, childCmd
	default:
		cmds := make([]tea.Cmd, len(m.tabs))
		for i, t := range m.tabs {
			m.tabs[i].View, cmds[i] = t.View.Update(msg)
		}

		return m, tea.Batch(cmds...)
	}

}

func (m model) View() string {
	doc := strings.Builder{}

	numTabs := len(m.tabs)
	tabWidth := (m.width / numTabs)

	var renderedTabs []string

	for i, t := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.activeTab

		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}

		width := tabWidth
		if isLast {
			width = m.width - (tabWidth * (numTabs - 1))
		}

		style = style.Width(width - 2).Border(border) 
		renderedTabs = append(renderedTabs, style.Render(t.Name))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	windowContent := windowStyle.Width(m.width - 2).Render(m.GetActiveTab().View.View()) 
	doc.WriteString(windowContent)
	doc.WriteString("\n")

	return doc.String()
}

func (m model) GetActiveTab() Tab {
	return m.tabs[m.activeTab]
}

func (m model) ShortHelp() []key.Binding {
	return slices.Concat(m.keymap.fkeys, m.getActiveTabShortHelp())
}

func (m model) FullHelp() [][]key.Binding {
	res := [][]key.Binding{m.keymap.fkeys, {m.keymap.prevTab, m.keymap.nextTab}}
	res = append(res, m.getActiveTabFullHelp()...)
	return res
}

func (m model) getActiveTabShortHelp() []key.Binding {
	return m.GetActiveTab().View.ShortHelp()
}

func (m model) getActiveTabFullHelp() [][]key.Binding {
	return m.GetActiveTab().View.FullHelp()
}

func initKeymap(tabs []Tab) keymap {
	return keymap{
		fkeys: lo.Map(tabs, func(t Tab, i int) key.Binding {
			return key.NewBinding(
				key.WithKeys(fmt.Sprintf("f%d", i+1)),
				key.WithHelp(fmt.Sprintf("F%d", i+1), t.Name),
			)
		}),
		nextTab: key.NewBinding(
			key.WithKeys("right", "tab"),
			key.WithHelp("→/tab", "next tab"),
		),
		prevTab: key.NewBinding(
			key.WithKeys("left", "shift+tab"),
			key.WithHelp("←/shift+tab", "previous tab"),
		),
	}
}

func New(defaultActive int, tabs ...Tab) types.View {

	return model{
		tabs:      tabs,
		activeTab: min(len(tabs)-1, defaultActive),
		keymap:    initKeymap(tabs),
	}
}
