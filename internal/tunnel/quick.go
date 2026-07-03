package tunnel

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/frank-chris/burrow/internal/install"
)

type QuickTunnel struct {
	cmd   *exec.Cmd
	urlCh chan string
	errCh chan error
}

func StartQuickTunnel(port int) (*QuickTunnel, error) {
	binPath, err := install.BinaryPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binPath, "tunnel", "--url", fmt.Sprintf("localhost:%d", port))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("could not attach to cloudflared output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start cloudflared: %w", err)
	}

	qt := &QuickTunnel{
		cmd:   cmd,
		urlCh: make(chan string, 1),
		errCh: make(chan error, 1),
	}

	go qt.parseOutput(bufio.NewScanner(stderr))
	return qt, nil
}

// WaitForURL blocks until cloudflared prints its public URL or the timeout elapses.
func (qt *QuickTunnel) WaitForURL(timeout time.Duration) (string, error) {
	select {
	case url := <-qt.urlCh:
		return url, nil
	case err := <-qt.errCh:
		return "", err
	case <-time.After(timeout):
		qt.Stop()
		return "", fmt.Errorf("timed out waiting for tunnel URL — is the local port reachable?")
	}
}

func (qt *QuickTunnel) Stop() error {
	if qt.cmd == nil || qt.cmd.Process == nil {
		return nil
	}
	return qt.cmd.Process.Kill()
}

func (qt *QuickTunnel) parseOutput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if url := extractURL(line); url != "" && strings.Contains(url, "trycloudflare.com") {
			qt.urlCh <- url
			return
		}
	}
	qt.errCh <- fmt.Errorf("cloudflared exited before providing a URL")
}

// extractURL finds the first https:// URL in a log line.
func extractURL(line string) string {
	idx := strings.Index(line, "https://")
	if idx == -1 {
		return ""
	}
	rest := line[idx:]
	end := strings.IndexAny(rest, " \t|")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
