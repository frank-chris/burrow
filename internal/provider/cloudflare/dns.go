package cloudflare

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type cfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cfDNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (c *Client) CreateDNSRecord(domain, tunnelID string) error {
	zoneID, err := c.getZoneForDomain(domain)
	if err != nil {
		return err
	}

	target := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)

	existing, err := c.getDNSRecord(zoneID, domain)
	if err != nil {
		return err
	}

	if existing != nil {
		if existing.Content == target {
			return nil // already correct, nothing to do
		}
		_, err = c.doRequest(http.MethodPut,
			fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, existing.ID),
			map[string]interface{}{
				"type":    "CNAME",
				"name":    domain,
				"content": target,
				"proxied": true,
			})
		return err
	}

	_, err = c.doRequest(http.MethodPost,
		fmt.Sprintf("/zones/%s/dns_records", zoneID),
		map[string]interface{}{
			"type":    "CNAME",
			"name":    domain,
			"content": target,
			"proxied": true,
		})
	if err != nil {
		var cfErr *cfAPIError
		if errors.As(err, &cfErr) && cfErr.hasCode(81057) {
			return nil // record already exists
		}
		return err
	}
	return nil
}

func (c *Client) getDNSRecord(zoneID, name string) (*cfDNSRecord, error) {
	result, err := c.doRequest(http.MethodGet,
		fmt.Sprintf("/zones/%s/dns_records?name=%s&type=CNAME", zoneID, name), nil)
	if err != nil {
		return nil, err
	}

	var records []cfDNSRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, fmt.Errorf("could not parse DNS records response")
	}

	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

func (c *Client) getZoneForDomain(domain string) (string, error) {
	parts := strings.Split(domain, ".")
	for i := range parts {
		candidate := strings.Join(parts[i:], ".")
		if !strings.Contains(candidate, ".") {
			continue
		}
		result, err := c.doRequest(http.MethodGet,
			fmt.Sprintf("/zones?name=%s&status=active", candidate), nil)
		if err != nil {
			continue
		}
		var zones []cfZone
		if json.Unmarshal(result, &zones) == nil && len(zones) > 0 {
			return zones[0].ID, nil
		}
	}
	return "", fmt.Errorf("no Cloudflare zone found for %q - ensure the domain is added to your Cloudflare account", domain)
}
