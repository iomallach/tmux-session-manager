package cmd

import (
	"fmt"
	"os"

	"github.com/iomallach/tmux-session-manager/internal/tsm"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	rootCmd.AddCommand(&sessionCmd)
}

var sessionCmd = cobra.Command{
	Use:   "sessions",
	Short: "Manage tmux sessions",
	Run: func(cmd *cobra.Command, args []string) {
		sessions := tsm.TmuxListSessions()

		p := tea.NewProgram(tsm.InitialModel(sessions))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}
