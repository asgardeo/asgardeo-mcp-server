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

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/asgardeo/mcp/internal/auth"
	internal_config "github.com/asgardeo/mcp/internal/config"
	"github.com/asgardeo/mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// setupServer configures the MCP server and registers tools.
func setupServer() *server.MCPServer {
	s := server.NewMCPServer(
		"Asgardeo Management MCP",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	appListTool, appListToolImpl := tools.GetListApplicationsTool()
	s.AddTool(appListTool, appListToolImpl)

	spaTool, spaToolImpl := tools.GetCreateSinglePageAppTool()
	s.AddTool(spaTool, spaToolImpl)

	webAppTool, webAppToolImpl := tools.GetCreateWebAppWithSSRTool()
	s.AddTool(webAppTool, webAppToolImpl)

	mobileAppTool, mobileAppToolImpl := tools.GetCreateMobileAppTool()
	s.AddTool(mobileAppTool, mobileAppToolImpl)

	m2mAppTool, m2mAppToolImpl := tools.GetCreateM2MAppTool()
	s.AddTool(m2mAppTool, m2mAppToolImpl)

	getAppByNameTool, getAppByNameToolmpl := tools.GetSearchApplicationByNameTool()
	s.AddTool(getAppByNameTool, getAppByNameToolmpl)

	getAppByClientIdTool, getAppByClientIdToolmpl := tools.GetSearchApplicationByClientIdTool()
	s.AddTool(getAppByClientIdTool, getAppByClientIdToolmpl)

	getAppUpdateTool, getAppUpdateToolImpl := tools.GetUpdateApplicationBasicInfoTool()
	s.AddTool(getAppUpdateTool, getAppUpdateToolImpl)

	getAppOAuthConfigUpdateTool, getAppUpdateOAuthConfigToolImpl := tools.GetUpdateApplicationOAuthConfigTool()
	s.AddTool(getAppOAuthConfigUpdateTool, getAppUpdateOAuthConfigToolImpl)

	updateApplicationClaimConfigTool, updateApplicationClaimConfigToolImpl := tools.GetUpdateApplicationClaimConfigTool()
	s.AddTool(updateApplicationClaimConfigTool, updateApplicationClaimConfigToolImpl)

	authorizeAPITool, authorizeAPIToolImpl := tools.GetAuthorizeAPITool()
	s.AddTool(authorizeAPITool, authorizeAPIToolImpl)

	authorizedAPIListTool, authorizedAPIListToolImpl := tools.GetListAuthorizedAPITool()
	s.AddTool(authorizedAPIListTool, authorizedAPIListToolImpl)

	updateLoginFlowTool, updateLoginFlowToolImpl := tools.GetUpdateLoginFlowTool()
	s.AddTool(updateLoginFlowTool, updateLoginFlowToolImpl)

	apiResourceListTool, apiResourceListToolImpl := tools.GetListAPIResourcesTool()
	s.AddTool(apiResourceListTool, apiResourceListToolImpl)

	apiResourceListByNameTool, apiResourceListByNameToolImpl := tools.GetSearchAPIResourcesByNameTool()
	s.AddTool(apiResourceListByNameTool, apiResourceListByNameToolImpl)

	apiResourceSearchByIdentifierTool, apiResourceSearchByIdentifierToolImpl := tools.GetSearchAPIResourceByIdentifierTool()
	s.AddTool(apiResourceSearchByIdentifierTool, apiResourceSearchByIdentifierToolImpl)

	apiResourceCreateTool, apiResourceCreateToolImpl := tools.GetCreateAPIResourceTool()
	s.AddTool(apiResourceCreateTool, apiResourceCreateToolImpl)

	userCreateTool, userCreateToolImpl := tools.GetCreateUserTool()
	s.AddTool(userCreateTool, userCreateToolImpl)

	listClaimsTool, listClaimsToolImpl := tools.GetListClaimsTool()
	s.AddTool(listClaimsTool, listClaimsToolImpl)

	return s
}

func main() {
	s := setupServer()

	transport := internal_config.GetTransportMode()
	log.Printf("Starting Asgardeo MCP server in %s mode", transport)

	switch transport {
	case internal_config.TransportSSE:
		port := internal_config.GetServerPort()
		baseURL := internal_config.GetServerBaseURL()
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", port)
		}

		sseServer := server.NewSSEServer(s, server.WithBaseURL(baseURL))

		// Build HTTP mux with auth middleware and health endpoint.
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Alias handlers — VS Code falls back to these paths when DCR fails and
		// the user manually provides a client ID. Redirect to the /oauth/* equivalents.
		mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/oauth/authorize?"+r.URL.RawQuery, http.StatusFound)
		})
		mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/oauth/token", http.StatusTemporaryRedirect)
		})
		mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/oauth/register", http.StatusTemporaryRedirect)
		})

		// oauthFlowEntry holds per-user credentials and redirect URI across the authorize→callback round trip.
		// Credentials come from X-Base-URL / X-Client-ID / X-Client-Secret headers (or server defaults).
		// Entries expire after 10 minutes to prevent unbounded growth from abandoned flows.
		type oauthFlowEntry struct {
			redirectURI  string
			clientID     string
			clientSecret string
			baseURL      string
			expiresAt    time.Time
		}
		var oauthFlowStore sync.Map // state → oauthFlowEntry
		var oauthCodeStore sync.Map // code → oauthFlowEntry (for /oauth/token credential lookup)

		// Purge stale entries every 5 minutes.
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				now := time.Now()
				oauthFlowStore.Range(func(key, value any) bool {
					if e, ok := value.(oauthFlowEntry); ok && now.After(e.expiresAt) {
						oauthFlowStore.Delete(key)
					}
					return true
				})
			}
		}()

		// proxyHTTPClient is used for token endpoint proxying.
		proxyHTTPClient := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: internal_config.GetSkipTLSVerify()},
			},
		}

		// callbackURI is the fixed redirect URI registered in Asgardeo.
		callbackURI := baseURL + "/oauth/callback"
		asgardeoBaseURL := internal_config.GetBaseURL()

		// OAuth Authorization Server Metadata (RFC 8414 / MCP Authorization spec).
		// issuer reflects the user's own Asgardeo tenant (from X-Base-URL header, or server default).
		mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
			issuer := r.Header.Get(internal_config.HeaderBaseURL)
			if issuer == "" {
				issuer = asgardeoBaseURL
			}
			metadata := map[string]any{
				"issuer":                               issuer,
				"authorization_endpoint":               baseURL + "/oauth/authorize",
				"token_endpoint":                       baseURL + "/oauth/token",
				"device_authorization_endpoint":        issuer + "/oauth2/device_authorize",
				"registration_endpoint":                baseURL + "/oauth/register",
				"scopes_supported":                     internal_config.DefaultMCPScopes,
				"response_types_supported":             []string{"code"},
				"grant_types_supported":                []string{"authorization_code", "urn:ietf:params:oauth:grant-type:device_code"},
				"code_challenge_methods_supported":     []string{"S256"},
				"token_endpoint_auth_methods_supported": []string{"none", "client_secret_basic"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metadata)
		})

		// Dynamic Client Registration (RFC 7591) — simplified.
		// Returns the user's own client ID (from X-Client-ID header, or server default).
		mux.HandleFunc("/oauth/register", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			clientID := r.Header.Get(internal_config.HeaderClientID)
			if clientID == "" {
				clientID = internal_config.GetMCPClientID()
			}
			log.Printf("OAuth DCR: registration request received")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"client_id":                  clientID,
				"client_secret_expires_at":   0,
				"token_endpoint_auth_method": "none",
				"grant_types":                []string{"authorization_code", "urn:ietf:params:oauth:grant-type:device_code"},
				"response_types":             []string{"code"},
				"code_challenge_methods":     []string{"S256"},
				"redirect_uris":              []string{callbackURI},
			})
		})

		// /oauth/authorize — proxy that swaps VS Code's dynamic redirect URI with our fixed one.
		// Saves the original redirect URI and per-user credentials keyed by state.
		mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			state := params.Get("state")
			originalRedirect := params.Get("redirect_uri")

			// Resolve per-user credentials from headers, fall back to server defaults.
			entryBaseURL := r.Header.Get(internal_config.HeaderBaseURL)
			if entryBaseURL == "" {
				entryBaseURL = asgardeoBaseURL
			}
			entryClientID := r.Header.Get(internal_config.HeaderClientID)
			if entryClientID == "" {
				entryClientID = internal_config.GetMCPClientID()
			}
			entryClientSecret := r.Header.Get(internal_config.HeaderClientSecret)
			if entryClientSecret == "" {
				entryClientSecret = internal_config.GetMCPClientSecret()
			}

			if state != "" && originalRedirect != "" {
				oauthFlowStore.Store(state, oauthFlowEntry{
					redirectURI:  originalRedirect,
					clientID:     entryClientID,
					clientSecret: entryClientSecret,
					baseURL:      entryBaseURL,
					expiresAt:    time.Now().Add(10 * time.Minute),
				})
			}

			params.Set("redirect_uri", callbackURI)
			target := entryBaseURL + "/oauth2/authorize?" + params.Encode()
			log.Printf("OAuth proxy: forwarding authorize request to Asgardeo")
			http.Redirect(w, r, target, http.StatusFound)
		})

		// /oauth/callback — receives Asgardeo's redirect and sends the code back to VS Code.
		mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
			state := r.URL.Query().Get("state")
			code := r.URL.Query().Get("code")
			errParam := r.URL.Query().Get("error")

			val, ok := oauthFlowStore.LoadAndDelete(state)
			if !ok {
				http.Error(w, "Unknown or expired state", http.StatusBadRequest)
				return
			}
			entry := val.(oauthFlowEntry)
			if time.Now().After(entry.expiresAt) {
				http.Error(w, "OAuth state has expired", http.StatusBadRequest)
				return
			}
			vsCodeRedirect := entry.redirectURI

			// Store credentials keyed by code so /oauth/token can retrieve them.
			// The browser redirect to VS Code carries no headers, so credentials
			// must be looked up from here rather than read from the token request.
			if code != "" {
				oauthCodeStore.Store(code, entry)
			}

			params := url.Values{}
			if code != "" {
				params.Set("code", code)
			}
			if state != "" {
				params.Set("state", state)
			}
			if errParam != "" {
				params.Set("error", errParam)
				if errDesc := r.URL.Query().Get("error_description"); errDesc != "" {
					params.Set("error_description", errDesc)
				}
			}

			finalURL := vsCodeRedirect + "?" + params.Encode()
			log.Printf("OAuth proxy: callback complete, redirecting to VS Code")
			http.Redirect(w, r, finalURL, http.StatusFound)
		})

		// /oauth/token — proxies token exchange to Asgardeo, swapping redirect_uri.
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			params, err := url.ParseQuery(string(body))
			if err != nil {
				http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
				return
			}
			params.Set("redirect_uri", callbackURI)

			// Resolve credentials: look up from code store first (most reliable),
			// then fall back to request headers, then server defaults.
			tokenBaseURL := asgardeoBaseURL
			tokenClientID := internal_config.GetMCPClientID()
			tokenClientSecret := internal_config.GetMCPClientSecret()

			if codeVal, ok := oauthCodeStore.LoadAndDelete(params.Get("code")); ok {
				e := codeVal.(oauthFlowEntry)
				tokenBaseURL = e.baseURL
				tokenClientID = e.clientID
				tokenClientSecret = e.clientSecret
			} else {
				if v := r.Header.Get(internal_config.HeaderBaseURL); v != "" {
					tokenBaseURL = v
				}
				if v := r.Header.Get(internal_config.HeaderClientID); v != "" {
					tokenClientID = v
				}
				if v := r.Header.Get(internal_config.HeaderClientSecret); v != "" {
					tokenClientSecret = v
				}
			}

			req, err := http.NewRequestWithContext(r.Context(), http.MethodPost,
				tokenBaseURL+"/oauth2/token", strings.NewReader(params.Encode()))
			if err != nil {
				http.Error(w, "Failed to create token request", http.StatusInternalServerError)
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			// VS Code uses public-client PKCE (no client_secret). Asgardeo requires
			// client credentials at the token endpoint — inject them transparently.
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			} else {
				req.SetBasicAuth(tokenClientID, tokenClientSecret)
			}

			resp, err := proxyHTTPClient.Do(req)
			if err != nil {
				log.Printf("Token proxy error: %v", err)
				http.Error(w, "Token endpoint unreachable", http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("OAuth token proxy: status=%d", resp.StatusCode)

			for k, vals := range resp.Header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			w.Write(respBody)
		})

		mux.Handle("/", sseServer)

		// Wrap with token validation middleware.
		// Per-request headers override these server-level defaults at runtime.
		// For single-app setup, CLIENT_ID / CLIENT_SECRET is sufficient — no separate
		// MCP_CLIENT_ID or MCP_INTROSPECT_* vars needed.
		handler := auth.Middleware(auth.MiddlewareConfig{
			DefaultBaseURL:      asgardeoBaseURL,
			DefaultClientID:     internal_config.GetMCPClientID(),
			DefaultClientSecret: internal_config.GetMCPClientSecret(),
			ExpectedAudience:    internal_config.GetMCPAudience(),
			SkipTLSVerify:       internal_config.GetSkipTLSVerify(),
		}, mux)

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
		}

		// Graceful shutdown.
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			log.Println("Shutting down server...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			httpServer.Shutdown(ctx)
		}()

		log.Printf("SSE server listening on :%s (auth enabled)", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("SSE server error: %v", err)
		}

	default:
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Stdio server error: %v", err)
		}
	}
}
