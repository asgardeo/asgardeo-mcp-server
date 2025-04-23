package applications

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shashimalcse/go-asgardeo/management"
	"github.com/thilinashashimalsenarath/asgardeo-mcp/internal/asgardeo"
)

func List() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}
	// Register ListAsgardeoApplication tool
	appListTool := mcp.NewTool("ListAsgardeoApplication",
		mcp.WithDescription("List Asgardeo Applications"),
	)

	var appListToolImpl server.ToolHandlerFunc

	appListToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := client.Applications().List(ctx, nil)
		if err != nil {
			log.Printf("Error listing applications: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return appListTool, appListToolImpl
}

func Get() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}
	// Register ListAsgardeoApplication tool
	appDetailFetchTool := mcp.NewTool("GetAsgardeoApplicationDetails",
		mcp.WithDescription("Get info about an Asgardeo Application"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the asgardeo application"),
		),
	)

	var appDetailFetchToolImpl server.ToolHandlerFunc

	appDetailFetchToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id := req.Params.Arguments["id"].(string)

		resp, err := client.Applications().Get(ctx, id)
		if err != nil {
			log.Printf("Error getting details of Asgardeo application with id %v: %v", id, err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return appDetailFetchTool, appDetailFetchToolImpl
}

func Create() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	createAppTool := mcp.NewTool("createAsgardeoAppTool",
		mcp.WithDescription("Create an Asgardeo Application"),

		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the Asgardeo application"),
		),

		// Advanced Configurations
		mcp.WithBoolean("discoverableByEndUsers",
			mcp.Description("Whether the app is discoverable by end users"),
		),
		mcp.WithBoolean("skipLogoutConsent",
			mcp.Description("Whether to skip logout consent"),
		),
		mcp.WithBoolean("skipLoginConsent",
			mcp.Description("Whether to skip login consent"),
		),

		// Authentication Sequence
		mcp.WithString("authenticationSequenceType",
			mcp.Description("Type of the authentication sequence (e.g., DEFAULT)"),
		),
		mcp.WithArray("authenticationSteps",
			mcp.Description("Authentication steps including idp and authenticator"),
		),

		// Claim Configuration
		mcp.WithString("claimDialect",
			mcp.Description("Dialect for requested claims (e.g., LOCAL)"),
		),
		mcp.WithArray("requestedClaims",
			mcp.Description("List of requested claim URIs"),
		),

		// OIDC Inbound Configuration
		mcp.WithArray("grantTypes",
			mcp.Description("OIDC grant types"),
		),
		mcp.WithArray("allowedOrigins",
			mcp.Description("Allowed CORS origins"),
		),
		mcp.WithArray("callbackURLs",
			mcp.Description("Authorized redirect URLs for the Asgardeo application"),
		),

		mcp.WithBoolean("pkceMandatory",
			mcp.Description("Is PKCE mandatory"),
		),
		mcp.WithBoolean("supportPlainTransformAlgorithm",
			mcp.Description("Support plain PKCE transformation algorithm"),
		),
		mcp.WithBoolean("publicClient",
			mcp.Description("Is the client public"),
		),

		// Token Configurations
		mcp.WithNumber("userAccessTokenExpiryInSeconds",
			mcp.Description("Access token expiry time in seconds"),
		),
		mcp.WithNumber("applicationAccessTokenExpiryInSeconds",
			mcp.Description("Application access token expiry time in seconds"),
		),
		mcp.WithString("accessTokenBindingType",
			mcp.Description("Access token binding type"),
		),
		mcp.WithBoolean("revokeTokensWhenIDPSessionTerminated",
			mcp.Description("Revoke tokens when IDP session terminates"),
		),
		mcp.WithBoolean("validateTokenBinding",
			mcp.Description("Validate token binding"),
		),
		mcp.WithNumber("refreshTokenExpiryInSeconds",
			mcp.Description("Refresh token expiry in seconds"),
		),
		mcp.WithBoolean("renewRefreshToken",
			mcp.Description("Whether to renew refresh tokens"),
		),

		// Template & Role Configurations
		mcp.WithString("templateId",
			mcp.DefaultString("custom-application-oidc"),
			mcp.Description("ID of the application template to use"),
		),
	)

	var createAppToolImpl server.ToolHandlerFunc

	createAppToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		name := req.Params.Arguments["name"].(string)
		// callbackURLs := req.Params.Arguments["callbackURLs"].([]string)

		// // Create a new application
		// var newCallbackURLs []string
		// if callbackURLs != nil {
		// 	newCallbackURLs = callbackURLs
		// } else {
		// 	newCallbackURLs = []string{}
		// }

		newApp := management.ApplicationCreateInput{
			Name: name,
			// InboundProtocolConfiguration: &management.InboundProtocolConfiguration{
			// 	OIDC: &management.InboundOIDCConfig{
			// 		CallbackURLs: newCallbackURLs,
			// 	},
			// },
		}

		resp, err := client.Applications().Create(ctx, newApp)
		if err != nil {
			log.Printf("Error creating an Asgardeo application: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return createAppTool, createAppToolImpl
}
