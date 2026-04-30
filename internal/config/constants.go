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

package config

const (
	BASE_URL_PARAM         = "BASE_URL"
	CLIENT_ID_PARAM        = "CLIENT_ID"
	CLIENT_SECRET_PARAM    = "CLIENT_SECRET"
	CERTIFICATE_PATH_PARAM = "CERT_PATH"
	PRODUCT_MODE_PARAM     = "PRODUCT_MODE"
	TRANSPORT_MODE_PARAM   = "TRANSPORT_MODE"
	SERVER_PORT_PARAM      = "SERVER_PORT"
	SERVER_BASE_URL_PARAM  = "SERVER_BASE_URL"
	// MCP_AUDIENCE_PARAM is the expected audience claim in validated tokens.
	// Optional — when set, the middleware enforces that the token's aud claim contains this value.
	MCP_AUDIENCE_PARAM = "MCP_AUDIENCE"

	// MCP_INTROSPECT_CLIENT_ID_PARAM and MCP_INTROSPECT_CLIENT_SECRET_PARAM are the
	// credentials used to authenticate the introspection call in SSE mode.
	// Set these to the standard app (e.g. MCP-test) credentials.
	// Falls back to CLIENT_ID / CLIENT_SECRET if not set.
	MCP_INTROSPECT_CLIENT_ID_PARAM     = "MCP_INTROSPECT_CLIENT_ID"
	MCP_INTROSPECT_CLIENT_SECRET_PARAM = "MCP_INTROSPECT_CLIENT_SECRET"

	// SKIP_TLS_VERIFY_PARAM disables TLS certificate verification.
	// Set to "true" for local development with self-signed certs only.
	// Never use in production.
	SKIP_TLS_VERIFY_PARAM = "SKIP_TLS_VERIFY"

	// MCP_CLIENT_ID_PARAM is the client ID of the user-facing MCP application.
	// Used in /.well-known/oauth-authorization-server and /oauth/register so MCP
	// clients (VS Code, Cursor) can discover which client_id to use for the OAuth flow.
	// Falls back to MCP_INTROSPECT_CLIENT_ID if not set.
	MCP_CLIENT_ID_PARAM = "MCP_CLIENT_ID"

	// Per-request header names for multi-tenant SSE mode.
	// Users set these in mcp.json to authenticate against their own Asgardeo org.
	// When present they override the server-level env var defaults.
	HeaderBaseURL      = "X-Base-URL"
	HeaderClientID     = "X-Client-ID"
	HeaderClientSecret = "X-Client-Secret"
)

// DefaultMCPScopes is the set of scopes advertised in the OAuth metadata endpoint.
// Covers all MCP tool operations.
var DefaultMCPScopes = []string{
	"openid",
	"internal_application_mgt_view",
	"internal_application_mgt_create",
	"internal_application_mgt_update",
	"internal_application_mgt_delete",
	"internal_api_resource_view",
	"internal_api_resource_create",
	"internal_api_resource_update",
	"internal_user_mgt_create",
	"internal_claim_meta_view",
	"internal_idp_view",
	"internal_authenticator_view",
}

const (
	TransportStdio = "stdio"
	TransportSSE   = "sse"
	DefaultPort    = "8080"
)

// Deprecated constants for backward compatibility
const (
	ASGARDEO_BASE_URL_PARAM      = "ASGARDEO_BASE_URL"
	ASGARDEO_CLIENT_ID_PARAM     = "ASGARDEO_CLIENT_ID"
	ASGARDEO_CLIENT_SECRET_PARAM = "ASGARDEO_CLIENT_SECRET"
)

var ProductModes = struct {
	WSO2IS   string
	Asgardeo string
}{
	WSO2IS:   "wso2is",
	Asgardeo: "asgardeo",
}

var ProductNames = struct {
	WSO2IS   string
	Asgardeo string
}{
	WSO2IS:   "WSO2 Identity Server",
	Asgardeo: "Asgardeo",
}
