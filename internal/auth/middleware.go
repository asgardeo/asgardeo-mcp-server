/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	internal_config "github.com/asgardeo/mcp/internal/config"
)

type contextKey string

const (
	userInfoKey contextKey = "userInfo"
	rawTokenKey contextKey = "rawToken"
	baseURLKey  contextKey = "baseURL"
)

// UserInfo holds the validated token information extracted from introspection.
type UserInfo struct {
	Sub      string
	Username string
	ClientID string
	Scopes   []string
}

// audience is a custom type that unmarshals either a JSON string or array of strings.
type audience []string

func (a *audience) UnmarshalJSON(data []byte) error {
	// Try array first.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*a = arr
		return nil
	}
	// Fall back to single string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = []string{s}
	return nil
}

// IntrospectionResponse represents the Asgardeo token introspection response (RFC 7662).
type IntrospectionResponse struct {
	Active   bool     `json:"active"`
	Sub      string   `json:"sub"`
	Username string   `json:"username"`
	ClientID string   `json:"client_id"`
	Scope    string   `json:"scope"`
	Exp      int64    `json:"exp"`
	Iat      int64    `json:"iat"`
	Iss      string   `json:"iss"`
	Aud      audience `json:"aud"`
}

// TokenValidator validates bearer tokens via Asgardeo introspection endpoint.
type TokenValidator struct {
	introspectionURL string
	clientID         string
	clientSecret     string
	expectedAudience string // optional — enforced when non-empty
	httpClient       *http.Client
}

// NewTokenValidator creates a new TokenValidator.
// expectedAudience should be the client ID of the application that issued tokens
// (e.g. the device code app client ID). Pass an empty string to skip aud validation.
// skipTLSVerify disables certificate verification — for local dev with self-signed certs only.
func NewTokenValidator(baseURL, clientID, clientSecret, expectedAudience string, skipTLSVerify bool) *TokenValidator {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	return &TokenValidator{
		introspectionURL: fmt.Sprintf("%s/oauth2/introspect", baseURL),
		clientID:         clientID,
		clientSecret:     clientSecret,
		expectedAudience: expectedAudience,
		httpClient:       &http.Client{Timeout: 10 * time.Second, Transport: transport},
	}
}

// Validate introspects a token and returns UserInfo if all claims are valid.
// Checks: active, exp, aud (when expectedAudience is configured).
func (v *TokenValidator) Validate(ctx context.Context, token string) (*UserInfo, error) {
	data := url.Values{}
	data.Set("token", token)
	data.Set("token_type_hint", "access_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(v.clientID, v.clientSecret)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection endpoint returned %d", resp.StatusCode)
	}

	var result IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	// Check active — covers revoked tokens and basic validity.
	if !result.Active {
		return nil, fmt.Errorf("token is not active")
	}

	// Explicit expiry check: reject if exp is set and already passed.
	// The introspection endpoint should already handle this, but we enforce
	// it locally to guard against clock skew or introspection bugs.
	if result.Exp > 0 && time.Now().Unix() >= result.Exp {
		return nil, fmt.Errorf("token has expired")
	}

	// Audience validation: ensure the token was issued for this MCP server.
	// Only enforced when expectedAudience is configured (MCP_AUDIENCE env var).
	if v.expectedAudience != "" {
		if !containsAudience(result.Aud, v.expectedAudience) {
			return nil, fmt.Errorf("token audience %v does not include expected audience %q", result.Aud, v.expectedAudience)
		}
	}

	return &UserInfo{
		Sub:      result.Sub,
		Username: result.Username,
		ClientID: result.ClientID,
		Scopes:   strings.Fields(result.Scope),
	}, nil
}

// containsAudience reports whether aud contains the target value.
func containsAudience(aud []string, target string) bool {
	for _, a := range aud {
		if a == target {
			return true
		}
	}
	return false
}

// MiddlewareConfig holds server-level defaults for token validation.
// Per-request headers (X-Base-URL, X-Client-ID, X-Client-Secret) override these when present.
type MiddlewareConfig struct {
	DefaultBaseURL      string
	DefaultClientID     string
	DefaultClientSecret string
	ExpectedAudience    string
	SkipTLSVerify       bool
}

// Middleware wraps an http.Handler with bearer token validation.
// Credentials are resolved per-request: headers take priority, server defaults are the fallback.
func Middleware(cfg MiddlewareConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public endpoints (OAuth discovery, proxy, health).
		publicPaths := []string{
			"/health",
			"/.well-known/oauth-authorization-server",
			"/oauth/register",
			"/oauth/authorize",
			"/oauth/callback",
			"/oauth/token",
			"/authorize",  // VS Code fallback path when DCR fails
			"/token",
			"/register",
		}
		reqPath := strings.TrimRight(r.URL.Path, "/")
		if reqPath == "" {
			reqPath = "/"
		}
		for _, p := range publicPaths {
			if reqPath == p {
				next.ServeHTTP(w, r)
				return
			}
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="asgardeo-mcp"`)
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.Header().Set("WWW-Authenticate", `Bearer realm="asgardeo-mcp"`)
			http.Error(w, `{"error":"invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// Resolve credentials: per-request headers override server defaults.
		baseURL := r.Header.Get(internal_config.HeaderBaseURL)
		if baseURL == "" {
			baseURL = cfg.DefaultBaseURL
		}
		clientID := r.Header.Get(internal_config.HeaderClientID)
		if clientID == "" {
			clientID = cfg.DefaultClientID
		}
		clientSecret := r.Header.Get(internal_config.HeaderClientSecret)
		if clientSecret == "" {
			clientSecret = cfg.DefaultClientSecret
		}

		validator := NewTokenValidator(baseURL, clientID, clientSecret, cfg.ExpectedAudience, cfg.SkipTLSVerify)
		userInfo, err := validator.Validate(r.Context(), token)
		if err != nil {
			log.Printf("Token validation failed for request to %s", r.URL.Path)
			w.Header().Set("WWW-Authenticate", `Bearer realm="asgardeo-mcp", error="invalid_token"`)
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Inject base URL, user info, and raw token into context for downstream handlers.
		ctx := context.WithValue(r.Context(), baseURLKey, baseURL)
		ctx = context.WithValue(ctx, userInfoKey, userInfo)
		ctx = context.WithValue(ctx, rawTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserInfo retrieves UserInfo from context.
func GetUserInfo(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(userInfoKey).(*UserInfo)
	return userInfo, ok
}

// GetRawToken retrieves the raw bearer token from context.
// Returns the token string and true if a user token was injected by the middleware.
func GetRawToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(rawTokenKey).(string)
	return token, ok && token != ""
}

// GetBaseURL retrieves the resolved Asgardeo base URL from context.
// Set by the middleware from the X-Base-URL header (or server default).
func GetBaseURL(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(baseURLKey).(string)
	return v, ok && v != ""
}
