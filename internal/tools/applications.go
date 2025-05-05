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

	"github.com/asgardeo/go/pkg/application"
	"github.com/asgardeo/mcp/internal/asgardeo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetApplicationListTool() (mcp.Tool, server.ToolHandlerFunc) {
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
				"Name": appName,
				"ID":   appID,
			})
		}

		return mcp.NewToolResultText(fmt.Sprintf("%+v", apps)), nil
	}

	return appListTool, appListToolImpl
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
		authorizedAPI := application.AddAuthorizedAPIJSONRequestBody{
			Id:               &id,
			PolicyIdentifier: &policyIdentifier,
			Scopes:           &scopes,
		}

		resp, err := client.Application.AuthorizeAPI(ctx, appId, authorizedAPI)
		if err != nil {
			log.Printf("Error authorizing API resource: %v", err)
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return authorizeAPITool, authorizeAPIToolImpl
}
