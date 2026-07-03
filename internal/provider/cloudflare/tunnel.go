package cloudflare

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frank-chris/burrow/internal/provider"
)

type cfTunnel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetTunnelByName(name string) (provider.Tunnel, bool, error) {
	result, err := c.doRequest(http.MethodGet,
		fmt.Sprintf("/accounts/%s/cfd_tunnel?name=%s&is_deleted=false", c.accountID, name), nil)
	if err != nil {
		return provider.Tunnel{}, false, err
	}

	var tunnels []cfTunnel
	if err := json.Unmarshal(result, &tunnels); err != nil {
		return provider.Tunnel{}, false, fmt.Errorf("could not parse tunnel list response")
	}

	if len(tunnels) == 0 {
		return provider.Tunnel{}, false, nil
	}

	t := tunnels[0]
	return provider.Tunnel{ID: t.ID, Name: t.Name}, true, nil
}

func (c *Client) CreateTunnel(name string) (provider.Tunnel, error) {
	secret, err := generateSecret()
	if err != nil {
		return provider.Tunnel{}, fmt.Errorf("could not generate tunnel secret: %w", err)
	}

	result, err := c.doRequest(http.MethodPost,
		fmt.Sprintf("/accounts/%s/cfd_tunnel", c.accountID),
		map[string]string{
			"name":          name,
			"tunnel_secret": secret,
			"config_src":    "cloudflare",
		})
	if err != nil {
		return provider.Tunnel{}, err
	}

	var t cfTunnel
	if err := json.Unmarshal(result, &t); err != nil {
		return provider.Tunnel{}, fmt.Errorf("could not parse created tunnel response")
	}

	return provider.Tunnel{ID: t.ID, Name: t.Name}, nil
}

func (c *Client) DeleteTunnel(id string) error {
	_, err := c.doRequest(http.MethodDelete,
		fmt.Sprintf("/accounts/%s/cfd_tunnel/%s", c.accountID, id), nil)
	return err
}

func (c *Client) ListTunnels() ([]provider.Tunnel, error) {
	result, err := c.doRequest(http.MethodGet,
		fmt.Sprintf("/accounts/%s/cfd_tunnel?is_deleted=false", c.accountID), nil)
	if err != nil {
		return nil, err
	}

	var cfTunnels []cfTunnel
	if err := json.Unmarshal(result, &cfTunnels); err != nil {
		return nil, fmt.Errorf("could not parse tunnels response")
	}

	tunnels := make([]provider.Tunnel, len(cfTunnels))
	for i, t := range cfTunnels {
		tunnels[i] = provider.Tunnel{ID: t.ID, Name: t.Name}
	}
	return tunnels, nil
}

func (c *Client) GetTunnelToken(tunnelID string) (string, error) {
	result, err := c.doRequest(http.MethodGet,
		fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/token", c.accountID, tunnelID), nil)
	if err != nil {
		return "", err
	}

	var token string
	if err := json.Unmarshal(result, &token); err != nil {
		return "", fmt.Errorf("could not parse tunnel token response")
	}
	return token, nil
}

func (c *Client) ConfigureTunnel(tunnelID string, routes []provider.TunnelRoute) error {
	type ingressRule struct {
		Hostname string `json:"hostname,omitempty"`
		Service  string `json:"service"`
	}

	rules := make([]ingressRule, 0, len(routes)+1)
	for _, r := range routes {
		rules = append(rules, ingressRule{
			Hostname: r.Hostname,
			Service:  fmt.Sprintf("http://localhost:%d", r.Port),
		})
	}
	// Cloudflare requires a catch-all rule at the end
	rules = append(rules, ingressRule{Service: "http_status:404"})

	_, err := c.doRequest(http.MethodPut,
		fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", c.accountID, tunnelID),
		map[string]interface{}{
			"config": map[string]interface{}{
				"ingress": rules,
			},
		})
	return err
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

