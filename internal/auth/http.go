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
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/asgardeo/mcp/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type accessTokenContextKey struct{}
type tenantBaseURLContextKey struct{}

type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported,omitempty"`
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
	ResourceName           string   `json:"resource_name,omitempty"`
}

type Middleware struct {
	baseURL          string
	httpConfig       config.HTTPConfig
	expectedClientID string

	mu       sync.RWMutex
	keyCache map[string]cachedKeys
}

type cachedKeys struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

func NewMiddleware(baseURL string, httpConfig config.HTTPConfig) *Middleware {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Middleware{
		baseURL:          baseURL,
		httpConfig:       httpConfig,
		expectedClientID: httpConfig.ClientID,
		keyCache:         map[string]cachedKeys{},
	}
}

func metadataURL(resourceURL string) string {
	parsed, err := url.Parse(resourceURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(resourceURL, "/") + "/.well-known/oauth-protected-resource"
	}
	return parsed.Scheme + "://" + parsed.Host + "/.well-known/oauth-protected-resource" + parsed.Path
}

func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenContextKey{}, token)
}

func AccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(accessTokenContextKey{}).(string)
	return token, ok && token != ""
}

func WithTenantBaseURL(ctx context.Context, baseURL string) context.Context {
	return context.WithValue(ctx, tenantBaseURLContextKey{}, baseURL)
}

func TenantBaseURLFromContext(ctx context.Context) (string, bool) {
	baseURL, ok := ctx.Value(tenantBaseURLContextKey{}).(string)
	return baseURL, ok && baseURL != ""
}

func (m *Middleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgHandle := orgHandleFromRequest(r)
		if m.isMetadataRequest(r) {
			if orgHandle == "" {
				http.Error(w, "missing organization handle", http.StatusBadRequest)
				return
			}
			m.writeMetadata(w, r, orgHandle)
			return
		}
		if orgHandle == "" {
			m.writeUnauthorized(w, r, "", "invalid_request", "missing organization handle")
			return
		}

		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			m.writeUnauthorized(w, r, orgHandle, "invalid_request", "missing bearer token")
			return
		}

		tenantBaseURL := config.TenantBaseURL(m.baseURL, orgHandle)
		if err := m.validate(r.Context(), tenantBaseURL, token); err != nil {
			m.writeUnauthorized(w, r, orgHandle, "invalid_token", err.Error())
			return
		}

		ctx := WithAccessToken(r.Context(), token)
		ctx = WithTenantBaseURL(ctx, tenantBaseURL)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) isMetadataRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/.well-known/oauth-protected-resource")
}

func (m *Middleware) writeMetadata(w http.ResponseWriter, r *http.Request, orgHandle string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(m.metadata(r, orgHandle))
}

func (m *Middleware) writeUnauthorized(w http.ResponseWriter, r *http.Request, orgHandle, code, description string) {
	resourceURL := m.resourceURL(r, orgHandle)
	challenge := fmt.Sprintf(`Bearer resource_metadata="%s"`, metadataURL(resourceURL))
	if len(m.httpConfig.Scopes) > 0 {
		challenge += fmt.Sprintf(`, scope="%s"`, strings.Join(m.httpConfig.Scopes, " "))
	}
	if code != "" {
		challenge += fmt.Sprintf(`, error="%s"`, code)
	}
	if description != "" {
		challenge += fmt.Sprintf(`, error_description="%s"`, sanitizeHeaderValue(description))
	}
	w.Header().Set("WWW-Authenticate", challenge)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (m *Middleware) validate(ctx context.Context, tenantBaseURL, tokenString string) error {
	claims := accessTokenClaims{}
	issuer := m.issuer(tenantBaseURL)
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing token key id")
		}
		return m.key(ctx, tenantBaseURL, kid)
	}, jwt.WithIssuer(issuer), jwt.WithAudience(m.audience(tenantBaseURL)), jwt.WithExpirationRequired())
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	if claims.ClientID != "" && claims.ClientID != m.expectedClientID {
		return fmt.Errorf("token client_id does not match expected MCP client")
	}
	if claims.AuthorizedParty != "" && claims.AuthorizedParty != m.expectedClientID {
		return fmt.Errorf("token azp does not match expected MCP client")
	}
	return nil
}

type accessTokenClaims struct {
	jwt.RegisteredClaims
	ClientID        string `json:"client_id,omitempty"`
	AuthorizedParty string `json:"azp,omitempty"`
}

func (m *Middleware) key(ctx context.Context, tenantBaseURL, kid string) (*rsa.PublicKey, error) {
	jwksURL := tenantBaseURL + "/oauth2/jwks"
	m.mu.RLock()
	cache := m.keyCache[jwksURL]
	key, ok := cache.keys[kid]
	valid := time.Now().Before(cache.expiresAt)
	m.mu.RUnlock()
	if ok && valid {
		return key, nil
	}

	if err := m.refreshKeys(ctx, jwksURL); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	cache = m.keyCache[jwksURL]
	key, ok = cache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("unknown token key id")
	}
	return key, nil
}

func (m *Middleware) refreshKeys(ctx context.Context, jwksURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: HTTP %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			KID string `json:"kid"`
			KTY string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	keys := map[string]*rsa.PublicKey{}
	for _, jwk := range jwks.Keys {
		if jwk.KTY != "RSA" || jwk.KID == "" {
			continue
		}
		key, err := rsaKey(jwk.N, jwk.E)
		if err != nil {
			continue
		}
		keys[jwk.KID] = key
	}

	m.mu.Lock()
	m.keyCache[jwksURL] = cachedKeys{
		keys:      keys,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	m.mu.Unlock()
	return nil
}

func (m *Middleware) metadata(r *http.Request, orgHandle string) ProtectedResourceMetadata {
	tenantBaseURL := config.TenantBaseURL(m.baseURL, orgHandle)
	return ProtectedResourceMetadata{
		Resource:               m.resourceURL(r, orgHandle),
		AuthorizationServers:   []string{m.issuer(tenantBaseURL)},
		ScopesSupported:        m.httpConfig.Scopes,
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Asgardeo Management MCP",
	}
}

func (m *Middleware) audience(tenantBaseURL string) string {
	if m.httpConfig.TokenAudience != "" {
		orgHandle := strings.TrimPrefix(tenantBaseURL, m.baseURL+config.TenantPathPrefix)
		return strings.ReplaceAll(m.httpConfig.TokenAudience, "<org>", orgHandle)
	}
	return tenantBaseURL + "/oauth2/token"
}

func (m *Middleware) issuer(tenantBaseURL string) string {
	if m.httpConfig.TokenIssuer != "" {
		orgHandle := strings.TrimPrefix(tenantBaseURL, m.baseURL+config.TenantPathPrefix)
		return strings.ReplaceAll(m.httpConfig.TokenIssuer, "<org>", orgHandle)
	}
	return tenantBaseURL + "/oauth2/token"
}

func (m *Middleware) resourceURL(r *http.Request, orgHandle string) string {
	serverURL := strings.ReplaceAll(m.httpConfig.ServerURL, "<org>", orgHandle)
	if !strings.Contains(serverURL, "<") {
		return serverURL
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s%s%s", scheme, r.Host, config.TenantPathPrefix, orgHandle, m.httpConfig.Endpoint)
}

func orgHandleFromRequest(r *http.Request) string {
	if orgHandle := strings.Trim(r.URL.Query().Get("org"), "/"); orgHandle != "" {
		return orgHandle
	}
	if orgHandle := strings.Trim(r.URL.Query().Get("org_handle"), "/"); orgHandle != "" {
		return orgHandle
	}
	if orgHandle := strings.Trim(r.Header.Get("X-Asgardeo-Org"), "/"); orgHandle != "" {
		return orgHandle
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "t" && parts[i+1] != "" {
			return parts[i+1]
		}
	}
	return ""
}

func rsaKey(nValue, eValue string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nValue)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eValue)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, `"`, "'")
	return strings.ReplaceAll(value, "\n", " ")
}
