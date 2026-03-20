package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ClientServiceClient fetches client data from client-service via HTTP.
// proof-service uses this to enrich the proof snapshot with client details.
type ClientServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClientServiceClient returns a new ClientServiceClient.
// baseURL is read from CLIENT_SERVICE_URL env var, e.g. http://localhost:3002
func NewClientServiceClient(baseURL string) *ClientServiceClient {
	return &ClientServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ClientData holds the client fields needed for a proof snapshot.
type ClientData struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
}

// GetClient fetches client details by internal ID.
func (c *ClientServiceClient) GetClient(ctx context.Context, clientID string) (*ClientData, error) {
	url := fmt.Sprintf("%s/clients/%s", c.baseURL, clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("upstream.ClientServiceClient: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream.ClientServiceClient: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("upstream.ClientServiceClient: client %s not found", clientID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream.ClientServiceClient: unexpected status %d from client-service", resp.StatusCode)
	}

	var result ClientData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("upstream.ClientServiceClient: decode response: %w", err)
	}

	return &result, nil
}
