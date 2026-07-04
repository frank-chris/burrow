package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/frank-chris/burrow/internal/config"
	"github.com/frank-chris/burrow/internal/constants"
	"github.com/frank-chris/burrow/internal/install"
	"github.com/frank-chris/burrow/internal/provider/cloudflare"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up burrow with your provider credentials",
	Long:  `Walks through first-time setup: provider selection, API token entry, and account configuration.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	if config.AuthExists() {
		if !install.IsInstalled() {
			fmt.Print("Installing cloudflared...")
			if err := install.Install(); err != nil {
				fmt.Println(" failed.")
				return err
			}
			fmt.Println(" ok.")
			fmt.Println("Burrow is ready.")
			return nil
		}
		fmt.Print("Burrow is already initialized. Overwrite existing credentials? [y/N] ")
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Println("Setting up Burrow with Cloudflare Tunnel.")
	fmt.Println()
	fmt.Println("You can find your API token at: " + constants.CloudflareAPITokenURL)
	fmt.Println("You can find your Account ID at: " + constants.CloudflareAccountURL + " (right sidebar)")
	fmt.Println()

	fmt.Print("Cloudflare API Token: ")
	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("could not read API token: %w", err)
	}
	apiToken := strings.TrimSpace(string(tokenBytes))
	if apiToken == "" {
		return fmt.Errorf("API token cannot be empty")
	}

	fmt.Print("Cloudflare Account ID: ")
	accountID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("could not read account ID: %w", err)
	}
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	fmt.Println()
	fmt.Print("Validating credentials...")
	client := cloudflare.New(apiToken, accountID)
	if err := client.Validate(); err != nil {
		fmt.Println(" failed.")
		return err
	}
	fmt.Println(" ok.")

	if err := config.SaveAuth(&config.Auth{
		Provider:  "cloudflare",
		APIToken:  apiToken,
		AccountID: accountID,
	}); err != nil {
		return err
	}

	fmt.Println()
	fmt.Print("Installing cloudflared...")
	if err := install.Install(); err != nil {
		fmt.Println(" failed.")
		return err
	}
	fmt.Println(" ok.")

	fmt.Println()
	fmt.Println("Burrow is ready. Next steps:")
	fmt.Println("  1. Add a .burrow.yaml to your project")
	fmt.Println("  2. Run `burrow up` to start your tunnels")
	return nil
}
