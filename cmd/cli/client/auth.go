package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Login authenticates with the JujuDB server and saves the session cookie
func Login(serverURL, password string) error {
	serverURL = strings.TrimRight(serverURL, "/")
	c := NewWithCredentials(serverURL)

	// POST /login with form data
	form := url.Values{}
	form.Set("password", password)

	req, err := http.NewRequest("POST", serverURL+"/login", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if login was successful (redirect to /dashboard)
	if resp.StatusCode == http.StatusSeeOther {
		loc := resp.Header.Get("Location")
		if loc == "/login?error=1" || strings.Contains(loc, "error") {
			return fmt.Errorf("authentication failed: invalid password")
		}
	} else if resp.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	// Extract session cookie
	var sessionCookie string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "jujudb-session" {
			sessionCookie = cookie.Name + "=" + cookie.Value
			break
		}
	}

	if sessionCookie == "" {
		return fmt.Errorf("no session cookie received from server")
	}

	// Save config and session
	if err := SaveConfig(serverURL); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if err := SaveSession(sessionCookie); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}
