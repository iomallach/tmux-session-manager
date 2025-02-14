package tsm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func TmuxListSessions() []string {
	cmd := exec.Command("tmux", "list-sessions")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running tmux list-sessions: %v", err)
		os.Exit(1)
	}
	sessions := strings.Split(string(out[:]), "\n")
	cleanedSessions := make([]string, 0)
	for _, item := range sessions {
		cleanedItem := strings.Split(item, ":")[0]
		if len(cleanedItem) > 0 {
			cleanedSessions = append(cleanedSessions, cleanedItem)
		}
	}

	return cleanedSessions
}

func TmuxKillSession(session string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", session)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error killing session: %v", err)
		return err
	}

	return nil
}

func TmuxSwitchSession(session string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", session)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error switching session: %v", err)
	}

	return nil
}

func TmuxCreateSession(session string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", session)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating session: %v", err)
	}

	return nil
}

func TmuxRenameSession(oldSession string, session string) error {
	cmd := exec.Command("tmux", "rename-session", "-t", oldSession, session)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error renaming session: %v", err)
	}

	return nil
}
