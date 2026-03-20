package cache

import (
	"fmt"
	"os"

	"github.com/valkey-io/valkey-go"
)

// NewClient creates a new Valkey client.
// It reads VALKEY_URL from environment variables.
//
// Example URL: valkey://localhost:6379
func NewClient() (valkey.Client, error) {
	url := os.Getenv("VALKEY_URL")
	if url == "" {
		return nil, fmt.Errorf("VALKEY_URL environment variable is required")
	}

	opts, err := valkey.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("cache: failed to parse VALKEY_URL: %w", err)
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("cache: failed to create client: %w", err)
	}

	return client, nil
}
