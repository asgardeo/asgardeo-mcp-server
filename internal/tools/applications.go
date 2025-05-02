package tools

import (
	"context"
	"fmt"
	"log"

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
