package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/frank-chris/burrow/internal/config"
	"github.com/frank-chris/burrow/internal/provider/cloudflare"
	"github.com/frank-chris/burrow/internal/state"
	"github.com/frank-chris/burrow/internal/tunnel"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [name]",
	Short: "Start tunnels defined in .burrow.yaml",
	Long:  `Starts all tunnels defined in .burrow.yaml. Pass an optional name to start a single tunnel.`,
	RunE:  runUp,
}

func runUp(cmd *cobra.Command, args []string) error {
	if err := tunnel.CheckCloudflared(); err != nil {
		return err
	}

	auth, err := config.LoadAuth()
	if err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	if len(args) == 1 {
		name := args[0]
		found := false
		for _, t := range cfg.Tunnels {
			if t.Name == name {
				cfg.Tunnels = []config.TunnelConfig{t}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tunnel %q not found in .burrow.yaml", name)
		}
	}

	client := cloudflare.New(auth.APIToken, auth.AccountID)
	manager := tunnel.NewManager(client)

	fmt.Println("Starting tunnels...")
	if err := manager.StartAll(cfg); err != nil {
		manager.StopAll()
		return err
	}

	if err := state.SavePIDs(manager.TunnelProcesses()); err != nil {
		manager.StopAll()
		return fmt.Errorf("could not save process state: %w", err)
	}

	fmt.Println("\nAll tunnels running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down...")
	manager.StopAll()
	state.ClearPIDs()
	return nil
}
