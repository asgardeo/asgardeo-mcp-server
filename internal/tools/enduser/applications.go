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

	"github.com/asgardeo/go/pkg/sdk"
	"github.com/asgardeo/go/pkg/self_service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func appTools(c *sdk.Client, clientID string) []server.ServerTool {
	return []server.ServerTool{
		newTool("list_my_applications",
			"Lists the applications the signed-in user can access, with their access URLs.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				params := self_service.ApplicationListParamsModel{}
				if filter := str(args, "filter"); filter != "" {
					params.Filter = &filter
				}
				if v, ok := args["limit"].(float64); ok {
					limit := int32(v)
					params.Limit = &limit
				}
				if v, ok := args["offset"].(float64); ok {
					offset := int32(v)
					params.Offset = &offset
				}
				return c.SelfService.ListApplications(ctx, &params)
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
				return c.SelfService.GetApplication(ctx, id)
			},
			mcp.WithString("application_id", mcp.Required(), mcp.Description("Application id from list_my_applications.")),
		),
		newTool("list_authorized_apps",
			"Lists the apps the signed-in user has authorized to access their account, with the approved scopes. "+
				"The app marked currentApp is the one this assistant uses.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				resp, err := c.SelfService.ListAuthorizedApps(ctx)
				if err != nil {
					return nil, err
				}
				apps := make([]authorizedApp, 0, len(resp))
				for _, app := range resp {
					apps = append(apps, authorizedApp{AuthorizedAppModel: app, CurrentApp: isCurrentApp(clientID, app)})
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
				app, err := c.SelfService.GetAuthorizedApp(ctx, id)
				if err != nil {
					return nil, err
				}
				if app != nil && isCurrentApp(clientID, *app) {
					return nil, fmt.Errorf("this is the app this assistant uses, so its access can't be revoked here")
				}
				return nil, c.SelfService.RevokeAuthorizedApp(ctx, id)
			},
			mcp.WithString("application_id", mcp.Required(), mcp.Description("App id from list_authorized_apps.")),
		),
	}
}

// authorizedApp marks the app this assistant uses.
type authorizedApp struct {
	self_service.AuthorizedAppModel
	CurrentApp bool `json:"currentApp,omitempty"`
}

func isCurrentApp(clientID string, app self_service.AuthorizedAppModel) bool {
	return clientID != "" && app.ClientId == clientID
}
