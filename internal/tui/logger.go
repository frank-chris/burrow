package tui

import "fmt"

func LogRequest(domain, method, path, status string) {
	fmt.Printf("[%s] %s %s %s\n", domain, method, path, status)
}
