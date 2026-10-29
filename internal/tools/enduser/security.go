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
	"net/http"
	"net/url"
	"strings"

	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const meV2 = "/api/users/v2/me"

const loadFailed = "could not be loaded"

func securityTools(c *selfservice.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("get_security_overview",
			"Summarises the signed-in user's sign-in security: authenticator app on/off, enabled MFA options, remaining backup codes, passkeys and push devices.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return securityOverview(ctx, c)
			},
		),
		newTool("rename_passkey",
			"Renames a passkey or security key.",
			write,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "credential_id")
				if err != nil {
					return nil, err
				}
				name, err := requiredStr(args, "display_name")
				if err != nil {
					return nil, err
				}
				body := []map[string]string{{"operation": "REPLACE", "path": "/displayName", "value": name}}
				_, err = c.Do(ctx, http.MethodPatch, meV2+"/webauthn/"+url.PathEscape(id), nil, body, "")
				return nil, err
			},
			mcp.WithString("credential_id", mcp.Required(), mcp.Description("Passkey credentialId from get_security_overview.")),
			mcp.WithString("display_name", mcp.Required(), mcp.Description("New name.")),
		),
		newTool("remove_passkey",
			"Removes a passkey or security key; it can no longer be used to sign in. Warn the user if it is their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "credential_id")
				if err != nil {
					return nil, err
				}
				return c.Delete(ctx, meV2+"/webauthn/"+url.PathEscape(id))
			},
			mcp.WithString("credential_id", mcp.Required(), mcp.Description("Passkey credentialId from get_security_overview.")),
		),
		newTool("remove_push_device",
			"Removes a device that receives push sign-in approvals. Warn the user if it is their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "device_id")
				if err != nil {
					return nil, err
				}
				return c.Delete(ctx, meV1+"/push/devices/"+url.PathEscape(id))
			},
			mcp.WithString("device_id", mcp.Required(), mcp.Description("deviceId from get_security_overview.")),
		),
		newTool("remove_authenticator_app",
			"Removes the user's authenticator app (TOTP); codes from the app stop working. Warn the user if it is their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				_, err := c.Delete(ctx, meV1+"/totp")
				return nil, err
			},
		),
		newTool("delete_backup_codes",
			"Deletes the user's remaining backup codes; they stop working. Warn the user if they are their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return c.Delete(ctx, meV1+"/backup-codes")
			},
		),
	}
}

func securityOverview(ctx context.Context, c *selfservice.Client) (interface{}, error) {
	out := map[string]interface{}{}

	// fetch skips unsupported features (404, 405) and marks failed ones. 401 and 403 stop the overview.
	var authErr error
	fetch := func(path string, query url.Values, keys ...string) (interface{}, bool) {
		if authErr != nil {
			return nil, false
		}
		resp, err := c.Get(ctx, path, query)
		if err != nil {
			switch {
			case selfservice.IsStatus(err, http.StatusUnauthorized, http.StatusForbidden):
				authErr = err
			case selfservice.IsStatus(err, http.StatusNotFound, http.StatusMethodNotAllowed):
			default:
				for _, key := range keys {
					out[key] = loadFailed
				}
			}
			return nil, false
		}
		return resp, true
	}

	if me, ok := fetch("/scim2/Me", url.Values{"attributes": {extensionSchema}}, "authenticatorAppEnabled", "backupCodesEnabled"); ok {
		attrs := asMap(asMap(me)[extensionSchema])
		out["authenticatorAppEnabled"] = attrs["totpEnabled"]
		out["backupCodesEnabled"] = attrs["backupCodeEnabled"]
	}
	if mfa, ok := fetch(meV1+"/mfa/authenticators", nil, "enabledMfaOptions"); ok {
		out["enabledMfaOptions"] = asMap(mfa)["enabledAuthenticators"]
	}
	if codes, ok := fetch(meV1+"/backup-codes", nil, "backupCodesRemaining"); ok {
		out["backupCodesRemaining"] = asMap(codes)["remainingBackupCodesCount"]
	}
	if resp, ok := fetch(meV2+"/webauthn", nil, "passkeys"); ok {
		passkeys := []map[string]interface{}{}
		for _, item := range asList(resp) {
			p := asMap(item)
			passkeys = append(passkeys, map[string]interface{}{
				"credentialId":     asMap(p["credential"])["credentialId"],
				"displayName":      p["displayName"],
				"registrationTime": p["registrationTime"],
			})
		}
		out["passkeys"] = passkeys
	}
	if resp, ok := fetch(meV1+"/push/devices", nil, "pushDevices"); ok {
		devices := []map[string]interface{}{}
		for _, item := range asList(resp) {
			d := asMap(item)
			devices = append(devices, map[string]interface{}{"deviceId": d["deviceId"], "name": d["name"], "model": d["model"]})
		}
		out["pushDevices"] = devices
	}

	if authErr != nil {
		return nil, authErr
	}
	if out["enabledMfaOptions"] != loadFailed {
		out["enabledMfaOptions"] = enrolledMfaOptions(out)
	}
	return out, nil
}

func enrolledMfaOptions(out map[string]interface{}) []string {
	options := []string{}
	raw, _ := out["enabledMfaOptions"].(string)
	for _, option := range strings.Split(raw, ",") {
		option = strings.TrimSpace(option)
		switch {
		case option == "":
		case option == "totp" && fmt.Sprint(out["authenticatorAppEnabled"]) == "false":
		case option == "backup-code-authenticator" && out["backupCodesRemaining"] == float64(0):
		default:
			options = append(options, option)
		}
	}
	return options
}
