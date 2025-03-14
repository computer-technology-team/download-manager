package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/ui/components/tabs"
	"github.com/computer-technology-team/download-manager.git/internal/ui/panes"
)

func NewDownloadManagerProgram() *tea.Program {
	tabsModel := tabs.New(1,
		tabs.Tab{Name: "Add Download", Pane: panes.NewAddDownloadPane()},
		tabs.Tab{Name: "Downloads List", Pane: panes.NewSamplePane("Downloads List")},
		tabs.Tab{Name: "Queues List", Pane: panes.NewSamplePane("Queues List")},
	)

	return tea.NewProgram(tabsModel)
}
