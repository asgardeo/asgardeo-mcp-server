package asgardeo

import (
	"context"
	"sync"
	"time"

	"github.com/asgardeo/go/pkg/config"
	"github.com/asgardeo/go/pkg/sdk"
	"github.com/asgardeo/mcp/internal/auth"
	internal_config "github.com/asgardeo/mcp/internal/config"
)

var (
	clientInstance *sdk.Client
	once           sync.Once
	initErr        error
)

// NewClient initializes an Asgardeo management client with client credentials.
func NewClient(ctx context.Context, baseURL, clientID, clientSecret string, certPath *string) (*sdk.Client, error) {

	cfg := config.DefaultClientConfig().
		WithBaseURL(baseURL).
		WithTimeout(10*time.Second).
		WithClientCredentials(clientID, clientSecret).
		WithCertificatePath(certPath)

	return sdk.New(cfg)
}

func NewTokenClient(ctx context.Context, baseURL, token string, certPath *string) (*sdk.Client, error) {
	cfg := config.DefaultClientConfig().
		WithBaseURL(baseURL).
		WithTimeout(10 * time.Second).
		WithToken(token).
		WithCertificatePath(certPath)

	return sdk.New(cfg)
}

// GetClient returns the singleton Asgardeo client.
func GetClientInstance(ctx context.Context) (*sdk.Client, error) {
	if token, ok := auth.AccessTokenFromContext(ctx); ok {
		runtimeConfig, err := internal_config.LoadRuntime()
		if err != nil {
			return nil, err
		}
		baseURL := runtimeConfig.BaseURL
		if tenantBaseURL, ok := auth.TenantBaseURLFromContext(ctx); ok {
			baseURL = tenantBaseURL
		}
		return NewTokenClient(ctx, baseURL, token, runtimeConfig.CertPath)
	}

	once.Do(func() {
		baseURL, clientID, clientSecret, certPath, err := internal_config.Load()
		if err != nil {
			initErr = err
			return
		}

		clientInstance, initErr = NewClient(ctx, baseURL, clientID, clientSecret, certPath)
	})

	return clientInstance, initErr
}
