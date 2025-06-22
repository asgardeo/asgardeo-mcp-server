package config

import (
	"fmt"
	"log"
	"os"
)

// Load loads required Asgardeo environment variables and validates them.
func Load() (baseURL, clientID, clientSecret string, certPath *string, err error) {
	baseURL = os.Getenv("ASGARDEO_BASE_URL")
	clientID = os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret = os.Getenv("ASGARDEO_CLIENT_SECRET")
	certPath = nil
	if os.Getenv("CERT_PATH") != "" {
		certPathValue := os.Getenv("CERT_PATH")
		certPath = &certPathValue
	}
	log.Printf("Env loaded: ASGARDEO_BASE_URL=%q, CLIENT_ID=%q", baseURL, clientID)
	if baseURL == "" || clientID == "" || clientSecret == "" {
		err = fmt.Errorf("missing required environment variables ASGARDEO_BASE_URL, ASGARDEO_CLIENT_ID, or ASGARDEO_CLIENT_SECRET")
	}
	return
}

func GetProductName() string {
	productMode := os.Getenv("PRODUCT_MODE")
	if productMode == ProductModes.WSO2IS {
		return ProductNames.WSO2IS
	}
	return ProductNames.Asgardeo
}
