package asgardeo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
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

// newHTTPClient returns an HTTP client with optional TLS verification skip.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: internal_config.GetSkipTLSVerify(),
			},
		},
	}
}

// NewClient initializes an Asgardeo management client with client credentials.
func NewClient(ctx context.Context, baseURL, clientID, clientSecret string, certPath *string) (*sdk.Client, error) {

	cfg := config.DefaultClientConfig().
		WithBaseURL(baseURL).
		WithTimeout(10*time.Second).
		WithClientCredentials(clientID, clientSecret).
		WithCertificatePath(certPath)

	return sdk.New(cfg)
}

// GetClientInstance returns the singleton M2M Asgardeo client.
func GetClientInstance(ctx context.Context) (*sdk.Client, error) {
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

// NewClientWithToken creates an Asgardeo client authenticated with a user-supplied token.
// Use this for the user token delegation path (SSE mode with authenticated users).
func NewClientWithToken(baseURL, token string, certPath *string) (*sdk.Client, error) {
	cfg := config.DefaultClientConfig().
		WithBaseURL(baseURL).
		WithHTTPClient(newHTTPClient()).
		WithToken(token).
		WithCertificatePath(certPath)

	return sdk.New(cfg)
}

// GetClientForContext returns an Asgardeo client authenticated with the user's Bearer token
// injected by the auth middleware (SSE mode). Returns an error if no token is in the context.
// Base URL is resolved from the request context (X-Base-URL header), falling back to BASE_URL env var.
func GetClientForContext(ctx context.Context) (*sdk.Client, error) {
	token, ok := auth.GetRawToken(ctx)
	if !ok {
		return nil, fmt.Errorf("no user token in request context: ensure the request includes a valid Bearer token")
	}
	baseURL, ok := auth.GetBaseURL(ctx)
	if !ok || baseURL == "" {
		baseURL = internal_config.GetBaseURL()
	}
	return NewClientWithToken(baseURL, token, internal_config.GetCertPath())
}
