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
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	appListTool, appListToolImpl := tools.GetListApplicationsTool()
	s.AddTool(appListTool, appListToolImpl)

	spaTool, spaToolImpl := tools.GetCreateSinglePageAppTool()
	s.AddTool(spaTool, spaToolImpl)

	mobileAppTool, mobileAppToolImpl := tools.GetCreateMobileAppTool()
	s.AddTool(mobileAppTool, mobileAppToolImpl)

	m2mAppTool, m2mAppToolImpl := tools.GetCreateM2MAppTool()
	s.AddTool(m2mAppTool, m2mAppToolImpl)

	getAppByNameTool, getAppByNameToolmpl := tools.GetSearchApplicationByNameTool()
	s.AddTool(getAppByNameTool, getAppByNameToolmpl)

	getAppByClientIdTool, getAppByClientIdToolmpl := tools.GetSearchApplicationByClientIdTool()
	s.AddTool(getAppByClientIdTool, getAppByClientIdToolmpl)

	return s
}

func main() {
	// Setup and start MCP server
	s := setupServer()
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
