package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download-manager",
		Short: "Starts download manager TUI in default state",
	}
	return cmd
}
