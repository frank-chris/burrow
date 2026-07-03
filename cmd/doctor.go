package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/frank-chris/burrow/internal/config"
	"github.com/frank-chris/burrow/internal/constants"
	"github.com/frank-chris/burrow/internal/install"
	"github.com/frank-chris/burrow/internal/provider/cloudflare"
	"github.com/frank-chris/burrow/internal/state"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your burrow setup for problems",
	Long:  `Runs a series of checks on your burrow setup and reports any issues found.`,
	RunE:  runDoctor,
}


type check struct {
	label  string
	status string // ok, fail, warn, skip
	detail string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("Running diagnostics...")
	fmt.Println()

	var checks []check

	// Auth
	auth, err := config.LoadAuth()
	if err != nil {
		checks = append(checks,
			newCheck("fail", "Auth configured", "run `burrow init` to set up credentials"),
			newCheck("skip", "Cloudflare API reachable", ""),
			newCheck("skip", "API token valid", ""),
		)
	} else {
		checks = append(checks, newCheck("ok", "Auth configured", "provider: "+auth.Provider))

		// API reachability
		if !apiReachable() {
			checks = append(checks,
				newCheck("fail", "Cloudflare API reachable", "check your internet connection"),
				newCheck("skip", "API token valid", ""),
			)
		} else {
			checks = append(checks, newCheck("ok", "Cloudflare API reachable", ""))

			client := cloudflare.New(auth.APIToken, auth.AccountID)
			if err := client.Validate(); err != nil {
				checks = append(checks, newCheck("fail", "API token valid", "run `burrow init` to update your token"))
			} else {
				checks = append(checks, newCheck("ok", "API token valid", ""))
			}
		}
	}

	// cloudflared binary
	if install.IsInstalled() {
		version := cloudflaredVersion()
		checks = append(checks, newCheck("ok", "cloudflared installed", version))
	} else {
		checks = append(checks, newCheck("fail", "cloudflared installed", "run `burrow init` to install it"))
	}

	// Config file
	if _, err := config.Load(""); err == nil {
		checks = append(checks, newCheck("ok", ".burrow.yaml found", ""))
	} else {
		checks = append(checks, newCheck("warn", ".burrow.yaml found", "not found in current directory — needed for `burrow up`"))
	}

	// Running tunnels
	tunnels, _ := state.LoadPIDs()
	if len(tunnels) == 0 {
		checks = append(checks, newCheck("ok", "Running tunnels", "none"))
	} else {
		alive := 0
		for _, t := range tunnels {
			if state.IsProcessAlive(t.PID) {
				alive++
			}
		}
		checks = append(checks, newCheck("ok", "Running tunnels", fmt.Sprintf("%d of %d active", alive, len(tunnels))))
	}

	// Print results
	failures := 0
	for _, c := range checks {
		label := fmt.Sprintf("[%-4s]", c.status)
		if c.detail != "" {
			fmt.Printf("  %s %s — %s\n", label, c.label, c.detail)
		} else {
			fmt.Printf("  %s %s\n", label, c.label)
		}
		if c.status == "fail" {
			failures++
		}
	}

	fmt.Println()
	if failures > 0 {
		fmt.Printf("%d issue(s) found.\n", failures)
	} else {
		fmt.Println("All checks passed.")
	}
	return nil
}

func newCheck(status, label, detail string) check {
	return check{label: label, status: status, detail: detail}
}

func apiReachable() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(constants.CloudflareAPIBase)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func cloudflaredVersion() string {
	binPath, err := install.BinaryPath()
	if err != nil {
		return ""
	}
	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		return ""
	}
	// Output: "cloudflared version 2024.9.1 (built ...)"
	parts := strings.Fields(string(out))
	if len(parts) >= 3 {
		return parts[2]
	}
	return strings.TrimSpace(string(out))
}
