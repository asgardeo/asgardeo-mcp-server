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

	"github.com/asgardeo/mcp/internal/selfservice"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const consentMgtV1 = "/api/identity/consent-mgt/v1.0/consents"

func consentTools(c *selfservice.Client) []server.ServerTool {
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
				resp, err := c.Get(ctx, meV1+"/consents/"+url.PathEscape(id), nil)
				if err != nil {
					return nil, err
				}
				return compactConsent(asMap(resp)), nil
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
				resp, err := c.Get(ctx, consentMgtV1+"/receipts/"+url.PathEscape(id), nil)
				if err != nil {
					return nil, err
				}
				return compactReceipt(asMap(resp)), nil
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
				_, err = c.Delete(ctx, consentMgtV1+"/receipts/"+url.PathEscape(id))
				return nil, err
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
				_, err = c.Do(ctx, http.MethodPost, meV1+"/consents/"+url.PathEscape(id)+"/revoke", nil, nil, "")
				return nil, err
			},
			mcp.WithString("consent_id", mcp.Required(), mcp.Description("Policy or communication preference id from list_consents.")),
		),
	}
}

// compactConsent keeps a consent's name, policy link and channels.
func compactConsent(consent map[string]interface{}) map[string]interface{} {
	first := map[string]interface{}{}
	if purposes := asList(consent["purposes"]); len(purposes) > 0 {
		first = asMap(purposes[0])
	}
	out := map[string]interface{}{
		"id": consent["id"], "state": consent["state"], "timestamp": consent["timestamp"],
		"type": first["type"], "name": first["name"],
	}
	if policyURL, ok := asMap(first["properties"])["policyUrl"]; ok {
		out["policyUrl"] = policyURL
	}
	channels := []interface{}{}
	for _, item := range asList(first["elements"]) {
		// "Preference" is a server placeholder, not a channel.
		if name := asMap(item)["displayName"]; name != nil && name != "Preference" {
			channels = append(channels, name)
		}
	}
	if len(channels) > 0 {
		out["channels"] = channels
	}
	return out
}

// compactReceipt keeps what an app consent covers, leaving out the controller and receipt metadata.
func compactReceipt(receipt map[string]interface{}) map[string]interface{} {
	apps := []interface{}{}
	for _, item := range asList(receipt["services"]) {
		service := asMap(item)
		purposes := []interface{}{}
		for _, rawPurpose := range asList(service["purposes"]) {
			purpose := asMap(rawPurpose)
			categories := []interface{}{}
			for _, rawCategory := range asList(purpose["piiCategory"]) {
				category := asMap(rawCategory)
				name := category["piiCategoryDisplayName"]
				if name == nil {
					name = category["piiCategoryName"]
				}
				categories = append(categories, name)
			}
			entry := map[string]interface{}{
				"purpose": purpose["purpose"], "dataCategories": categories,
				"thirdPartyDisclosure": purpose["thirdPartyDisclosure"], "termination": purpose["termination"],
			}
			if name, ok := purpose["thirdPartyName"].(string); ok && name != "" {
				entry["thirdPartyName"] = name
			}
			purposes = append(purposes, entry)
		}
		apps = append(apps, map[string]interface{}{"app": service["serviceDisplayName"], "purposes": purposes})
	}
	out := map[string]interface{}{
		"id": receipt["consentReceiptID"], "state": receipt["state"], "timestamp": receipt["consentTimestamp"], "apps": apps,
	}
	if policyURL, ok := receipt["policyUrl"].(string); ok && policyURL != "" && policyURL != "NONE" {
		out["policyUrl"] = policyURL
	}
	return out
}

// listConsents uses the consent v2 API, or v1 when v2 isn't available.
func listConsents(ctx context.Context, c *selfservice.Client, state string) (interface{}, error) {
	query := url.Values{"attributes": {"purposes"}}
	if state != "" {
		query.Set("state", state)
	}
	resp, err := c.Get(ctx, meV1+"/consents", query)
	if selfservice.IsStatus(err, http.StatusNotFound, http.StatusMethodNotAllowed) {
		return listConsentsV1(ctx, c, state)
	}
	if err != nil {
		return nil, err
	}

	apps, policies, preferences := []interface{}{}, []interface{}{}, []interface{}{}
	for _, item := range asList(resp) {
		consent := asMap(item)
		purposes := asList(consent["purposes"])
		first := map[string]interface{}{}
		if len(purposes) > 0 {
			first = asMap(purposes[0])
		}
		entry := map[string]interface{}{"id": consent["id"], "state": consent["state"], "timestamp": consent["timestamp"]}
		switch first["type"] {
		case "Policy":
			entry["name"], entry["version"] = first["name"], first["version"]
			policies = append(policies, entry)
		case "Preference":
			entry["name"] = first["name"]
			preferences = append(preferences, entry)
		default:
			var names []interface{}
			for _, p := range purposes {
				names = append(names, asMap(p)["name"])
			}
			entry["app"], entry["purposes"] = consent["serviceId"], names
			apps = append(apps, entry)
		}
	}
	return map[string]interface{}{"apps": apps, "policies": policies, "communicationPreferences": preferences}, nil
}

func listConsentsV1(ctx context.Context, c *selfservice.Client, state string) (interface{}, error) {
	me, err := c.Get(ctx, "/scim2/Me", url.Values{"attributes": {"userName"}})
	if err != nil {
		return nil, err
	}
	userName, _ := asMap(me)["userName"].(string)
	if userName == "" {
		return nil, fmt.Errorf("could not read the signed-in user's username")
	}
	query := url.Values{"piiPrincipalId": {userName}}
	if state != "" {
		query.Set("state", state)
	}
	resp, err := c.Get(ctx, consentMgtV1, query)
	if err != nil {
		return nil, err
	}
	apps := []interface{}{}
	for _, item := range asList(resp) {
		receipt := asMap(item)
		apps = append(apps, map[string]interface{}{"id": receipt["consentReceiptID"], "app": receipt["spDisplayName"], "state": receipt["state"]})
	}
	return map[string]interface{}{"apps": apps}, nil
}
