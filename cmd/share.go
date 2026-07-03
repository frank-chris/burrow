package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/frank-chris/burrow/internal/tunnel"
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share <port>",
	Short: "Quickly share a local port publicly",
	Long:  `Starts a one-off tunnel for the given port without requiring a .burrow.yaml config. Uses a temporary trycloudflare.com URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShare,
}

func runShare(cmd *cobra.Command, args []string) error {
	if err := tunnel.CheckCloudflared(); err != nil {
		return err
	}

	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q — must be a number between 1 and 65535", args[0])
	}

	fmt.Printf("Sharing localhost:%d...\n\n", port)

	qt, err := tunnel.StartQuickTunnel(port)
	if err != nil {
		return err
	}

	url, err := qt.WaitForURL(30 * time.Second)
	if err != nil {
		return err
	}

	fmt.Println("Public URL:", url)
	fmt.Println()
	fmt.Println("Note: this URL is temporary and will stop working when you press Ctrl+C.")
	fmt.Println("Use `burrow up` with a .burrow.yaml for persistent named URLs.")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nStopping...")
	return qt.Stop()
}
