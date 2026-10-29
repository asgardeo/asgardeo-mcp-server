/*
 * Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package enduser

import (
	"context"
	"net/url"

	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const meV1 = "/api/users/v1/me"

func sessionTools(c *selfservice.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("list_sessions",
			"Lists where the signed-in user is signed in: browser and OS (userAgent), IP, login and last access time (epoch milliseconds) and the apps used in each session.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return c.Get(ctx, meV1+"/sessions", nil)
			},
		),
		newTool("terminate_session",
			"Signs the user out of one session, identified by its id from list_sessions. "+
				"If it is the session this assistant uses, the assistant loses access and the user has to sign in again.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "session_id")
				if err != nil {
					return nil, err
				}
				return c.Delete(ctx, meV1+"/sessions/"+url.PathEscape(id))
			},
			mcp.WithString("session_id", mcp.Required(), mcp.Description("Session id from list_sessions.")),
		),
		newTool("terminate_all_sessions",
			"Signs the user out of every session on every device, possibly including the current one.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return c.Delete(ctx, meV1+"/sessions")
			},
		),
	}
}
