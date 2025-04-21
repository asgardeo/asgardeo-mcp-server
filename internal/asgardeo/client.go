package asgardeo

import (
	"context"

	"github.com/shashimalcse/go-asgardeo/management"
)

// NewClient initializes an Asgardeo management client with client credentials.
func NewClient(ctx context.Context, baseURL, clientID, clientSecret string) (*management.Client, error) {
	return management.New(
		baseURL,
		management.WithClientCredentials(ctx, clientID, clientSecret),
	)
}
