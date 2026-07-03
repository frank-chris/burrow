package tunnel

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/frank-chris/burrow/internal/install"
	"github.com/frank-chris/burrow/internal/state"
)

type Process struct {
	name    string
	cmd     *exec.Cmd
	stderr  *bytes.Buffer
	logFile *os.File
}

func CheckCloudflared() error {
	if !install.IsInstalled() {
		return fmt.Errorf("cloudflared is not installed — run `burrow init` first")
	}
	return nil
}

func StartProcess(name, token string) (*Process, error) {
	binPath, err := install.BinaryPath()
	if err != nil {
		return nil, err
	}

	logPath, err := state.LogPath(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("could not open log file for %q: %w", name, err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(binPath, "tunnel", "run", "--token", token)
	cmd.Stderr = io.MultiWriter(&stderr, logFile)

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("could not start cloudflared for %q: %w", name, err)
	}

	p := &Process{name: name, cmd: cmd, stderr: &stderr, logFile: logFile}
	go p.watch()
	return p, nil
}

func (p *Process) watch() {
	defer p.logFile.Close()
	if err := p.cmd.Wait(); err != nil {
		msg := p.stderr.String()
		if msg != "" {
			fmt.Fprintf(os.Stderr, "\ntunnel %q stopped: %s\n", p.name, msg)
		} else {
			fmt.Fprintf(os.Stderr, "\ntunnel %q stopped unexpectedly\n", p.name)
		}
	}
}

func (p *Process) PID() int {
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *Process) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}
