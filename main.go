package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	applications "github.com/thilinashashimalsenarath/asgardeo-mcp/internal/tools"
)

// setupServer configures the MCP server and registers tools.
func setupServer() *server.MCPServer {
	s := server.NewMCPServer(
		"Asgardeo Management MCP",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	appListTool, appListToolImpl := applications.List()
	s.AddTool(appListTool, appListToolImpl)

	appDetailFetchTool, appDetailFetchToolImpl := applications.Get()
	s.AddTool(appDetailFetchTool, appDetailFetchToolImpl)

	appCreateTool, appCreateToolImpl := applications.Create()
	s.AddTool(appCreateTool, appCreateToolImpl)

	return s
}

func main() {
	// Setup and start MCP server
	s := setupServer()
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
