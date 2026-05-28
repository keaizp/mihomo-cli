package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client communicates with the mihomo REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an API client for the given mihomo API address.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Proxy represents a single proxy node returned by the mihomo API.
type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now,omitempty"`
	All     []string `json:"all,omitempty"`
	History []struct {
		Time  time.Time `json:"time"`
		Delay int       `json:"delay"`
	} `json:"history,omitempty"`
}

// ProxyGroup is a top-level proxy group from GET /proxies.
type ProxyGroup struct {
	Name string   `json:"name"`
	Type string   `json:"type"`
	Now  string   `json:"now"`
	All  []string `json:"all"`
}

// ProxiesResponse is the full response from GET /proxies.
type ProxiesResponse struct {
	Proxies map[string]Proxy `json:"proxies"`
}

// Connection represents an active connection.
type Connection struct {
	ID       string `json:"id"`
	Metadata struct {
		Network string `json:"network"`
		Host    string `json:"host"`
	} `json:"metadata"`
	Upload   int64    `json:"upload"`
	Download int64    `json:"download"`
	Start    string   `json:"start"`
	Chains   []string `json:"chains"`
	Rule     string   `json:"rule"`
}

// ConnectionsResponse is the full response from GET /connections.
type ConnectionsResponse struct {
	Connections []Connection `json:"connections"`
}

// DelayResult is the result of a single latency test.
type DelayResult struct {
	Delay int `json:"delay"`
}

// GetProxies fetches all proxy groups and nodes.
func (c *Client) GetProxies() (*ProxiesResponse, error) {
	var resp ProxiesResponse
	if err := c.do("GET", "/proxies", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SwitchProxy selects a specific node within a proxy group.
func (c *Client) SwitchProxy(group, node string) error {
	body := fmt.Sprintf(`{"name":"%s"}`, node)
	return c.do("PUT", "/proxies/"+url.PathEscape(group), strings.NewReader(body), nil)
}

// TestDelay tests the latency of a specific proxy node.
func (c *Client) TestDelay(name string, timeout time.Duration) (int, error) {
	u := fmt.Sprintf("%s/proxies/%s/delay?timeout=%d&url=https://www.gstatic.com/generate_204",
		c.baseURL, url.PathEscape(name), int(timeout.Milliseconds()))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result DelayResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.Delay == 0 {
		return 0, fmt.Errorf("timeout")
	}
	return result.Delay, nil
}

// GetConnections fetches all active connections.
func (c *Client) GetConnections() (*ConnectionsResponse, error) {
	var resp ConnectionsResponse
	if err := c.do("GET", "/connections", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CloseConnection closes a single connection by ID.
func (c *Client) CloseConnection(id string) error {
	return c.do("DELETE", "/connections/"+url.PathEscape(id), nil, nil)
}

// ReloadConfig triggers mihomo to reload its configuration.
func (c *Client) ReloadConfig() error {
	body := strings.NewReader(`{"path":""}`)
	return c.do("PUT", "/configs", body, nil)
}

// HealthCheck returns nil if the mihomo API is reachable.
func (c *Client) HealthCheck() error {
	var v any
	return c.do("GET", "/version", nil, &v)
}

func (c *Client) do(method, path string, body io.Reader, result any) error {
	u := c.baseURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error %d: %s", resp.StatusCode, string(b))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
