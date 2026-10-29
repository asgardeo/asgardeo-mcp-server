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
	"fmt"
	"net/url"
	"strconv"

	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func appTools(c *selfservice.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("list_my_applications",
			"Lists the applications the signed-in user can access, with their access URLs.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				query := url.Values{}
				if filter := str(args, "filter"); filter != "" {
					query.Set("filter", filter)
				}
				for _, key := range []string{"limit", "offset"} {
					if v, ok := args[key].(float64); ok {
						query.Set(key, strconv.Itoa(int(v)))
					}
				}
				return c.Get(ctx, meV1+"/applications", query)
			},
			mcp.WithString("filter", mcp.Description("Filter such as 'name co travel' or 'name eq Pickup'.")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of results.")),
			mcp.WithNumber("offset", mcp.Description("Number of results to skip.")),
		),
		newTool("get_my_application",
			"Gets one application the signed-in user can access.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "application_id")
				if err != nil {
					return nil, err
				}
				return c.Get(ctx, meV1+"/applications/"+url.PathEscape(id), nil)
			},
			mcp.WithString("application_id", mcp.Required(), mcp.Description("Application id from list_my_applications.")),
		),
		newTool("list_authorized_apps",
			"Lists the apps the signed-in user has authorized to access their account, with the approved scopes. "+
				"The app marked currentApp is the one this assistant uses.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				resp, err := c.Get(ctx, meV2+"/authorized-apps", nil)
				if err != nil {
					return nil, err
				}
				apps := asList(resp)
				for _, item := range apps {
					if app := asMap(item); app != nil && isCurrentApp(c, app) {
						app["currentApp"] = true
					}
				}
				return apps, nil
			},
		),
		newTool("revoke_app_access",
			"Revokes an app's access to the user's account; the app's tokens stop working and it must ask for permission again. "+
				"Data the app already collected is governed by the app's own policy. "+
				"If it is the app this assistant uses, the assistant loses access and the user has to sign in again.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "application_id")
				if err != nil {
					return nil, err
				}
				path := meV2 + "/authorized-apps/" + url.PathEscape(id)
				app, err := c.Get(ctx, path, nil)
				if err != nil {
					return nil, err
				}
				if isCurrentApp(c, asMap(app)) {
					return nil, fmt.Errorf("this is the app this assistant uses, so its access can't be revoked here")
				}
				return c.Delete(ctx, path)
			},
			mcp.WithString("application_id", mcp.Required(), mcp.Description("App id from list_authorized_apps.")),
		),
	}
}

func isCurrentApp(c *selfservice.Client, app map[string]interface{}) bool {
	clientID, _ := app["clientId"].(string)
	return c.ClientID() != "" && clientID == c.ClientID()
}
