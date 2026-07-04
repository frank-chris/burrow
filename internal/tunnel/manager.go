package tunnel

import (
	"fmt"

	"github.com/frank-chris/burrow/internal/config"
	"github.com/frank-chris/burrow/internal/provider"
	"github.com/frank-chris/burrow/internal/state"
)

type Manager struct {
	provider  provider.Provider
	processes []*Process
}

func NewManager(p provider.Provider) *Manager {
	return &Manager{provider: p}
}

func (m *Manager) StartAll(cfg *config.Config) error {
	for _, t := range cfg.Tunnels {
		if err := m.start(t); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) start(t config.TunnelConfig) error {
	tun, exists, err := m.provider.GetTunnelByName(t.Name)
	if err != nil {
		return fmt.Errorf("could not check tunnel %q: %w", t.Name, err)
	}
	if !exists {
		fmt.Printf("  Creating tunnel %q...\n", t.Name)
		tun, err = m.provider.CreateTunnel(t.Name)
		if err != nil {
			return fmt.Errorf("could not create tunnel %q: %w", t.Name, err)
		}
	}

	if err := m.provider.ConfigureTunnel(tun.ID, []provider.TunnelRoute{
		{Hostname: t.Domain, Port: t.Port},
	}); err != nil {
		return fmt.Errorf("could not configure tunnel %q: %w", t.Name, err)
	}

	if err := m.provider.CreateDNSRecord(t.Domain, tun.ID); err != nil {
		return fmt.Errorf("could not set up DNS for %q: %w", t.Domain, err)
	}

	token, err := m.provider.GetTunnelToken(tun.ID)
	if err != nil {
		return fmt.Errorf("could not get token for tunnel %q: %w", t.Name, err)
	}

	WarnIfPortClosed(t.Port)
	proc, err := StartProcess(t.Name, token)
	if err != nil {
		return err
	}
	m.processes = append(m.processes, proc)
	fmt.Printf("  [up] %s -> https://%s\n", t.Name, t.Domain)
	return nil
}

func (m *Manager) TunnelProcesses() []state.TunnelProcess {
	result := make([]state.TunnelProcess, 0, len(m.processes))
	for _, p := range m.processes {
		result = append(result, state.TunnelProcess{Name: p.name, PID: p.PID()})
	}
	return result
}

func (m *Manager) StopAll() {
	for _, p := range m.processes {
		p.Stop()
	}
}
