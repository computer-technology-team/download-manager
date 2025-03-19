package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/ui/components/tabs"
	"github.com/computer-technology-team/download-manager.git/internal/ui/views"
)

func NewDownloadManagerProgram() *tea.Program {
	tabsModel := tabs.New(1,
		tabs.Tab{Name: "Add Download", View: views.NewAddDownloadPane()},
		tabs.Tab{Name: "Downloads List", View: views.NewDownloadsList()},
		tabs.Tab{Name: "Queues List", View: views.NewQueuesList()},
	)

	return tea.NewProgram(tabsModel)
}
