package cmd

import (
	"fmt"
	"os"

	"github.com/frank-chris/burrow/internal/state"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all running tunnels",
	Long:  `Stops all tunnels started by the last 'burrow up' session.`,
	RunE:  runDown,
}

func runDown(cmd *cobra.Command, args []string) error {
	tunnels, err := state.LoadPIDs()
	if err != nil {
		return err
	}
	if len(tunnels) == 0 {
		fmt.Println("No tunnels are running.")
		return nil
	}

	for _, t := range tunnels {
		proc, err := os.FindProcess(t.PID)
		if err != nil {
			fmt.Printf("  [skip] %s - process not found\n", t.Name)
			continue
		}
		if err := proc.Kill(); err != nil {
			fmt.Printf("  [fail] %s - %s\n", t.Name, err)
			continue
		}
		fmt.Printf("  [down] %s\n", t.Name)
	}

	state.ClearPIDs()
	return nil
}
