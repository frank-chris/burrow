package cmd

import (
	"fmt"

	"github.com/frank-chris/burrow/internal/config"
	"github.com/frank-chris/burrow/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all active tunnels",
	Long:  `Displays the current state of all tunnels started by the last 'burrow up' session.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	tunnels, err := state.LoadPIDs()
	if err != nil {
		return err
	}
	if len(tunnels) == 0 {
		fmt.Println("No tunnels are running.")
		return nil
	}

	// Load config for domain names - optional, gracefully ignored if not found
	domains := make(map[string]string)
	if cfg, err := config.Load(""); err == nil {
		for _, t := range cfg.Tunnels {
			domains[t.Name] = t.Domain
		}
	}

	fmt.Println()
	for _, t := range tunnels {
		status := "running"
		if !state.IsProcessAlive(t.PID) {
			status = "stopped"
		}
		if domain, ok := domains[t.Name]; ok {
			fmt.Printf("  [%-7s] %-20s https://%s\n", status, t.Name, domain)
		} else {
			fmt.Printf("  [%-7s] %-20s (pid %d)\n", status, t.Name, t.PID)
		}
	}
	fmt.Println()
	return nil
}
