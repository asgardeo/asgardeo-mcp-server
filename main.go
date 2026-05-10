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
	"log"
	"net/http"
	"os"

	"github.com/asgardeo/mcp/internal/auth"
	"github.com/asgardeo/mcp/internal/config"
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

	if len(os.Args) > 1 && os.Args[1] == "http" {
		os.Setenv(config.TRANSPORT_PARAM, config.HTTPTransport)
	}

	if config.GetTransport() == config.HTTPTransport {
		httpConfig, err := config.LoadHTTP()
		if err != nil {
			log.Fatalf("HTTP config error: %v", err)
		}

		streamableHTTPServer := server.NewStreamableHTTPServer(
			s,
			server.WithEndpointPath(httpConfig.Endpoint),
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				if token, ok := auth.AccessTokenFromContext(r.Context()); ok {
					ctx = auth.WithAccessToken(ctx, token)
				}
				if tenantBaseURL, ok := auth.TenantBaseURLFromContext(r.Context()); ok {
					ctx = auth.WithTenantBaseURL(ctx, tenantBaseURL)
				}
				return ctx
			}),
		)

		runtimeConfig, err := config.LoadRuntime()
		if err != nil {
			log.Fatalf("Runtime config error: %v", err)
		}

		mux := http.NewServeMux()
		mux.Handle(httpConfig.Endpoint, streamableHTTPServer)
		mux.Handle(config.TenantPathPrefix, streamableHTTPServer)
		authMiddleware := auth.NewMiddleware(runtimeConfig.BaseURL, httpConfig)
		handler := authMiddleware.Protect(mux)

		log.Printf("Asgardeo MCP streamable HTTP server listening on %s%s", httpConfig.Addr, httpConfig.Endpoint)
		if err := http.ListenAndServe(httpConfig.Addr, handler); err != nil {
			log.Fatalf("Server error: %v", err)
		}
		return
	}

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
