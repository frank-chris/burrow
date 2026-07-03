package tui

import "fmt"

func PrintStatus(tunnels []string) {
	for _, t := range tunnels {
		fmt.Printf("  ● %s\n", t)
	}
}
