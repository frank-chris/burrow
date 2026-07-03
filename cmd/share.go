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

var (
	sharePassword string
	shareTTL      string
)

var shareCmd = &cobra.Command{
	Use:   "share <port>",
	Short: "Quickly share a local port publicly",
	Long:  `Starts a one-off tunnel for the given port without requiring a .burrow.yaml config. Uses a temporary trycloudflare.com URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShare,
}

func init() {
	shareCmd.Flags().StringVarP(&sharePassword, "password", "p", "", "Require a password to access the tunnel")
	shareCmd.Flags().StringVar(&shareTTL, "ttl", "", "Auto-expire the tunnel after this duration (e.g. 30m, 2h)")
}

func runShare(cmd *cobra.Command, args []string) error {
	if err := tunnel.CheckCloudflared(); err != nil {
		return err
	}

	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q - must be a number between 1 and 65535", args[0])
	}

	var ttl time.Duration
	if shareTTL != "" {
		ttl, err = time.ParseDuration(shareTTL)
		if err != nil {
			return fmt.Errorf("invalid --ttl %q - use Go duration format, e.g. 30m or 2h", shareTTL)
		}
		if ttl <= 0 {
			return fmt.Errorf("--ttl must be a positive duration")
		}
	}

	opts := tunnel.QuickOptions{
		Password: sharePassword,
		TTL:      ttl,
	}

	fmt.Printf("Sharing localhost:%d...\n\n", port)

	qt, err := tunnel.StartQuickTunnel(port, opts)
	if err != nil {
		return err
	}

	url, err := qt.WaitForURL(30 * time.Second)
	if err != nil {
		return err
	}

	fmt.Println("Public URL:", url)
	fmt.Println()

	if sharePassword != "" {
		fmt.Println("Password protection: enabled (leave username blank, enter the password you set)")
	}
	if ttl > 0 {
		fmt.Printf("Auto-expires: %s from now\n", ttl)
		fmt.Println()
	}

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
