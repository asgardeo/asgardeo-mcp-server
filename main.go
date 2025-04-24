package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/thilinashashimalsenarath/asgardeo-mcp/internal/tools"
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

	appListTool, appListToolImpl := tools.GetApplicationListTool()
	s.AddTool(appListTool, appListToolImpl)

	appDetailFetchTool, appDetailFetchToolImpl := tools.GetApplicationDetailTool()
	s.AddTool(appDetailFetchTool, appDetailFetchToolImpl)

	appCreateTool, appCreateToolImpl := tools.GetApplicationCreateTool()
	s.AddTool(appCreateTool, appCreateToolImpl)

	apiResourceCreateTool, apiResourceCreateToolImpl := tools.GetApiResourceCreateTool()
	s.AddTool(apiResourceCreateTool, apiResourceCreateToolImpl)

	apiResourceListTool, apiResourceListToolImpl := tools.GetApiResourceListTool()
	s.AddTool(apiResourceListTool, apiResourceListToolImpl)

	apiResourcePatchTool, apiResourcePatchToolImpl := tools.GetApiResourcePatchTool()
	s.AddTool(apiResourcePatchTool, apiResourcePatchToolImpl)

	return s
}

func main() {
	// Setup and start MCP server
	s := setupServer()
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
