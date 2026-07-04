package cmd

import (
	"fmt"
	"os"

	"github.com/frank-chris/burrow/internal/state"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [name]",
	Short: "Stop all running tunnels, or a specific one by name",
	Long:  `Stops all tunnels started by 'burrow up', or a single tunnel if a name is given.`,
	Args:  cobra.MaximumNArgs(1),
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

	if len(args) == 1 {
		name := args[0]
		var found *state.TunnelProcess
		for i := range tunnels {
			if tunnels[i].Name == name {
				found = &tunnels[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("no running tunnel named %q", name)
		}
		proc, err := os.FindProcess(found.PID)
		if err != nil {
			return fmt.Errorf("process not found for %q", name)
		}
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("could not stop %q: %w", name, err)
		}
		fmt.Printf("  [down] %s\n", name)
		return state.RemovePID(name)
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
