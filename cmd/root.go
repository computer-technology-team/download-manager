package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download-manager",
		Short: "Starts download manager TUI in default state",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("starting download manager tui program")

			ctx := cmd.Context()

			db, err := state.SetupDatabase(ctx)
			if err != nil {
				slog.Error("failed to setup database", "error", err)
				return err
			}

			queueManager := queues.New(db)

			_, err = ui.NewDownloadManagerProgram(queueManager).Run()
			return err
		},
	}
	return cmd
}
