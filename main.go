package main

import (
	"log"

	"github.com/asgardeo/mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// setupServer configures the MCP server and registers tools.
func setupServer() *server.MCPServer {
	s := server.NewMCPServer(
		"Asgardeo Management MCP",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	appListTool, appListToolImpl := tools.GetApplicationListTool()
	s.AddTool(appListTool, appListToolImpl)

	return s
}

func main() {
	// Setup and start MCP server
	s := setupServer()
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
