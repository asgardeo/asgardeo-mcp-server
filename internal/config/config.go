package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type RuntimeConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	CertPath     *string
}

type HTTPConfig struct {
	Addr          string
	Endpoint      string
	ServerURL     string
	ClientID      string
	Scopes        []string
	TokenAudience string
	TokenIssuer   string
}

// Load loads required Asgardeo environment variables and validates them.
func Load() (baseURL, clientID, clientSecret string, certPath *string, err error) {
	baseURL = TenantBaseURL(getBaseURL(), getOrgHandle())
	clientID = getClientID()
	clientSecret = getClientSecret()
	certPath = getCertPath()
	log.Printf("Env loaded: BASE_URL=%q, CLIENT_ID=%q", baseURL, clientID)
	if baseURL == "" || clientID == "" || clientSecret == "" {
		err = fmt.Errorf("missing required environment variables BASE_URL, CLIENT_ID, or CLIENT_SECRET")
	}
	return
}

func LoadRuntime() (RuntimeConfig, error) {
	baseURL := TenantBaseURL(getBaseURL(), getOrgHandle())
	if baseURL == "" {
		return RuntimeConfig{}, fmt.Errorf("missing required environment variable BASE_URL")
	}
	return RuntimeConfig{
		BaseURL:      baseURL,
		ClientID:     getClientID(),
		ClientSecret: getClientSecret(),
		CertPath:     getCertPath(),
	}, nil
}

func LoadHTTP() (HTTPConfig, error) {
	baseURL := getBaseURL()
	if baseURL == "" {
		return HTTPConfig{}, fmt.Errorf("missing required environment variable BASE_URL")
	}

	endpoint := os.Getenv(HTTP_ENDPOINT_PARAM)
	if endpoint == "" {
		endpoint = DefaultHTTPEndpoint
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	serverURL := strings.TrimRight(os.Getenv(MCP_SERVER_URL_PARAM), "/")
	if serverURL == "" {
		serverURL = "http://localhost" + defaultPort(os.Getenv(HTTP_ADDR_PARAM)) + TenantPathPrefix + "<org>" + endpoint
	}

	clientID := os.Getenv(MCP_CLIENT_ID_PARAM)
	if clientID == "" {
		clientID = DefaultMCPClientID
	}

	return HTTPConfig{
		Addr:          getEnvOrDefault(HTTP_ADDR_PARAM, DefaultHTTPAddr),
		Endpoint:      endpoint,
		ServerURL:     serverURL,
		ClientID:      clientID,
		Scopes:        splitScopes(os.Getenv(MCP_SCOPES_PARAM)),
		TokenAudience: getEnvOrDefault(TOKEN_AUDIENCE_PARAM, serverURL),
		TokenIssuer:   getEnvOrDefault(TOKEN_ISSUER_PARAM, strings.TrimRight(baseURL, "/")+TenantPathPrefix+"<org>/oauth2/token"),
	}, nil
}

func GetTransport() string {
	return strings.ToLower(getEnvOrDefault(TRANSPORT_PARAM, DefaultTransport))
}

func TenantBaseURL(baseURL, orgHandle string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	orgHandle = strings.Trim(orgHandle, "/")
	if baseURL == "" || orgHandle == "" || strings.Contains(baseURL, TenantPathPrefix) {
		return baseURL
	}
	return baseURL + TenantPathPrefix + orgHandle
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

func getOrgHandle() string {
	return os.Getenv(ORG_HANDLE_PARAM)
}

func getCertPath() *string {
	if os.Getenv(CERTIFICATE_PATH_PARAM) == "" {
		return nil
	}
	certPathValue := os.Getenv(CERTIFICATE_PATH_PARAM)
	return &certPathValue
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func splitScopes(scopes string) []string {
	if strings.TrimSpace(scopes) == "" {
		return nil
	}
	return strings.Fields(scopes)
}

func defaultPort(addr string) string {
	if addr == "" {
		addr = DefaultHTTPAddr
	}
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	if index := strings.LastIndex(addr, ":"); index >= 0 {
		return addr[index:]
	}
	return DefaultHTTPAddr
}
