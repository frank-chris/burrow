package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/frank-chris/burrow/internal/constants"
	"github.com/frank-chris/burrow/internal/provider"
)

// compile-time check that Client implements Provider
var _ provider.Provider = (*Client)(nil)

type Client struct {
	apiToken  string
	accountID string
	http      *http.Client
}

func New(apiToken, accountID string) *Client {
	return &Client{
		apiToken:  apiToken,
		accountID: accountID,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

type cfResponse struct {
	Success bool            `json:"success"`
	Errors  []cfError       `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfAPIError struct {
	errors []cfError
}

func (e *cfAPIError) Error() string {
	if len(e.errors) > 0 {
		return fmt.Sprintf("Cloudflare API error: %s", e.errors[0].Message)
	}
	return "Cloudflare API error"
}

func (e *cfAPIError) hasCode(code int) bool {
	for _, err := range e.errors {
		if err.Code == code {
			return true
		}
	}
	return false
}

func (c *Client) doRequest(method, path string, body interface{}) (json.RawMessage, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, constants.CloudflareAPIBase+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Cloudflare API: %w", err)
	}
	defer resp.Body.Close()

	var result cfResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unexpected response from Cloudflare API")
	}

	if !result.Success {
		return nil, &cfAPIError{errors: result.Errors}
	}

	return result.Result, nil
}

func (c *Client) Validate() error {
	_, err := c.doRequest(http.MethodGet, constants.CloudflareVerifyPath, nil)
	if err != nil {
		return fmt.Errorf("invalid API token: %w", err)
	}
	return nil
}
