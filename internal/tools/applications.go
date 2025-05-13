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

package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asgardeo/go/pkg/application"
	"github.com/asgardeo/mcp/internal/asgardeo"
	"github.com/asgardeo/mcp/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetListApplicationsTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}
	// Register ListAsgardeoApplication tool
	appListTool := mcp.NewTool("list_applications",
		mcp.WithDescription("List all applications in Asgardeo"),
	)

	appListToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := client.Application.List(ctx, 10, 0)
		if err != nil {
			log.Printf("Error listing applications: %v", err)
			return nil, err
		}
		apps := []interface{}{}
		for _, app := range *resp.Applications {
			appName := *app.Name
			appID := *app.Id
			apps = append(apps, map[string]interface{}{
				"name": appName,
				"id":   appID,
			})
		}

		return mcp.NewToolResultText(fmt.Sprintf("%+v", apps)), nil
	}

	return appListTool, appListToolImpl
}

func GetCreateSinglePageAppTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	spaTool := mcp.NewTool("create_single_page_app",
		mcp.WithDescription("Create a new Single Page Application in Asgardeo"),
		mcp.WithString("application_name", mcp.Description("Name of the application"), mcp.Required()),
		mcp.WithString("redirect_url", mcp.Description("Redirect URL of the application"), mcp.Required()),
	)

	spaToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["application_name"].(string)
		redirectURL := req.Params.Arguments["redirect_url"].(string)

		spa, err := client.Application.CreateSinglePageApp(ctx, appName, redirectURL)
		if err != nil {
			log.Printf("Error creating SPA: %v", err)
			return nil, err
		}

		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             spa.Name,
				"id":               spa.Id,
				"client_id":        spa.ClientId,
				"redirect_url":     spa.RedirectURL,
				"scope":            spa.AuthorizedScopes,
				"application_type": string(spa.AppType),
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return spaTool, spaToolImpl
}

func GetCreateWebAppWithSSRTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	webappTool := mcp.NewTool("create_webapp_with_ssr",
		mcp.WithDescription("Create a new regular web application that implements server side rendring in Asgardeo"),
		mcp.WithString("application_name", mcp.Description("Name of the application"), mcp.Required()),
		mcp.WithString("redirect_url", mcp.Description("Redirect URL of the application"), mcp.Required()),
	)

	webappToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["application_name"].(string)
		redirectURL := req.Params.Arguments["redirect_url"].(string)

		webapp, err := client.Application.CreateWebAppWithSSR(ctx, appName, redirectURL)
		if err != nil {
			log.Printf("Error creating SPA: %v", err)
			return nil, err
		}

		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             webapp.Name,
				"id":               webapp.Id,
				"client_id":        webapp.ClientId,
				"client_secret":    webapp.ClientSecret,
				"redirect_url":     webapp.RedirectURL,
				"scope":            webapp.AuthorizedScopes,
				"application_type": string(webapp.AppType),
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return webappTool, webappToolImpl
}

func GetCreateMobileAppTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	mobileAppTool := mcp.NewTool("create_mobile_app",
		mcp.WithDescription("Create a new Mobile Application in Asgardeo"),
		mcp.WithString("application_name", mcp.Description("Name of the application"), mcp.Required()),
		mcp.WithString("redirect_url", mcp.Description("Redirect URL of the application"), mcp.Required()),
	)

	mobileAppToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["application_name"].(string)
		redirectURL := req.Params.Arguments["redirect_url"].(string)

		mobileApp, err := client.Application.CreateMobileApp(ctx, appName, redirectURL)
		if err != nil {
			log.Printf("Error creating mobile app: %v", err)
			return nil, err
		}

		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             mobileApp.Name,
				"id":               mobileApp.Id,
				"client_id":        mobileApp.ClientId,
				"redirect_url":     mobileApp.RedirectURL,
				"scope":            mobileApp.AuthorizedScopes,
				"application_type": string(mobileApp.AppType),
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return mobileAppTool, mobileAppToolImpl
}

func GetCreateM2MAppTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	mobileAppTool := mcp.NewTool("create_m2m_app",
		mcp.WithDescription("Create a new M2M Application in Asgardeo"),
		mcp.WithString("application_name", mcp.Description("Name of the application"), mcp.Required()),
	)

	mobileAppToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["application_name"].(string)

		m2mApp, err := client.Application.CreateM2MApp(ctx, appName)
		if err != nil {
			log.Printf("Error creating mobile app: %v", err)
			return nil, err
		}

		// todo: Need to decide on exposing the client secret to the user here which is the LLM
		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             m2mApp.Name,
				"id":               m2mApp.Id,
				"client_id":        m2mApp.ClientId,
				"client_secret":    m2mApp.ClientSecret,
				"application_type": string(m2mApp.AppType),
			},
			"oauth_endpoints": map[string]string{
				"token_url":    fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":     fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url": fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return mobileAppTool, mobileAppToolImpl
}

func GetSearchApplicationByNameTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	getApplicationByNameTool := mcp.NewTool("get_application_by_name",
		mcp.WithDescription("Get details of an application by name"),
		mcp.WithString("application_name", mcp.Description("Name of the application"), mcp.Required()),
	)

	getApplicationByNameToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["application_name"].(string)

		app, err := client.Application.GetByName(ctx, appName)
		if err != nil {
			log.Printf("Error retrieving app: %v", err)
			return nil, err
		}

		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             app.Name,
				"id":               app.Id,
				"client_id":        app.ClientId,
				"client_secret":    app.ClientSecret,
				"redirect_url":     app.RedirectURL,
				"scope":            app.AuthorizedScopes,
				"application_type": string(app.AppType),
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return getApplicationByNameTool, getApplicationByNameToolImpl
}

func GetSearchApplicationByClientIdTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	getApplicationByClientIDTool := mcp.NewTool("get_application_by_client_id",
		mcp.WithDescription("Get details of an application by client ID"),
		mcp.WithString("client_id", mcp.Description("Client ID of the application"), mcp.Required()),
	)

	getApplicationByClientIDToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appName := req.Params.Arguments["client_id"].(string)

		app, err := client.Application.GetByClienId(ctx, appName)
		if err != nil {
			log.Printf("Error retrieving app: %v", err)
			return nil, err
		}

		baseURL := client.Config.BaseURL
		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":             app.Name,
				"id":               app.Id,
				"client_id":        app.ClientId,
				"client_secret":    app.ClientSecret,
				"redirect_url":     app.RedirectURL,
				"scope":            app.AuthorizedScopes,
				"application_type": string(app.AppType),
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := utils.MarshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return getApplicationByClientIDTool, getApplicationByClientIDToolImpl
}

func GetUpdateApplicationBasicInfoTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	updateApplicationBasicInfoTool := mcp.NewTool("update_application_basic_info",
		mcp.WithDescription("Update basic information of an application"),
		mcp.WithString("id", mcp.Description("ID of the application"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Name of the application")),
		mcp.WithString("description", mcp.Description("Description of the application")),
		mcp.WithString("image_url", mcp.Description("URL of the application image icon")),
		mcp.WithString("access_url", mcp.Description("Access URL of the application")),
		mcp.WithString("logout_return_url", mcp.Description("A URL of the application to return upon logout")),
	)

	updateApplicationBasicInfoToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appId := req.Params.Arguments["id"].(string)

		basicInfoUpdate := application.NewBasicInfoUpdate()
		if name, ok := req.Params.Arguments["name"]; ok && name != nil {
			basicInfoUpdate.WithName(name.(string))
		}
		if description, ok := req.Params.Arguments["description"]; ok && description != nil {
			basicInfoUpdate.WithDescription(description.(string))
		}
		if imageUrl, ok := req.Params.Arguments["image_url"]; ok && imageUrl != nil {
			basicInfoUpdate.WithImageUrl(imageUrl.(string))
		}
		if accessUrl, ok := req.Params.Arguments["access_url"]; ok && accessUrl != nil {
			basicInfoUpdate.WithAccessUrl(accessUrl.(string))
		}
		if logoutReturnUrl, ok := req.Params.Arguments["logout_return_url"]; ok && logoutReturnUrl != nil {
			basicInfoUpdate.WithLogoutReturnUrl(logoutReturnUrl.(string))
		}
		if name, ok := req.Params.Arguments["name"]; ok && name != nil {
			basicInfoUpdate.WithName(name.(string))
		}

		err := client.Application.UpdateBasicInfo(ctx, appId, *basicInfoUpdate)
		if err != nil {
			log.Printf("Error updating application: %v", err)
			return nil, err
		}

		return mcp.NewToolResultText("Successfully updated the application."), nil
	}

	return updateApplicationBasicInfoTool, updateApplicationBasicInfoToolImpl
}

func GetUpdateApplicationOAuthConfigTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	stringTypeSchema := map[string]interface{}{"type": "string"}

	updateApplicationOAuthConfigTool := mcp.NewTool("update_application_oauth_config",
		mcp.WithDescription("Update OAuth/OIDC configurations of an application"),
		mcp.WithString("id", mcp.Description("ID of the application"), mcp.Required()),
		mcp.WithArray("redirect_urls", mcp.Description("Redirect URLs of the application"), mcp.Items(stringTypeSchema)),
		mcp.WithNumber("user_access_token_expiry_time", mcp.Description("Expiry time of the access token issued on behalf of the user")),
		mcp.WithNumber("application_access_token_expiry_time", mcp.Description("Expiry time of the access token issued on behalf of the application")),
		mcp.WithNumber("refresh_token_expiry_time", mcp.Description("Expiry time of the refresh token")),
		mcp.WithArray("allowed_origins", mcp.Description("Allowed origins for CORS"), mcp.Items(stringTypeSchema)),
		mcp.WithBoolean("revoke_tokens_when_idp_session_terminated", mcp.Description("Revoke tokens when IDP session is terminated")),
		mcp.WithArray("access_token_attributes", mcp.Description("Access token attributes"), mcp.Items(stringTypeSchema)),
	)

	updateApplicationOAuthConfigToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appId := req.Params.Arguments["id"].(string)

		OAuthConfigUpdate := application.NewOAuthConfigUpdate()
		if redirectURLs, ok := req.Params.Arguments["redirect_urls"]; ok && redirectURLs != nil {
			urls := convertToStringSlice(redirectURLs)
			OAuthConfigUpdate.WithCallbackURLs(urls)
		}

		if allowedOrigins, ok := req.Params.Arguments["allowed_origins"]; ok && allowedOrigins != nil {
			origins := convertToStringSlice(allowedOrigins)
			OAuthConfigUpdate.WithAllowedOrigins(origins)
		}

		if userExpiry, ok := req.Params.Arguments["user_access_token_expiry_time"]; ok && userExpiry != nil {
			OAuthConfigUpdate.WithUserAccessTokenExpiry(int64(userExpiry.(float64)))
		}

		if appExpiry, ok := req.Params.Arguments["application_access_token_expiry_time"]; ok && appExpiry != nil {
			OAuthConfigUpdate.WithApplicationAccessTokenExpiry(int64(appExpiry.(float64)))
		}

		if refreshExpiry, ok := req.Params.Arguments["refresh_token_expiry_time"]; ok && refreshExpiry != nil {
			OAuthConfigUpdate.WithRefreshTokenExpiry(int64(refreshExpiry.(float64)))
		}

		if attributes, ok := req.Params.Arguments["access_token_attributes"]; ok && attributes != nil {
			attrs := convertToStringSlice(attributes)
			OAuthConfigUpdate.WithAccessTokenAttributes(attrs)
		}

		err := client.Application.UpdateOAuthConfig(ctx, appId, *OAuthConfigUpdate)
		if err != nil {
			log.Printf("Error updating application: %v", err)
			return nil, err
		}

		return mcp.NewToolResultText("Successfully updated the application."), nil
	}

	return updateApplicationOAuthConfigTool, updateApplicationOAuthConfigToolImpl
}

func GetUpdateApplicationClaimConfigTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	stringTypeSchema := map[string]interface{}{"type": "string"}

	updateApplicationClaimConfigTool := mcp.NewTool("update_application_claim_config",
		mcp.WithDescription("Update claim configurations of an application."),
		mcp.WithString("id",
			mcp.Description("ID of the application"),
			mcp.Required(),
		),
		mcp.WithArray("claims",
			mcp.Description("List of claims to be added as requested claims in the application. Eg: list of URIs like http://wso2.org/claims/username, http://wso2.org/claims/emailaddress"),
			mcp.Items(stringTypeSchema),
			mcp.Required(),
		),
	)

	updateApplicationClaimConfigToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appId := req.Params.Arguments["id"].(string)
		claimsInput, ok := req.Params.Arguments["claims"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid claim format: expected an array of strings")
		}

		claimConfigs := make([]application.RequestedClaimModel, len(claimsInput))
		mandatory := false

		for i, claimInput := range claimsInput {
			uri, uriOk := claimInput.(string)
			if !uriOk {
				return nil, fmt.Errorf("invalid claim URI at index %d: expected a string", i)
			}

			claimConfigs[i] = application.RequestedClaimModel{
				Claim: application.ClaimModel{
					Uri: uri,
				},
				Mandatory: &mandatory,
			}
		}

		claimConfiguration := application.ApplicationClaimConfigurationUpdateModel{
			RequestedClaims: &claimConfigs,
		}
		err := client.Application.UpdateClaimConfig(ctx, appId, claimConfiguration)
		if err != nil {
			log.Printf("Error updating the claim configuration of the application: %v", err)
			return nil, err
		}

		return mcp.NewToolResultText("Successfully updated the claim configuration of the application."), nil
	}

	return updateApplicationClaimConfigTool, updateApplicationClaimConfigToolImpl
}

func GetAuthorizeAPITool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	stringTypeSchema := map[string]interface{}{"type": "string"}

	authorizeAPITool := mcp.NewTool("authorize_api",
		mcp.WithDescription("Authorize Asgardeo API"),
		mcp.WithString("appId",
			mcp.Required(),
			mcp.Description("This is the id of the application."),
		),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("This is the id of the API resource to be authorized."),
		),
		mcp.WithString("policyIdentifier",
			mcp.Required(),
			mcp.DefaultString("RBAC"),
			mcp.Description("This indicates the authorization policy of the API authorization."),
		),
		mcp.WithArray("scopes",
			mcp.Required(),
			mcp.DefaultArray([]string{}),
			mcp.Description("This is the list of scope names for the API resource."),
			mcp.Items(stringTypeSchema),
		),
	)
	authorizeAPIToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appId := req.Params.Arguments["appId"].(string)
		id := req.Params.Arguments["id"].(string)
		policyIdentifier := req.Params.Arguments["policyIdentifier"].(string)
		rawScopes := req.Params.Arguments["scopes"].([]interface{})
		scopes := make([]string, len(rawScopes))
		for i, s := range rawScopes {
			scopes[i] = s.(string)
		}
		authorizedAPI := application.AuthorizedAPICreateModel{
			Id:               &id,
			PolicyIdentifier: &policyIdentifier,
			Scopes:           &scopes,
		}

		err := client.Application.AuthorizeAPI(ctx, appId, authorizedAPI)
		if err != nil {
			log.Printf("Error authorizing API resource: %v", err)
			return nil, err
		}

		return mcp.NewToolResultText("API authorization successful."), nil
	}

	return authorizeAPITool, authorizeAPIToolImpl
}

func GetListAuthorizedAPITool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	authorizedAPIListTool := mcp.NewTool("list_authorized_api",
		mcp.WithDescription("List authorized API resources of an application"),
		mcp.WithString("app_id",
			mcp.Required(),
			mcp.Description("This is the id of the application."),
		),
	)

	authorizedAPIListToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appId := req.Params.Arguments["app_id"].(string)

		resp, err := client.Application.GetAuthorizedAPIs(ctx, appId)
		if err != nil {
			log.Printf("Error listing authorized APIs: %v", err)
			return nil, err
		}
		authorizedAPIs := []interface{}{}
		for _, api := range *resp {
			authorizedScopes := []interface{}{}
			for _, scope := range *api.AuthorizedScopes {
				authorizedScopes = append(authorizedScopes, map[string]interface{}{
					"id":           *scope.Id,
					"name":         *scope.Name,
					"display_name": *scope.DisplayName,
				})
			}
			authorizedAPIs = append(authorizedAPIs, map[string]interface{}{
				"id":                *api.Id,
				"identifier":        *api.Identifier,
				"display_name":      *api.DisplayName,
				"policy_id":         *api.PolicyId,
				"type":              *api.Type,
				"authorized_scopes": authorizedScopes,
			})
		}

		jsonData, err := utils.MarshalResponse(authorizedAPIs)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return authorizedAPIListTool, authorizedAPIListToolImpl
}

func GetUpdateLoginFlowTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	updateLoginFlowTool := mcp.NewTool("update_login_flow",
		mcp.WithDescription("Update login flow in an application for given user prompt."),
		mcp.WithString("app_id",
			mcp.Required(),
			mcp.Description(
				"This is the id of the application for which the login flow is updated.",
			),
		),
		mcp.WithString("user_prompt",
			mcp.Required(),
			mcp.Description(
				"This is the user prompt for the login flow. "+
					"Eg: \"Username and password as first factor and Email OTP as second factor\"",
			),
		),
	)

	updateLoginFlowToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userPrompt := req.Params.Arguments["user_prompt"].(string)
		appId := req.Params.Arguments["app_id"].(string)

		loginFlowResponse, err := client.Application.GenerateLoginFlow(ctx, userPrompt)
		if err != nil {
			log.Printf("Error generating login flow: %v", err)
			return nil, err
		}
		flowId := loginFlowResponse.OperationId
		var statusResponse *application.LoginFlowStatusResponseModel
		for {
			statusResponse, err = client.Application.GetLoginFlowGenerationStatus(ctx, *flowId)
			if err != nil {
				log.Printf("Error getting login flow generation status: %v", err)
				return nil, err
			}
			if statusResponse.Status != nil {
				allTrue := true
				for _, v := range *statusResponse.Status {
					if v != true {
						allTrue = false
						break
					}
				}
				if allTrue {
					log.Printf("Login flow generation completed successfully.")
					break
				}
			}
			log.Printf("Login flow generation in progress. Retrying...")
			time.Sleep(2 * time.Second)
		}
		resultResponse, err := client.Application.GetLoginFlowGenerationResult(ctx, *flowId)
		if err != nil {
			log.Printf("Error getting login flow generation result: %v", err)
			return nil, err
		}
		loginFlowResultData := *resultResponse.Data
		err = client.Application.UpdateLoginFlow(ctx, appId, loginFlowResultData)

		if err != nil {
			log.Printf("Error updating login flow: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText("Login flow generated sucessfully."), nil
	}

	return updateLoginFlowTool, updateLoginFlowToolImpl
}

func convertToStringSlice(input interface{}) []string {
	inputSlice := input.([]interface{})
	result := make([]string, len(inputSlice))
	for i, v := range inputSlice {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}
