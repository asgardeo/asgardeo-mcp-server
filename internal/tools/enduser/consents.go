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

	"github.com/asgardeo/go/pkg/consent"
	"github.com/asgardeo/go/pkg/sdk"
	"github.com/asgardeo/go/pkg/self_service"
	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func consentTools(c *sdk.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("list_consents",
			"Lists the signed-in user's consents grouped as apps (data shared with an app), policies (accepted terms and privacy policies) "+
				"and communicationPreferences (subscribed purposes). Timestamps are epoch milliseconds.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				state := str(args, "state")
				if state == "" {
					state = "ACTIVE"
				}
				if state == "ALL" {
					state = ""
				}
				return listConsents(ctx, c, state)
			},
			mcp.WithString("state", mcp.Description("Which consents to list. Defaults to ACTIVE."), mcp.Enum("ACTIVE", "REVOKED", "ALL")),
		),
		newTool("get_consent",
			"Gets one policy or communication preference consent: its name, policy link and subscribed channels.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "consent_id")
				if err != nil {
					return nil, err
				}
				consent, err := c.SelfService.GetConsent(ctx, id)
				if err != nil || consent == nil {
					return nil, err
				}
				return compactConsent(*consent), nil
			},
			mcp.WithString("consent_id", mcp.Required(), mcp.Description("Consent id from list_consents.")),
		),
		newTool("get_app_consent",
			"Gets what an app consent covers: the app, purposes, data categories, third-party sharing, policy URL and when consent was given.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "consent_id")
				if err != nil {
					return nil, err
				}
				receipt, err := c.Consent.GetReceipt(ctx, id)
				if err != nil || receipt == nil {
					return nil, err
				}
				return compactReceipt(*receipt), nil
			},
			mcp.WithString("consent_id", mcp.Required(), mcp.Description("App consent id from list_consents.")),
		),
		newTool("withdraw_app_consent",
			"Withdraws the user's consent to share data with an app. The app may ask again or stop working for the user.",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "consent_id")
				if err != nil {
					return nil, err
				}
				return nil, c.Consent.RevokeReceipt(ctx, id)
			},
			mcp.WithString("consent_id", mcp.Required(), mcp.Description("App consent id from list_consents.")),
		),
		newTool("revoke_consent",
			"Withdraws acceptance of a policy (the user may be asked to accept it again at next sign-in), "+
				"or unsubscribes from a communication preference (stops every channel of that purpose; name the channels when confirming).",
			destructive,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				id, err := requiredStr(args, "consent_id")
				if err != nil {
					return nil, err
				}
				return nil, c.SelfService.RevokeConsent(ctx, id)
			},
			mcp.WithString("consent_id", mcp.Required(), mcp.Description("Policy or communication preference id from list_consents.")),
		),
	}
}

// compactConsent keeps a consent's name, policy link and channels.
func compactConsent(consent self_service.ConsentModel) map[string]interface{} {
	first := self_service.ConsentPurposeModel{}
	if purposes := deref(consent.Purposes); len(purposes) > 0 {
		first = purposes[0]
	}
	out := map[string]interface{}{
		"id": consent.Id, "state": consent.State, "timestamp": consent.Timestamp,
		"type": first.Type, "name": first.Name,
	}
	if policyURL, ok := deref(first.Properties)["policyUrl"]; ok {
		out["policyUrl"] = policyURL
	}
	channels := []string{}
	for _, element := range deref(first.Elements) {
		// "Preference" is a server placeholder, not a channel.
		if name := deref(element.DisplayName); name != "" && name != "Preference" {
			channels = append(channels, name)
		}
	}
	if len(channels) > 0 {
		out["channels"] = channels
	}
	return out
}

// compactReceipt keeps what an app consent covers.
func compactReceipt(receipt consent.ConsentReceiptModel) map[string]interface{} {
	apps := []interface{}{}
	for _, service := range receipt.Services {
		purposes := []interface{}{}
		for _, purpose := range deref(service.Purposes) {
			categories := []string{}
			for _, category := range purpose.PiiCategory {
				name := deref(category.PiiCategoryDisplayName)
				if name == "" {
					name = deref(category.PiiCategoryName)
				}
				categories = append(categories, name)
			}
			entry := map[string]interface{}{
				"purpose": purpose.Purpose, "dataCategories": categories,
				"thirdPartyDisclosure": purpose.ThirdPartyDisclosure, "termination": purpose.Termination,
			}
			if purpose.ThirdPartyName != "" {
				entry["thirdPartyName"] = purpose.ThirdPartyName
			}
			purposes = append(purposes, entry)
		}
		apps = append(apps, map[string]interface{}{"app": service.ServiceDisplayName, "purposes": purposes})
	}
	out := map[string]interface{}{
		"id": receipt.ConsentReceiptID, "state": receipt.State, "timestamp": receipt.ConsentTimestamp, "apps": apps,
	}
	if receipt.PolicyUrl != "" && receipt.PolicyUrl != "NONE" {
		out["policyUrl"] = receipt.PolicyUrl
	}
	return out
}

// listConsents uses the user consents API, or the consent management API when it isn't available.
func listConsents(ctx context.Context, c *sdk.Client, state string) (interface{}, error) {
	attributes := "purposes"
	params := self_service.ConsentListParamsModel{Attributes: &attributes}
	if state != "" {
		consentState := self_service.ConsentStateModel(state)
		params.State = &consentState
	}
	consents, err := c.SelfService.ListConsents(ctx, &params)
	if selfservice.IsStatus(err, http.StatusNotFound, http.StatusMethodNotAllowed) {
		return listConsentsV1(ctx, c, state)
	}
	if err != nil {
		return nil, err
	}

	apps, policies, preferences := []interface{}{}, []interface{}{}, []interface{}{}
	for _, consent := range consents {
		purposes := deref(consent.Purposes)
		first := self_service.ConsentPurposeSummaryModel{}
		if len(purposes) > 0 {
			first = purposes[0]
		}
		entry := map[string]interface{}{"id": consent.Id, "state": consent.State, "timestamp": consent.Timestamp}
		switch deref(first.Type) {
		case "Policy":
			entry["name"], entry["version"] = first.Name, first.Version
			policies = append(policies, entry)
		case "Preference":
			entry["name"] = first.Name
			preferences = append(preferences, entry)
		default:
			var names []interface{}
			for _, p := range purposes {
				names = append(names, p.Name)
			}
			entry["app"], entry["purposes"] = consent.ServiceId, names
			apps = append(apps, entry)
		}
	}
	return map[string]interface{}{"apps": apps, "policies": policies, "communicationPreferences": preferences}, nil
}

func listConsentsV1(ctx context.Context, c *sdk.Client, state string) (interface{}, error) {
	me, err := c.User.GetMe(ctx, "userName")
	if err != nil {
		return nil, err
	}
	userName, _ := me["userName"].(string)
	if userName == "" {
		return nil, fmt.Errorf("could not read the signed-in user's username")
	}
	params := consent.ConsentListParamsModel{PiiPrincipalId: &userName}
	if state != "" {
		params.State = &state
	}
	receipts, err := c.Consent.List(ctx, &params)
	if err != nil {
		return nil, err
	}
	apps := []interface{}{}
	for _, receipt := range receipts {
		apps = append(apps, map[string]interface{}{"id": receipt.ConsentReceiptID, "app": receipt.SpDisplayName, "state": receipt.State})
	}
	return map[string]interface{}{"apps": apps}, nil
}
