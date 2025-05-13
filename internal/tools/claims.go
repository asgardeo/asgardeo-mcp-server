/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package tools

import (
	"context"
	"log"

	"github.com/asgardeo/go/pkg/claim"
	"github.com/asgardeo/mcp/internal/asgardeo"
	"github.com/asgardeo/mcp/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetListClaimsTool() (mcp.Tool, server.ToolHandlerFunc) {
	client, err := asgardeo.GetClientInstance(context.Background())
	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	listClaimsTool := mcp.NewTool("list_claims",
		mcp.WithDescription("List all claims in Asgardeo"),
	)

	listClaimsToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		excludeHiddenClaims := true
		listLocalClaimParams := claim.LocalClaimListParamsModel{
			ExcludeHiddenClaims: &excludeHiddenClaims,
		}
		resp, err := client.Claim.ListLocalClaims(ctx, &listLocalClaimParams)
		if err != nil {
			log.Printf("Error listing claims: %v", err)
			return nil, err
		}
		jsonData, err := utils.MarshalResponse(resp)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(jsonData), nil
	}
	return listClaimsTool, listClaimsToolImpl
}
