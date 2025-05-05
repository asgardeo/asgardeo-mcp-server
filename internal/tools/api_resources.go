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

	"github.com/asgardeo/go/pkg/api_resource"
	"github.com/asgardeo/mcp/internal/asgardeo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetListAPIResourcesTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())
	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	apiResourceListTool := mcp.NewTool("list_api_resources",
		mcp.WithDescription("List API Resources registered in Asgardeo"),
		mcp.WithString("filter",
			mcp.Description(`Filter expression to apply, e.g., name eq Payments API`),
		),
		mcp.WithString("before",
			mcp.Description(`The before cursor to use for pagination. The API will return results before this cursor`),
		),
		mcp.WithString("after",
			mcp.Description(`The after cursor to use for pagination. The API will return results after this cursor`),
		),
		mcp.WithNumber("limit",
			mcp.Description(`The maximum number of results to return.`),
		),
	)

	var apiResourceListToolImpl server.ToolHandlerFunc
	apiResourceListToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.Params.Arguments
		limit := getOptionalParam[int](args, "limit")
		filter := getOptionalParam[string](args, "filter")
		before := getOptionalParam[string](args, "before")
		after := getOptionalParam[string](args, "after")
		params := api_resource.GetAPIResourcesParams{
			Limit:  limit,
			Filter: filter,
			Before: before,
			After:  after,
		}
		resp, err := client.APIResource.List(ctx, &params)
		if err != nil {
			log.Printf("Error listing api resources: %v", err)
			return nil, err
		}

		api_resources := []interface{}{}
		for _, apiResource := range *resp.APIResources {
			apiResourceMap := map[string]interface{}{
				"id":   apiResource.Id,
				"name": apiResource.Name,
			}
			if apiResource.Type != nil {
				apiResourceMap["type"] = *apiResource.Type
			}
			if apiResource.RequiresAuthorization != nil {
				apiResourceMap["requiresAuthorization"] = *apiResource.RequiresAuthorization
			}
			api_resources = append(api_resources, apiResourceMap)
		}

		return mcp.NewToolResultText(fmt.Sprintf("%+v", api_resources)), nil
	}

	return apiResourceListTool, apiResourceListToolImpl
}

func GetSearchAPIResourcesByNameTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())
	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}
	apiResourceGetByNameTool := mcp.NewTool("list_api_resources_by_name",
		mcp.WithDescription("List API Resources registered in Asgardeo by name"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("This is the name of the API resource."),
		),
	)
	var apiResourceListByNameToolImpl server.ToolHandlerFunc
	apiResourceListByNameToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.Params.Arguments["name"].(string)
		resp, err := client.APIResource.GetByName(ctx, name)
		if err != nil {
			log.Printf("Error getting api resource list by name: %v", err)
			return nil, err
		}
		api_resources := []interface{}{}
		for _, apiResource := range *resp {
			apiResourceMap := map[string]interface{}{
				"id":   apiResource.Id,
				"name": apiResource.Name,
			}
			if apiResource.Type != nil {
				apiResourceMap["type"] = *apiResource.Type
			}
			if apiResource.RequiresAuthorization != nil {
				apiResourceMap["requiresAuthorization"] = *apiResource.RequiresAuthorization
			}
			api_resources = append(api_resources, apiResourceMap)
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", api_resources)), nil
	}
	return apiResourceGetByNameTool, apiResourceListByNameToolImpl
}

func GetSearchAPIResourceByIdentifierTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())
	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}
	apiResourceGetByIdentifierTool := mcp.NewTool("get_api_resource_by_identifier",
		mcp.WithDescription("Get API Resource registered in Asgardeo by identifier"),
		mcp.WithString("identifier",
			mcp.Required(),
			mcp.Description("This is the identifier of the API resource."),
		),
	)
	var apiResourceGetByIdentifierToolImpl server.ToolHandlerFunc
	apiResourceGetByIdentifierToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		identifier := req.Params.Arguments["identifier"].(string)
		resp, err := client.APIResource.GetByIdentifier(ctx, identifier)
		if err != nil {
			log.Printf("Error getting api resource by identifier: %v", err)
			return nil, err
		}
		apiResourceMap := map[string]interface{}{
			"id":   resp.Id,
			"name": resp.Name,
		}
		if resp.Type != nil {
			apiResourceMap["type"] = *resp.Type
		}
		if resp.RequiresAuthorization != nil {
			apiResourceMap["requiresAuthorization"] = *resp.RequiresAuthorization
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", apiResourceMap)), nil
	}
	return apiResourceGetByIdentifierTool, apiResourceGetByIdentifierToolImpl
}

func GetCreateAPIResourceTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	apiResourceCreateTool := mcp.NewTool("create_api_resource",
		mcp.WithDescription("Create an API Resource in Asgardeo"),

		mcp.WithString("identifier",
			mcp.Required(),
			mcp.Description("This is the identifier for the API resource."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("This is the name of the API resource."),
		),
		mcp.WithBoolean("requiresAuthorization",
			mcp.Required(),
			mcp.DefaultBool(true),
			mcp.Description("This indicates whether the API resource requires authorization."),
		),
		mcp.WithArray("scopes",
			mcp.Required(),
			mcp.DefaultArray([]api_resource.ScopeCreationModel{}),
			mcp.Description("This is the list of scopes for the API resource."),
		),
	)

	var apiResourceCreateToolImpl server.ToolHandlerFunc
	apiResourceCreateToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.Params.Arguments["name"].(string)
		identifier := req.Params.Arguments["identifier"].(string)
		inputScopes := req.Params.Arguments["scopes"].([]interface{})
		scopes := make([]api_resource.ScopeCreationModel, len(inputScopes))
		for i, inputScope := range inputScopes {
			scope := api_resource.ScopeCreationModel{}

			switch scopeData := inputScope.(type) {
			case string:
				// Simplified form: only scope name provided.
				scope.Name = scopeData

			case map[string]interface{}:
				// Structured form: detailed fields provided.
				name, ok := scopeData["name"].(string)
				if !ok {
					return nil, fmt.Errorf("scope Name is required and must be a string at index %d", i)
				}
				scope.Name = name

				if displayName, ok := scopeData["displayName"].(string); ok {
					scope.DisplayName = &displayName
				}

				if description, ok := scopeData["description"].(string); ok {
					scope.Description = &description
				}

			default:
				return nil, fmt.Errorf("unexpected scope format at index %d", i)
			}

			scopes[i] = scope
		}

		requiresAuthorization := req.Params.Arguments["requiresAuthorization"].(bool)
		newApiResource := api_resource.APIResourceCreationModel{
			Name:                  name,
			Identifier:            identifier,
			Scopes:                &scopes,
			RequiresAuthorization: &requiresAuthorization,
		}

		resp, err := client.APIResource.Create(ctx, &newApiResource)
		if err != nil {
			log.Printf("Error while creating API resource: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}
	return apiResourceCreateTool, apiResourceCreateToolImpl
}

func getOptionalParam[T any](args map[string]interface{}, key string) *T {
	if val, ok := args[key]; ok {
		if typedVal, ok := val.(T); ok {
			return &typedVal
		}
	}
	return nil
}
