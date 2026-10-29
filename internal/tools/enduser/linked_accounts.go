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

func linkedAccountTools(c *selfservice.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("list_linked_accounts",
			"Lists the signed-in user's linked local accounts and linked social / enterprise logins.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				local, err := c.Get(ctx, meV1+"/associations", nil)
				if err != nil {
					return nil, err
				}
				social, err := c.Get(ctx, meV1+"/federated-associations", nil)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"localAccounts": asList(local), "socialAccounts": asList(social)}, nil
			},
		),
		newTool("unlink_local_account",
			"Unlinks one linked local account.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "user_id")
				if err != nil {
					return nil, err
				}
				return c.Delete(ctx, meV1+"/associations/"+url.PathEscape(id))
			},
			mcp.WithString("user_id", mcp.Required(), mcp.Description("userId of the local account from list_linked_accounts.")),
		),
		newTool("unlink_social_account",
			"Unlinks one social / enterprise login.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "id")
				if err != nil {
					return nil, err
				}
				return c.Delete(ctx, meV1+"/federated-associations/"+url.PathEscape(id))
			},
			mcp.WithString("id", mcp.Required(), mcp.Description("Association id (not idp.id) from list_linked_accounts.")),
		),
		newTool("unlink_all_local_accounts",
			"Unlinks every linked local account. List each account that will be unlinked when confirming.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return c.Delete(ctx, meV1+"/associations")
			},
		),
		newTool("unlink_all_social_accounts",
			"Unlinks every social / enterprise login. List each login that will be unlinked when confirming.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				social, err := c.Get(ctx, meV1+"/federated-associations", nil)
				if err != nil {
					return nil, err
				}
				if len(asList(social)) == 0 {
					return map[string]interface{}{"message": "No social logins to unlink."}, nil
				}
				return c.Delete(ctx, meV1+"/federated-associations")
			},
		),
	}
}
