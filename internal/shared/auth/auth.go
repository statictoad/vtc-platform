package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Verifier verifies Clerk JWTs using the JWKS endpoint.
// It caches the public keys and refreshes them automatically.
type Verifier struct {
	cache   *jwk.Cache
	jwksURL string
}

// NewVerifier creates a new JWT Verifier.
// It reads CLERK_JWKS_URL from environment variables and
// starts a background refresh of the public keys.
func NewVerifier(ctx context.Context) (*Verifier, error) {
	jwksURL := os.Getenv("CLERK_JWKS_URL")
	if jwksURL == "" {
		return nil, fmt.Errorf("auth: CLERK_JWKS_URL environment variable is required")
	}

	cache := jwk.NewCache(ctx)

	// Register the JWKS URL with a refresh interval.
	// Keys are refreshed every hour — Clerk rotates them infrequently.
	if err := cache.Register(jwksURL, jwk.WithRefreshInterval(time.Hour)); err != nil {
		return nil, fmt.Errorf("auth: failed to register JWKS URL: %w", err)
	}

	// Perform an initial fetch to fail fast if the URL is unreachable.
	if _, err := cache.Refresh(ctx, jwksURL); err != nil {
		return nil, fmt.Errorf("auth: failed to fetch JWKS: %w", err)
	}

	return &Verifier{
		cache:   cache,
		jwksURL: jwksURL,
	}, nil
}

// Claims holds the verified claims extracted from a Clerk JWT.
type Claims struct {
	// ClerkUserID is the Clerk user ID (sub claim).
	// Used to look up the internal client record via client-service.
	ClerkUserID string
}

// Verify parses and verifies a JWT string.
// Returns the extracted claims if the token is valid.
func (v *Verifier) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	keySet, err := v.cache.Get(ctx, v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to get JWKS: %w", err)
	}

	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid token: %w", err)
	}

	return &Claims{
		ClerkUserID: token.Subject(),
	}, nil
}
