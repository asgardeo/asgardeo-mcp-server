package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shashimalcse/go-asgardeo/management"
	"github.com/thilinashashimalsenarath/asgardeo-mcp/internal/asgardeo"
	"github.com/thilinashashimalsenarath/asgardeo-mcp/internal/config"
)

// setupServer configures the MCP server and registers tools.
func setupServer(client *management.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"Asgardeo Management MCP",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// Register ListAsgardeoApplication tool
	appList := mcp.NewTool("ListAsgardeoApplication",
		mcp.WithDescription("List Asgardeo Applications"),
	)
	s.AddTool(appList, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := client.Applications().List(ctx, nil)
		if err != nil {
			log.Printf("Error listing applications: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	})
	return s
}

func main() {
	// Load and validate configuration
	baseURL, clientID, clientSecret, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()
	// Initialize Asgardeo client
	client, err := asgardeo.NewClient(ctx, baseURL, clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to create Asgardeo client: %v", err)
	}

	// Setup and start MCP server
	s := setupServer(client)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
