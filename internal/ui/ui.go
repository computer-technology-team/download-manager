package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/tabs"
	"github.com/computer-technology-team/download-manager.git/internal/ui/views"
)

func NewDownloadManagerProgram(ctx context.Context, queueManager queues.QueueManager) (*tea.Program, error) {
	downloadsList, err := views.NewDownloadsList(ctx, queueManager)
	if err != nil {
		return nil, fmt.Errorf("could not create downloads list: %w", err)
	}

	queueList, err := views.NewQueuesList(ctx, queueManager)
	if err != nil {
		return nil, fmt.Errorf("could not create queue list: %w", err)
	}

	addDonwload, err := views.NewAddDownloadPane(ctx, queueManager)
	if err != nil {
		return nil, fmt.Errorf("could not create add download: %w", err)
	}

	tabsModel := tabs.New(1,
		tabs.Tab{Name: "Add Download", View: addDonwload},
		tabs.Tab{Name: "Downloads List", View: downloadsList},
		tabs.Tab{Name: "Queues List", View: queueList},
	)

	downloadManagerM := newDownloadManagerViewModel(tabsModel)

	return tea.NewProgram(downloadManagerM), nil
}
