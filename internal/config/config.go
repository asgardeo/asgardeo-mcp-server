package config

import (
	"fmt"
	"log"
	"os"
)

// Load loads required Asgardeo environment variables and validates them.
func Load() (baseURL, clientID, clientSecret string, certPath *string, err error) {
	baseURL = getBaseURL()
	clientID = getClientID()
	clientSecret = getClientSecret()
	certPath = nil
	if os.Getenv(CERTIFICATE_PATH_PARAM) != "" {
		certPathValue := os.Getenv(CERTIFICATE_PATH_PARAM)
		certPath = &certPathValue
	}
	log.Printf("Env loaded: BASE_URL=%q, CLIENT_ID=%q", baseURL, clientID)
	if baseURL == "" || clientID == "" || clientSecret == "" {
		err = fmt.Errorf("missing required environment variables BASE_URL, CLIENT_ID, or CLIENT_SECRET")
	}
	return
}

// GetTransportMode returns the transport mode (stdio or sse). Defaults to stdio.
func GetTransportMode() string {
	mode := os.Getenv(TRANSPORT_MODE_PARAM)
	if mode == TransportSSE {
		return TransportSSE
	}
	return TransportStdio
}

// GetServerPort returns the HTTP port for SSE mode. Defaults to 8080.
func GetServerPort() string {
	port := os.Getenv(SERVER_PORT_PARAM)
	if port == "" {
		return DefaultPort
	}
	return port
}

// GetServerBaseURL returns the base URL for SSE mode.
func GetServerBaseURL() string {
	return os.Getenv(SERVER_BASE_URL_PARAM)
}

// GetMCPAudience returns the expected audience for validated tokens.
// When set, the middleware enforces that the token's aud claim contains this value.
func GetMCPAudience() string {
	return os.Getenv(MCP_AUDIENCE_PARAM)
}

// GetIntrospectClientID returns the client ID used to authenticate introspection calls.
// Uses MCP_INTROSPECT_CLIENT_ID if set, otherwise falls back to CLIENT_ID.
func GetIntrospectClientID() string {
	if v := os.Getenv(MCP_INTROSPECT_CLIENT_ID_PARAM); v != "" {
		return v
	}
	return getClientID()
}

// GetIntrospectClientSecret returns the client secret used to authenticate introspection calls.
// Uses MCP_INTROSPECT_CLIENT_SECRET if set, otherwise falls back to CLIENT_SECRET.
func GetIntrospectClientSecret() string {
	if v := os.Getenv(MCP_INTROSPECT_CLIENT_SECRET_PARAM); v != "" {
		return v
	}
	return getClientSecret()
}

// GetSkipTLSVerify returns true when SKIP_TLS_VERIFY=true is set.
// For local development with self-signed certs only — never use in production.
func GetSkipTLSVerify() bool {
	return os.Getenv(SKIP_TLS_VERIFY_PARAM) == "true"
}

// GetBaseURL returns the Asgardeo base URL.
func GetBaseURL() string {
	return getBaseURL()
}

// GetMCPClientID returns the client ID of the user-facing MCP application.
// Falls back to MCP_INTROSPECT_CLIENT_ID → CLIENT_ID.
func GetMCPClientID() string {
	if v := os.Getenv(MCP_CLIENT_ID_PARAM); v != "" {
		return v
	}
	return GetIntrospectClientID()
}

// GetMCPClientSecret is the symmetric counterpart to GetMCPClientID.
// Falls back to MCP_INTROSPECT_CLIENT_SECRET → CLIENT_SECRET.
func GetMCPClientSecret() string {
	return GetIntrospectClientSecret()
}

// GetCertPath returns the certificate path if configured.
func GetCertPath() *string {
	if v := os.Getenv(CERTIFICATE_PATH_PARAM); v != "" {
		return &v
	}
	return nil
}

func GetProductName() string {
	productMode := os.Getenv(PRODUCT_MODE_PARAM)
	if productMode == ProductModes.WSO2IS {
		return ProductNames.WSO2IS
	}
	return ProductNames.Asgardeo
}

func getBaseURL() string {
	baseURL := os.Getenv(BASE_URL_PARAM)
	if baseURL == "" {
		// Fallback to ASGARDEO_BASE_URL for backward compatibility
		baseURL = os.Getenv(ASGARDEO_BASE_URL_PARAM)
	}
	return baseURL
}

func getClientID() string {
	clientID := os.Getenv(CLIENT_ID_PARAM)
	if clientID == "" {
		// Fallback to ASGARDEO_CLIENT_ID for backward compatibility
		clientID = os.Getenv(ASGARDEO_CLIENT_ID_PARAM)
	}
	return clientID
}

func getClientSecret() string {
	clientSecret := os.Getenv(CLIENT_SECRET_PARAM)
	if clientSecret == "" {
		// Fallback to ASGARDEO_CLIENT_SECRET for backward compatibility
		clientSecret = os.Getenv(ASGARDEO_CLIENT_SECRET_PARAM)
	}
	return clientSecret
}
