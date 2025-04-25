# Asgardeo MCP Server
MCP server to interact with your Asgardeo organization through LLM tools

## How to use

### On Asgardeo

1. Create an M2M application in your Asgardeo organization.
2. Authorize the management APIs you want the MCP server to consume, to the created M2M application.
3. Copy the client ID, and client secret of the M2M application.

### On your machine

4. Clone the repo.
5. Build the repo. This will create an executable named `asgardeo-mcp`.

```bash
go build
```

6. Configure your MCP client.

- Claude Desktop

  - Open Claude Desktop.
  - Click on Claude > Settings.
  - Switch to `Developer` tab.
  - Click on `Edit Config` button at the bottom. This will point to `claude_desktop_config.json` file in the file explorer.
  - Open the `claude_desktop_config.json` file in a code editor and in the `mcpServers` object, add the following.

    ```js
    "asgardeo-mcp": {
        "command": "<absolute path to the asgardeo-mcp executable>",
        "args": [],
        "env": {
            "ASGARDEO_BASE_URL" : "https://api.asgardeo.io/t/<asgardeo organization>",
            "ASGARDEO_CLIENT_ID" : "<client ID>",
            "ASGARDEO_CLIENT_SECRET" : "<client secret>"
        }
    }
    ```

  - Restart Claude Desktop.

- Cursor

  - Open Cursor.
  - Click on Cursor > Settings > Cursor Settings.
  - Switch to `MCP` tab.
  - Click on `Add new global MCP server` button at the bottom. This will open `mcp.json` file in the editor itself.
  - In the `mcpServers` object, add the following.

    ```js
    "asgardeo-mcp": {
        "command": "<absolute path to the asgardeo-mcp executable>",
        "args": [],
        "env": {
            "ASGARDEO_BASE_URL" : "https://api.asgardeo.io/t/<asgardeo organization>",
            "ASGARDEO_CLIENT_ID" : "<client ID>",
            "ASGARDEO_CLIENT_SECRET" : "<client secret>"
        }
    }
    ```
