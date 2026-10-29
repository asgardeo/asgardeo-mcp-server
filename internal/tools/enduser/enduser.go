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

// Package enduser holds the Self Service MCP tools, which act as the signed-in end user.
package enduser

import (
	"context"
	"fmt"
	"strings"

	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/asgardeo/mcp/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tools returns every self-service tool.
func Tools(c *selfservice.Client) []server.ServerTool {
	var tools []server.ServerTool
	tools = append(tools, profileTools(c)...)
	tools = append(tools, sessionTools(c)...)
	tools = append(tools, securityTools(c)...)
	tools = append(tools, linkedAccountTools(c)...)
	tools = append(tools, appTools(c)...)
	tools = append(tools, consentTools(c)...)
	return tools
}

type kind int

const (
	read kind = iota
	write
	destructive
)

type handlerFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// newTool builds a tool that returns the redacted handler result as JSON.
func newTool(name, description string, k kind, handler handlerFunc, opts ...mcp.ToolOption) server.ServerTool {
	if k == write || k == destructive {
		description += " Before calling, describe the exact change and its target to the user and get their explicit yes to it. " +
			"The user's request is not a yes, and a yes to an earlier or different change doesn't carry over."
	}
	tool := mcp.NewTool(name, append([]mcp.ToolOption{
		mcp.WithDescription(description),
		mcp.WithReadOnlyHintAnnotation(k == read),
		mcp.WithDestructiveHintAnnotation(k == destructive),
		mcp.WithOpenWorldHintAnnotation(true),
	}, opts...)...)

	return server.ServerTool{
		Tool: tool,
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := handler(ctx, req.GetArguments())
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if result == nil {
				return mcp.NewToolResultText("Done."), nil
			}
			text, err := utils.MarshalResponse(redact(result))
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(text), nil
		},
	}
}

// sensitiveKeys are response fields never returned to the agent.
var sensitiveKeys = map[string]bool{
	"qrcodeurl": true, "backupcodes": true, "access_token": true, "refresh_token": true, "id_token": true,
	"token": true, "devicetoken": true, "challenge": true, "publickey": true, "publickeycose": true,
	"attestationobject": true, "clientdatajson": true, "signature": true,
}

func redact(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, item := range val {
			lower := strings.ToLower(k)
			if sensitiveKeys[lower] || strings.Contains(lower, "secret") || strings.Contains(lower, "password") {
				continue
			}
			out[k] = redact(item)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = redact(item)
		}
		return out
	default:
		return v
	}
}

func str(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func requiredStr(args map[string]interface{}, key string) (string, error) {
	v := str(args, key)
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

func asMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func asList(v interface{}) []interface{} {
	l, _ := v.([]interface{})
	if l == nil {
		return []interface{}{}
	}
	return l
}
