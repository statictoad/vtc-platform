package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FleetServiceClient fetches vehicle data from fleet-service via HTTP.
// proof-service uses this to enrich the proof snapshot with the number plate.
type FleetServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFleetServiceClient returns a new FleetServiceClient.
// baseURL is read from FLEET_SERVICE_URL env var, e.g. http://localhost:3003
func NewFleetServiceClient(baseURL string) *FleetServiceClient {
	return &FleetServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// VehicleData holds the vehicle fields needed for a proof snapshot.
type VehicleData struct {
	NumberPlate string `json:"number_plate"`
}

// GetVehicle fetches vehicle details by internal ID.
func (c *FleetServiceClient) GetVehicle(ctx context.Context, vehicleID string) (*VehicleData, error) {
	url := fmt.Sprintf("%s/vehicles/%s", c.baseURL, vehicleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("upstream.FleetServiceClient: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream.FleetServiceClient: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("upstream.FleetServiceClient: vehicle %s not found", vehicleID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream.FleetServiceClient: unexpected status %d from fleet-service", resp.StatusCode)
	}

	var result VehicleData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("upstream.FleetServiceClient: decode response: %w", err)
	}

	return &result, nil
}
