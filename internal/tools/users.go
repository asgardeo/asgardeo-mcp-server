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
	"fmt"
	"log"

	"github.com/asgardeo/go/pkg/user"
	"github.com/asgardeo/mcp/internal/asgardeo"
	"github.com/asgardeo/mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetCreateUserTool() (mcp.Tool, server.ToolHandlerFunc) {
	productName := config.GetProductName()
	client, err := asgardeo.GetClientInstance(context.Background())

	if err != nil {
		log.Printf("Error initializing client instance: %v", err)
	}

	userCreateTool := mcp.NewTool("create_user",
		mcp.WithDescription(fmt.Sprintf("Create a user in %s", productName)),

		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("This is the username of the user. This should be an email address."),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("This is the password of the user. Eg; atGHL1234#"),
		),
		mcp.WithString("email",
			mcp.Required(),
			mcp.Description("This is the email of the user."),
		),
		mcp.WithString("first_name",
			mcp.Required(),
			mcp.Description("This is the first name of the user."),
		),
		mcp.WithString("last_name",
			mcp.Required(),
			mcp.Description("This is the last name of the user."),
		),
		mcp.WithString("userstore_domain",
			mcp.Description("This is the userstore domain of the user."),
			mcp.DefaultString("DEFAULT"),
		),
	)

	userCreateToolImpl := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		username := req.Params.Arguments["username"].(string)
		password := req.Params.Arguments["password"].(string)
		email := req.Params.Arguments["email"].(string)
		firstName := req.Params.Arguments["first_name"].(string)
		lastName := req.Params.Arguments["last_name"].(string)
		userstoreDomain := "DEFAULT"
		if req.Params.Arguments["userstore_domain"] != nil {
			userstoreDomain = req.Params.Arguments["userstore_domain"].(string)
		}

		user := user.UserCreateModel{
			Username:  userstoreDomain + "/" + username,
			Password:  password,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}
		resp, err := client.User.CreateUser(ctx, user)
		if err != nil {
			log.Printf("Error creating the user: %v", err)
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
	}
	return userCreateTool, userCreateToolImpl
}
