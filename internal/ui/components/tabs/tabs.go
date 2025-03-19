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
	tabManager TabManager
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width

		newMsg := tea.WindowSizeMsg{Width: msg.Width - 2, Height: msg.Height - 10}

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

	}

	var childCmd tea.Cmd
	m.tabs[m.activeTab].View, childCmd = m.GetActiveTab().View.Update(msg)

	return m, childCmd
}

func (m model) View() string {
	doc := strings.Builder{}

	// Calculate the base width for each tab
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

		// Adjust borders for connecting tabs
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

		// Add extra width to the last tab to fill the space
		width := tabWidth
		if isLast {
			width = m.width - (tabWidth * (numTabs - 1))
		}

		style = style.Width(width - 2).Border(border) // -2 for borders
		renderedTabs = append(renderedTabs, style.Render(t.Name))
	}

	// Join tabs horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	// Render window content
	windowContent := windowStyle.Width(m.width - 2).Render(m.GetActiveTab().View.View()) // -2 for borders
	doc.WriteString(windowContent)
	doc.WriteString("\n")

	m.keymap.tabManager = m
	doc.WriteString(m.help.View(m.keymap))

	return doc.String()
}

func (m model) GetActiveTab() Tab {
	return m.tabs[m.activeTab]
}

func (k keymap) ShortHelp() []key.Binding {
	return slices.Concat([]key.Binding{k.helpToggle}, k.fkeys, k.getActiveTabShortHelp())
}

func (k keymap) FullHelp() [][]key.Binding {
	res := [][]key.Binding{k.fkeys, {k.helpToggle, k.quit, k.prevTab, k.nextTab}}
	res = append(res, k.getActiveTabFullHelp()...)
	return res
}

func (k keymap) getActiveTabShortHelp() []key.Binding {
	return k.tabManager.GetActiveTab().View.ShortHelp()
}

func (k keymap) getActiveTabFullHelp() [][]key.Binding {
	return k.tabManager.GetActiveTab().View.FullHelp()
}

func initKeymap(tabs []Tab) keymap {
	return keymap{
		fkeys: lo.Map(tabs, func(t Tab, i int) key.Binding {
			return key.NewBinding(
				key.WithKeys(fmt.Sprintf("f%d", i+1)),
				key.WithHelp(fmt.Sprintf("F%d", i+1), t.Name),
			)
		}),
		helpToggle: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggles full help"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
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

func New(defaultActive int, tabs ...Tab) tea.Model {
	helpModel := help.New()
	helpModel.ShowAll = true

	return model{
		tabs:      tabs,
		activeTab: min(len(tabs)-1, defaultActive),
		keymap:    initKeymap(tabs),
		help:      helpModel,
	}
}
