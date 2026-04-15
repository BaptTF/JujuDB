package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Client is the HTTP client for JujuDB API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	cookie     string
}

// configDir returns the path to ~/.config/jujudb/
func configDir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot find home directory: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "jujudb"), nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

// SaveConfig saves the server URL to config file
func SaveConfig(serverURL string) error {
	dir, err := ensureConfigDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config"), []byte(serverURL), 0600)
}

// LoadConfig loads the server URL from config file
func LoadConfig() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "config"))
	if err != nil {
		return "", fmt.Errorf("not configured. Run 'jujudb login --server URL --password PASSWORD' first")
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveSession saves the session cookie to disk
func SaveSession(cookie string) error {
	dir, err := ensureConfigDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "session"), []byte(cookie), 0600)
}

// LoadSession loads the session cookie from disk
func LoadSession() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "session"))
	if err != nil {
		return "", fmt.Errorf("not authenticated. Run 'jujudb login --server URL --password PASSWORD' first")
	}
	return strings.TrimSpace(string(data)), nil
}

// New creates a new Client from saved config and session
func New() (*Client, error) {
	baseURL, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cookie, err := LoadSession()
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects — we need to detect auth redirects
				return http.ErrUseLastResponse
			},
		},
		cookie: cookie,
	}, nil
}

// NewWithCredentials creates a new Client for login (no session yet)
func NewWithCredentials(serverURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(serverURL, "/"),
		HTTPClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// doRequest executes an HTTP request with auth cookie attached
func (c *Client) doRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	u := c.BaseURL + path

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Detect auth redirect (server redirects to /login or / when not authenticated)
	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusFound {
		loc := resp.Header.Get("Location")
		if loc == "/" || loc == "/login" || strings.HasSuffix(loc, "/login") {
			resp.Body.Close()
			return nil, fmt.Errorf("session expired. Run 'jujudb login --server %s --password PASSWORD' to re-authenticate", c.BaseURL)
		}
	}

	return resp, nil
}

// Get performs a GET request to the API
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

// Post performs a POST request with JSON body
func (c *Client) Post(path string, jsonBody []byte) ([]byte, error) {
	var body io.Reader
	ct := "application/json"
	if jsonBody != nil {
		body = strings.NewReader(string(jsonBody))
	}
	resp, err := c.doRequest("POST", path, body, ct)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

// Put performs a PUT request with JSON body
func (c *Client) Put(path string, jsonBody []byte) ([]byte, error) {
	resp, err := c.doRequest("PUT", path, strings.NewReader(string(jsonBody)), "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

// Delete performs a DELETE request
func (c *Client) Delete(path string, query url.Values) error {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		// Try to parse as JSON (dependency conflict)
		var conflict map[string]interface{}
		if json.Unmarshal(data, &conflict) == nil {
			if code, ok := conflict["code"].(string); ok && code == "HAS_DEPENDENCIES" {
				msg := fmt.Sprintf("Conflict: %s", conflict["message"])
				if items, ok := conflict["related_items"].([]interface{}); ok && len(items) > 0 {
					msg += fmt.Sprintf("\n  Related items: %d", len(items))
				}
				if subs, ok := conflict["related_sublocations"].([]interface{}); ok && len(subs) > 0 {
					msg += fmt.Sprintf("\n  Related sub-locations: %d", len(subs))
				}
				msg += "\n  Use --force to force deletion"
				return fmt.Errorf("%s", msg)
			}
		}
		return fmt.Errorf("error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return nil
}

// PostNoContent performs a POST request that expects no JSON body response (like sync)
func (c *Client) PostNoContent(path string) ([]byte, error) {
	resp, err := c.doRequest("POST", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}
