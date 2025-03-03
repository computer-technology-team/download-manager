package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/computer-technology-team/download-manager.git/internal/ui"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download-manager",
		Short: "Starts download manager TUI in default state",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("starting download manager tui program")

			_, err := ui.NewDownloadManagerProgram().Run()
			return err
		},
	}
	return cmd
}
