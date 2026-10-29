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
	"strings"

	"github.com/asgardeo/go/pkg/sdk"
	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const loadFailed = "could not be loaded"

func securityTools(c *sdk.Client) []server.ServerTool {
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
				return nil, c.SelfService.RenamePasskey(ctx, id, name)
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
				return nil, c.SelfService.DeletePasskey(ctx, id)
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
				return nil, c.SelfService.DeletePushDevice(ctx, id)
			},
			mcp.WithString("device_id", mcp.Required(), mcp.Description("deviceId from get_security_overview.")),
		),
		newTool("remove_authenticator_app",
			"Removes the user's authenticator app (TOTP); codes from the app stop working. Warn the user if it is their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, c.SelfService.DeleteTOTP(ctx)
			},
		),
		newTool("delete_backup_codes",
			"Deletes the user's remaining backup codes; they stop working. Warn the user if they are their last second factor.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, c.SelfService.DeleteBackupCodes(ctx)
			},
		),
	}
}

func securityOverview(ctx context.Context, c *sdk.Client) (interface{}, error) {
	out := map[string]interface{}{}

	// failed skips unsupported features, marks other failures and returns auth errors.
	failed := func(err error, keys ...string) error {
		switch {
		case selfservice.IsStatus(err, http.StatusUnauthorized, http.StatusForbidden):
			return err
		case selfservice.IsStatus(err, http.StatusNotFound, http.StatusMethodNotAllowed):
		default:
			for _, key := range keys {
				out[key] = loadFailed
			}
		}
		return nil
	}

	if me, err := c.User.GetMe(ctx, extensionSchema); err != nil {
		if err := failed(err, "authenticatorAppEnabled", "backupCodesEnabled"); err != nil {
			return nil, err
		}
	} else {
		attrs := asMap(me[extensionSchema])
		out["authenticatorAppEnabled"] = attrs["totpEnabled"]
		out["backupCodesEnabled"] = attrs["backupCodeEnabled"]
	}
	if mfa, err := c.SelfService.GetEnabledMFAAuthenticators(ctx); err != nil {
		if err := failed(err, "enabledMfaOptions"); err != nil {
			return nil, err
		}
	} else if mfa != nil {
		out["enabledMfaOptions"] = mfa.EnabledAuthenticators
	}
	if codes, err := c.SelfService.GetRemainingBackupCodes(ctx); err != nil {
		if err := failed(err, "backupCodesRemaining"); err != nil {
			return nil, err
		}
	} else if codes != nil && codes.RemainingBackupCodesCount != nil {
		out["backupCodesRemaining"] = *codes.RemainingBackupCodesCount
	}
	if passkeys, err := c.SelfService.ListPasskeys(ctx); err != nil {
		if err := failed(err, "passkeys"); err != nil {
			return nil, err
		}
	} else {
		// Keep only what identifies the passkey, not its key material.
		summaries := []map[string]interface{}{}
		for _, p := range passkeys {
			summaries = append(summaries, map[string]interface{}{
				"credentialId":     asMap(p["credential"])["credentialId"],
				"displayName":      p["displayName"],
				"registrationTime": p["registrationTime"],
			})
		}
		out["passkeys"] = summaries
	}
	if devices, err := c.SelfService.ListPushDevices(ctx); err != nil {
		if err := failed(err, "pushDevices"); err != nil {
			return nil, err
		}
	} else {
		out["pushDevices"] = devices
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
		case option == "backup-code-authenticator" && out["backupCodesRemaining"] == 0:
		default:
			options = append(options, option)
		}
	}
	return options
}
