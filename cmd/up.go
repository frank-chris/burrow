package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/frank-chris/burrow/internal/clipboard"
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

	var named, quick []config.TunnelConfig
	for _, t := range cfg.Tunnels {
		if t.Domain != "" {
			named = append(named, t)
		} else {
			quick = append(quick, t)
		}
	}

	var allPIDs []state.TunnelProcess
	var manager *tunnel.Manager
	var quickProcs []*tunnel.QuickTunnel

	if len(named) > 0 {
		auth, err := config.LoadAuth()
		if err != nil {
			return fmt.Errorf("tunnels with domains require credentials - run `burrow init` first: %w", err)
		}
		namedCfg := &config.Config{Provider: cfg.Provider, Tunnels: named}
		client := cloudflare.New(auth.APIToken, auth.AccountID)
		manager = tunnel.NewManager(client)
		fmt.Println("Starting tunnels...")
		if err := manager.StartAll(namedCfg); err != nil {
			manager.StopAll()
			return err
		}
		allPIDs = append(allPIDs, manager.TunnelProcesses()...)
	}

	if len(quick) > 0 {
		fmt.Println("Starting quick tunnels...")
		for _, t := range quick {
			tunnel.WarnIfPortClosed(t.Port)
			qt, err := tunnel.StartQuickTunnel(t.Port, tunnel.QuickOptions{})
			if err != nil {
				stopQuick(quickProcs)
				if manager != nil {
					manager.StopAll()
				}
				return err
			}
			quickProcs = append(quickProcs, qt)
		}

		type result struct {
			index int
			url   string
			err   error
		}
		ch := make(chan result, len(quick))
		for i, qt := range quickProcs {
			go func() {
				url, err := qt.WaitForURL(30*time.Second, 0)
				ch <- result{index: i, url: url, err: err}
			}()
		}

		urls := make([]string, len(quick))
		for range quick {
			r := <-ch
			if r.err != nil {
				stopQuick(quickProcs)
				if manager != nil {
					manager.StopAll()
				}
				return r.err
			}
			urls[r.index] = r.url
		}

		for i, t := range quick {
			fmt.Printf("  [up] %s -> %s\n", t.Name, urls[i])
			allPIDs = append(allPIDs, state.TunnelProcess{Name: t.Name, PID: quickProcs[i].PID()})
			if len(cfg.Tunnels) == 1 {
				if err := clipboard.Copy(urls[i]); err == nil {
					fmt.Println("  (copied to clipboard)")
				}
			}
		}
	}

	if err := state.SavePIDs(allPIDs); err != nil {
		stopQuick(quickProcs)
		if manager != nil {
			manager.StopAll()
		}
		return fmt.Errorf("could not save process state: %w", err)
	}

	fmt.Println("\nAll tunnels running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down...")
	stopQuick(quickProcs)
	if manager != nil {
		manager.StopAll()
	}
	state.ClearPIDs()
	return nil
}

func stopQuick(procs []*tunnel.QuickTunnel) {
	for _, qt := range procs {
		qt.Stop()
	}
}
