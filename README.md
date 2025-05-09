# Asgardeo MCP Server
MCP server to interact with your Asgardeo organization through LLM tools

## How to use

### On Asgardeo

1. Create an M2M application in your Asgardeo organization.
2. Authorize following management APIs

  | API | Scopes |
|-----|--------|
| Application Management API (`/api/server/v1/applications`) | `internal_application_mgt_view` `internal_application_mgt_update` `internal_application_mgt_create` |
| API Resource Management API (`/api/server/v1/api-resources`) | `internal_api_resource_update` `internal_api_resource_create` `internal_api_resource_view` |
| Identity Provider Management API (`/api/server/v1/identity-providers`) | `internal_idp_view` |
| Authenticators Management API (`/api/server/v1/authenticators`) | `internal_authenticator_view` |
| Claim Management API (`/api/server/v1/claim-dialects`) | `internal_claim_meta_view` |
| SCIM2 Users API (`/scim2/Users`) | `internal_user_mgt_create` |

3. Copy the client ID, and client secret of the M2M application.

### On your machine

4. Clone the repo.
5. Install dependencies.

```
go mod tidy
```

6. Build the repo. This will create an executable named `asgardeo-mcp`.

```bash
go build -o asgardeo-mcp
```

7. Configure your MCP client.

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

8. Try out.

    You can try following operations with the tool list available.
    - List applications
    - Create a Single Page App, Mobile App or M2M App
    - Update application basic info or OAuth configs
    - Add an API resource
    - Authorize API resource(s) to the application
    - Configure the login flow with a prompt to your app (limited capabilities only)
    - Create a test user