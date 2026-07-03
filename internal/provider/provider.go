package provider

type Tunnel struct {
	ID   string
	Name string
}

type TunnelRoute struct {
	Hostname string
	Port     int
}

type Provider interface {
	Validate() error
	GetTunnelByName(name string) (Tunnel, bool, error)
	CreateTunnel(name string) (Tunnel, error)
	DeleteTunnel(id string) error
	ListTunnels() ([]Tunnel, error)
	GetTunnelToken(tunnelID string) (string, error)
	ConfigureTunnel(tunnelID string, routes []TunnelRoute) error
	CreateDNSRecord(domain, tunnelID string) error
}
