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
