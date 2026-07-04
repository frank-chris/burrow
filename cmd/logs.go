package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/frank-chris/burrow/internal/state"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "Show tunnel request logs",
	Long:  `Shows logs from running tunnels. Optionally filter by tunnel name with the first argument.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogs,
}

var (
	flagFollow bool
	flagLines  int
)

func init() {
	logsCmd.Flags().BoolVarP(&flagFollow, "follow", "f", false, "Follow log output in real time")
	logsCmd.Flags().IntVarP(&flagLines, "lines", "n", 20, "Number of past lines to show")
}

func runLogs(cmd *cobra.Command, args []string) error {
	logDir, err := state.LogDir()
	if err != nil {
		return err
	}

	var logFiles []struct{ name, path string }

	if len(args) == 1 {
		name := args[0]
		path, err := state.LogPath(name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			if isQuickTunnel(name) {
				return fmt.Errorf("tunnel %q is a quick tunnel - logs are only available for tunnels with a domain in .burrow.yaml", name)
			}
			return fmt.Errorf("no logs found for tunnel %q - has it been run with `burrow up`?", name)
		}
		logFiles = append(logFiles, struct{ name, path string }{name, path})
	} else {
		entries, err := os.ReadDir(logDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				pids, _ := state.LoadPIDs()
				if len(pids) > 0 {
					return fmt.Errorf("no logs found - running tunnels are quick tunnels (no domain). Logs are only available for tunnels with a domain in .burrow.yaml")
				}
				return fmt.Errorf("no logs found - run `burrow up` first")
			}
			return fmt.Errorf("could not read log directory: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
				name := strings.TrimSuffix(e.Name(), ".log")
				logFiles = append(logFiles, struct{ name, path string }{name, filepath.Join(logDir, e.Name())})
			}
		}
		if len(logFiles) == 0 {
			pids, _ := state.LoadPIDs()
			if len(pids) > 0 {
				return fmt.Errorf("no logs found - running tunnels are quick tunnels (no domain). Logs are only available for tunnels with a domain in .burrow.yaml")
			}
			return fmt.Errorf("no logs found - run `burrow up` first")
		}
	}

	multiTunnel := len(logFiles) > 1

	// Print last N lines from each file
	for _, lf := range logFiles {
		lines, err := lastNLines(lf.path, flagLines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not read logs for %q: %s\n", lf.name, err)
			continue
		}
		for _, line := range lines {
			if multiTunnel {
				fmt.Printf("[%s] %s\n", lf.name, line)
			} else {
				fmt.Println(line)
			}
		}
	}

	if !flagFollow {
		return nil
	}

	// Follow mode: tail each file in its own goroutine
	fmt.Println("--- following ---")
	var wg sync.WaitGroup
	for _, lf := range logFiles {
		wg.Add(1)
		go func(name, path string) {
			defer wg.Done()
			tailFile(name, path, multiTunnel)
		}(lf.name, lf.path)
	}
	wg.Wait()
	return nil
}

func lastNLines(path string, n int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}

func tailFile(name, path string, prefix bool) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open log file for %q: %s\n", name, err)
		return
	}
	defer f.Close()

	// Seek to end so we only show new lines
	f.Seek(0, io.SeekEnd)

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\n")
			if prefix {
				fmt.Printf("[%s] %s\n", name, line)
			} else {
				fmt.Println(line)
			}
		}
		if err != nil {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// isQuickTunnel reports whether name appears in the PID file but has no log file.
func isQuickTunnel(name string) bool {
	pids, err := state.LoadPIDs()
	if err != nil {
		return false
	}
	for _, p := range pids {
		if p.Name == name {
			return true
		}
	}
	return false
}
