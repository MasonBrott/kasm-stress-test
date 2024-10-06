package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"kasm-stress-test/internal/config"
)

// Client represents the API client for interacting with the Kasm API
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// post sends a POST request to the specified endpoint
func (c *Client) post(endpoint string, body interface{}) ([]byte, error) {
	// Ensure the APIHost ends with a slash if it doesn't already
	apiBase := strings.TrimSuffix(c.config.APIHost, "/") + "/"
	// Ensure the endpoint doesn't start with a slash
	endpoint = strings.TrimPrefix(endpoint, "/")

	url := apiBase + endpoint

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// apiRequest is a helper function to handle common API request structure
func (c *Client) apiRequest(endpoint string, additionalData map[string]interface{}) ([]byte, error) {
	requestBody := map[string]interface{}{
		"api_key":        c.config.APIKey,
		"api_key_secret": c.config.APISecret,
	}

	// Merge additionalData into requestBody
	for k, v := range additionalData {
		requestBody[k] = v
	}

	return c.post(endpoint, requestBody)
}
