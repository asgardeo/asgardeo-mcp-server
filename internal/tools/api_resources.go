package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shashimalcse/go-asgardeo/management"
	"github.com/thilinashashimalsenarath/asgardeo-mcp/internal/asgardeo"
)

func GetApiResourceCreateTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	createApiResourceTool := mcp.NewTool("createAsgardeoApiResource",
		mcp.WithDescription("Create an Asgardeo API Resource"),

		mcp.WithString("identifier",
			mcp.Required(),
			mcp.Description("This is the unique identifier for the API resource."),
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
			mcp.DefaultArray([]management.ScopeCreationModel{}),
			mcp.Description("This is the list of scopes for the API resource."),
		),
	)

	var createApiResourceToolImpl server.ToolHandlerFunc

	createApiResourceToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		name := req.Params.Arguments["name"].(string)
		identifier := req.Params.Arguments["identifier"].(string)
		rawScopes := req.Params.Arguments["scopes"].([]interface{})
		scopes := make([]management.ScopeCreationModel, len(rawScopes))
		for i, rawScope := range rawScopes {
			scopeMap := rawScope.(map[string]interface{})
			scopes[i] = management.ScopeCreationModel{
				Name:        scopeMap["Name"].(string),
				Description: scopeMap["Description"].(string),
			}
		}
		requiresAuthorization := req.Params.Arguments["requiresAuthorization"].(bool)

		newApiResource := management.APIResourceCreateInput{
			Name:                  name,
			Identifier:            identifier,
			Scopes:                scopes,
			RequiresAuthorization: requiresAuthorization,
		}

		resp, err := client.APIResources().Create(ctx, newApiResource)
		if err != nil {
			log.Printf("Error creating an Asgardeo api resource: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return createApiResourceTool, createApiResourceToolImpl
}

func GetApiResourceListTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	listApiResourceTool := mcp.NewTool("listAsgardeoApiResources",
		mcp.WithDescription("List all Asgardeo API Resources"),
	)

	var listApiResourceToolImpl server.ToolHandlerFunc

	listApiResourceToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		resp, err := client.APIResources().List(ctx)
		if err != nil {
			log.Printf("Error creating an Asgardeo application: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}

	return listApiResourceTool, listApiResourceToolImpl
}

func GetApiResourcePatchTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	patchApiResourceTool := mcp.NewTool("patchAsgardeoApiResource",
		mcp.WithDescription("Patch an Asgardeo API Resource"),

		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("This is the ID of the API resource to be patched."),
		),
		mcp.WithString("name",
			mcp.Description("This is the name of the API resource."),
		),
		mcp.WithString("description",
			mcp.Description("This is the description of the API resource."),
		),
		mcp.WithArray("addedScopes",
			mcp.DefaultArray([]management.ScopeCreationModel{}),
			mcp.Description("This is the list of scopes to be added to the API resource."),
		),
		mcp.WithArray("removedScopes",
			mcp.DefaultArray([]string{}),
			mcp.Description("This is the list of scopes to be removed from the API resource."),
		),
	)

	var patchApiResourceToolImpl server.ToolHandlerFunc

	patchApiResourceToolImpl = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id := req.Params.Arguments["id"].(string)
		name := req.Params.Arguments["name"].(string)
		description := req.Params.Arguments["description"].(string)
		rawAddedScopes := req.Params.Arguments["addedScopes"].([]interface{})
		addedScopes := make([]management.ScopeCreationModel, len(rawAddedScopes))
		for i, rawScope := range rawAddedScopes {
			scopeMap := rawScope.(map[string]interface{})
			addedScopes[i] = management.ScopeCreationModel{
				Name:        scopeMap["Name"].(string),
				Description: scopeMap["Description"].(string),
			}
		}
		rawRemovedScopes := req.Params.Arguments["removedScopes"].([]interface{})
		var removedScopes []string
		for _, rawScope := range rawRemovedScopes {
			scopeMap := rawScope.(map[string]interface{})
			name := scopeMap["Name"].(string)
			if name != "" {
				removedScopes = append(removedScopes, name)
			}
		}
		patchApiResource := management.APIResourcePatchModel{
			Name:          &name,
			Description:   &description,
			AddedScopes:   addedScopes,
			RemovedScopes: removedScopes,
		}
		resp := client.APIResources().Patch(ctx, id, patchApiResource)
		if err != nil {
			log.Printf("Error updating Asgardeo api resource: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}
	return patchApiResourceTool, patchApiResourceToolImpl
}
