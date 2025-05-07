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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/asgardeo/go/pkg/application"
	"github.com/asgardeo/mcp/internal/asgardeo"
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
				"name":          spa.Name,
				"id":            spa.Id,
				"client_id":     spa.ClientId,
				"redirect_url":  spa.RedirectURL,
				"scope":         spa.AuthorizedScopes,
				"response_type": "code",
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := marshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return spaTool, spaToolImpl
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
				"name":         mobileApp.Name,
				"id":           mobileApp.Id,
				"client_id":    mobileApp.ClientId,
				"redirect_url": mobileApp.RedirectURL,
				"scope":        mobileApp.AuthorizedScopes,
			},
			"oauth_endpoints": map[string]string{
				"base_url":      baseURL,
				"authorize_url": fmt.Sprintf("%s/oauth2/authorize", baseURL),
				"token_url":     fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":      fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url":  fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := marshalResponse(response)
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
				"name":          m2mApp.Name,
				"id":            m2mApp.Id,
				"client_id":     m2mApp.ClientId,
				"client_secret": m2mApp.ClientSecret,
			},
			"oauth_endpoints": map[string]string{
				"token_url":    fmt.Sprintf("%s/oauth2/token", baseURL),
				"jwks_url":     fmt.Sprintf("%s/oauth2/jwks", baseURL),
				"userinfo_url": fmt.Sprintf("%s/oauth2/userinfo", baseURL),
			},
		}

		jsonData, err := marshalResponse(response)
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

		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":          app.Name,
				"id":            app.Id,
				"client_id":     app.ClientId,
				"client_secret": app.ClientSecret,
			},
		}

		jsonData, err := marshalResponse(response)
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

		response := map[string]interface{}{
			"application_configurations": map[string]string{
				"name":          app.Name,
				"id":            app.Id,
				"client_id":     app.ClientId,
				"client_secret": app.ClientSecret,
			},
		}

		jsonData, err := marshalResponse(response)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return getApplicationByClientIDTool, getApplicationByClientIDToolImpl
}

func GetAuthorizeAPITool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

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
		),
	)
	var authorizeAPIToolImpl server.ToolHandlerFunc
	authorizeAPIToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func GetGenerateLoginFlowTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	generateLoginFlowTool := mcp.NewTool("generate_login_flow",
		mcp.WithDescription("Generate login flow for an application for given user prompt."),
		mcp.WithString("user_prompt",
			mcp.Required(),
			mcp.Description(
				"This is the user prompt for the login flow generation. "+
					"Eg: \"Username and password as first factor and Email OTP as second factor\"",
			),
		),
	)

	generateLoginFlowToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userPrompt := req.Params.Arguments["user_prompt"].(string)

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
		jsonData, err := marshalResponse(resultResponse)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}

	return generateLoginFlowTool, generateLoginFlowToolImpl
}

func marshalResponse(response interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return "", err
	}
	return string(jsonData), nil
}
