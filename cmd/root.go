package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "burrow",
	Short: "Persistent tunnel URLs for development teams",
	Long:  `Burrow creates named, persistent tunnel URLs backed by your chosen provider, committed to your repo so your whole team shares the same domains every time.`,
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(uninstallCmd)
}
