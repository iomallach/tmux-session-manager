package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iomallach/tmux-session-manager/internal/tsm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tsm",
	Short: "Tmux session manager is a very simple tui session manager for tmux",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tsm.InitialRootModel(&tsm.Tmux{}))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
