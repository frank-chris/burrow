package tunnel

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/frank-chris/burrow/internal/install"
)

type QuickOptions struct {
	Password string
	TTL      time.Duration
}

type QuickTunnel struct {
	cmd      *exec.Cmd
	urlCh    chan string
	errCh    chan error
	listener net.Listener // auth proxy listener, if any
}

func StartQuickTunnel(port int, opts QuickOptions) (*QuickTunnel, error) {
	binPath, err := install.BinaryPath()
	if err != nil {
		return nil, err
	}

	targetPort := port
	var proxyListener net.Listener

	if opts.Password != "" {
		l, proxyPort, err := startAuthProxy(port, opts.Password)
		if err != nil {
			return nil, fmt.Errorf("could not start auth proxy: %w", err)
		}
		targetPort = proxyPort
		proxyListener = l
	}

	cmd := exec.Command(binPath, "tunnel", "--url", fmt.Sprintf("localhost:%d", targetPort))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		if proxyListener != nil {
			proxyListener.Close()
		}
		return nil, fmt.Errorf("could not attach to cloudflared output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		if proxyListener != nil {
			proxyListener.Close()
		}
		return nil, fmt.Errorf("could not start cloudflared: %w", err)
	}

	qt := &QuickTunnel{
		cmd:      cmd,
		urlCh:    make(chan string, 1),
		errCh:    make(chan error, 1),
		listener: proxyListener,
	}

	go qt.parseOutput(bufio.NewScanner(stderr))

	if opts.TTL > 0 {
		go func() {
			time.Sleep(opts.TTL)
			fmt.Println("\nTunnel TTL expired. Stopping.")
			qt.Stop()
			os.Exit(0)
		}()
	}

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
		return "", fmt.Errorf("timed out waiting for tunnel URL - is the local port reachable?")
	}
}

func (qt *QuickTunnel) Stop() error {
	if qt.listener != nil {
		qt.listener.Close()
	}
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

// startAuthProxy starts a local HTTP reverse proxy with basic auth in front of targetPort.
// Returns the listener and the port it is listening on.
func startAuthProxy(targetPort int, password string) (net.Listener, int, error) {
	target, err := url.Parse(fmt.Sprintf("http://localhost:%d", targetPort))
	if err != nil {
		return nil, 0, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, fmt.Errorf("could not bind auth proxy: %w", err)
	}
	proxyPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, p, ok := r.BasicAuth()
		if !ok || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="burrow"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	go http.Serve(listener, mux)
	return listener, proxyPort, nil
}
