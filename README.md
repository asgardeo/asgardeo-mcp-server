# Asgardeo MCP Server

## How to use

### On Asgardeo

1. Create an M2M application in your Asgardeo organization.
2. Authorize the management APIs you want the MCP server to consume, to the created M2M application.
3. Copy the client ID, and client secret of the M2M application.

### On your machine

4. Clone the repo.
5. Build the repo. This will create an executable named asgardeo-mcp.

```bash
go build
```

6. Configure your MCP client.

```js
"asgardeo-mcp": {
    "command": "<absolute path to the asgardeo-mcp executable>",
    "args": [],
    "env": {
        "ASGARDEO_BASE_URL" : "https://api.asgardeo.io/t/<asgardeo base URL>",
        "ASGARDEO_CLIENT_ID" : "<client ID>",
        "ASGARDEO_CLIENT_SECRET" : "<client secret>"
    }
}
```