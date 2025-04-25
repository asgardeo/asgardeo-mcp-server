package asgardeo

import (
	"context"
	"sync"

	"github.com/asgardeo/go/management"
	"github.com/asgardeo/mcp/internal/config"
)

var (
	clientInstance *management.Client
	once           sync.Once
	initErr        error
)

// NewClient initializes an Asgardeo management client with client credentials.
func NewClient(ctx context.Context, baseURL, clientID, clientSecret string) (*management.Client, error) {
	return management.New(
		baseURL,
		management.WithClientCredentials(ctx, clientID, clientSecret),
	)
}

// GetClient returns the singleton Asgardeo client.
func GetClientInstance(ctx context.Context) (*management.Client, error) {
	once.Do(func() {
		baseURL, clientID, clientSecret, err := config.Load()
		if err != nil {
			initErr = err
			return
		}

		clientInstance, initErr = NewClient(ctx, baseURL, clientID, clientSecret)
	})

	return clientInstance, initErr
}
