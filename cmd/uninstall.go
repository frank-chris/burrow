package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/frank-chris/burrow/internal/constants"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all burrow data from your system",
	Long:  `Removes credentials, logs, downloaded binaries, and all other burrow data from ~/.burrow/.`,
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	dir, err := burrowDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Nothing to uninstall. Burrow has not been set up on this machine.")
		return nil
	}

	fmt.Printf("This will permanently delete %s\n", dir)
	fmt.Print("Are you sure? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		fmt.Println("Aborted.")
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("could not remove %s: %w", dir, err)
	}

	fmt.Println("Burrow data removed.")
	fmt.Println("To remove the burrow binary itself, delete it from your PATH manually.")
	return nil
}

func burrowDir() (string, error) {
	if dir := os.Getenv(constants.ConfigDirEnvVar); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, constants.ConfigDirName), nil
}
