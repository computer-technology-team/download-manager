package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/tabs"
	"github.com/computer-technology-team/download-manager.git/internal/ui/views"
)

func NewDownloadManagerProgram(ctx context.Context, queueManager queues.QueueManager) *tea.Program {
	tabsModel := tabs.New(1,
		tabs.Tab{Name: "Add Download", View: views.NewAddDownloadPane(ctx, queueManager)},
		tabs.Tab{Name: "Downloads List", View: views.NewDownloadsList(ctx, queueManager)},
		tabs.Tab{Name: "Queues List", View: views.NewQueuesList(ctx, queueManager)},
	)

	return tea.NewProgram(tabsModel)
}
