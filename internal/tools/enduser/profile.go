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
	"maps"
	"slices"
	"strings"

	"github.com/asgardeo/go/pkg/sdk"
	"github.com/asgardeo/go/pkg/user"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	coreUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	extensionSchema = "urn:scim:wso2:schema"
)

// protectedAttributes are never changed by the agent.
var protectedAttributes = map[string]bool{
	"username": true, "accountlocked": true, "accountdisabled": true,
	"verifiedemailaddresses": true, "verifiedmobilenumbers": true, "verifyemail": true, "verifymobile": true,
}

// contactAttributes are changed with update_email and update_mobile.
var contactAttributes = map[string]bool{
	"emails": true, "phoneNumbers.mobile": true, "emailAddresses": true, "mobileNumbers": true,
}

// profileAttribute is one field of the end-user profile.
type profileAttribute struct {
	Name            string        `json:"name"`
	DisplayName     string        `json:"displayName,omitempty"`
	Type            interface{}   `json:"type,omitempty"`
	MultiValued     bool          `json:"multiValued,omitempty"`
	Editable        bool          `json:"editable"`
	Required        bool          `json:"required,omitempty"`
	CanonicalValues []interface{} `json:"canonicalValues,omitempty"`

	schema, attr, sub string
	// typed is set for a multi-valued entry, such as phoneNumbers.home.
	typed bool
}

func profileTools(c *sdk.Client) []server.ServerTool {
	return []server.ServerTool{
		newTool("get_profile",
			"Gets the signed-in user's profile: name, emails, mobile numbers, addresses and other attributes. "+
				"An email or mobile number not in verifiedEmailAddresses / verifiedMobileNumbers is unverified.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return getProfile(ctx, c, str(args, "attributes"))
			},
			mcp.WithString("attributes", mcp.Description("Comma-separated attributes to return, e.g. name,emails. Returns all when empty.")),
		),
		newTool("list_profile_attributes",
			"Lists the attributes in the user's profile: their names, display names, and whether they are editable or required. "+
				"Emails and mobile numbers are changed with update_email and update_mobile; everything else with update_profile.",
			read,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return loadProfileAttributes(ctx, c)
			},
		),
		newTool("update_profile",
			"Updates attributes in the signed-in user's profile. Use the attribute names exactly as list_profile_attributes returns them, "+
				"with the new value: text, true/false, or a list for a multiValued attribute. An empty value clears the attribute. "+
				`Example: {"name.givenName":"Anne","country":"Sri Lanka"}. Use update_email and update_mobile for emails and mobile numbers.`,
			write,
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				values := asMap(args["attributes"])
				if len(values) == 0 {
					return nil, fmt.Errorf("attributes is required")
				}
				return updateProfile(ctx, c, values)
			},
			mcp.WithObject("attributes",
				mcp.Required(),
				mcp.Description("Attribute names from list_profile_attributes mapped to their new values."),
			),
		),
		contactTool(c, emailContact, "update_email", "email address"),
		contactTool(c, mobileContact, "update_mobile", "mobile number"),
	}
}

func contactTool(c *sdk.Client, ct contactType, name, label string) server.ServerTool {
	return newTool(name,
		fmt.Sprintf("Adds a %[1]s to the signed-in user's account, removes one, or makes one the primary %[1]s. "+
			"The primary %[1]s can't be removed; make another one primary first.", label),
		destructive,
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			action, err := requiredStr(args, "action")
			if err != nil {
				return nil, err
			}
			value, err := requiredStr(args, ct.kind)
			if err != nil {
				return nil, err
			}
			return updateContact(ctx, c, ct, action, value)
		},
		mcp.WithString("action", mcp.Required(), mcp.Enum("add", "remove", "make_primary")),
		mcp.WithString(ct.kind, mcp.Required(), mcp.Description(fmt.Sprintf("The %s.", label))),
	)
}

// loadProfileAttributes reads the end-user profile from the server's SCIM schemas.
func loadProfileAttributes(ctx context.Context, c *sdk.Client) ([]profileAttribute, error) {
	schemas, err := c.User.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	var attrs []profileAttribute
	for _, rawSchema := range schemaList(schemas) {
		schema := asMap(rawSchema)
		schemaID, _ := schema["id"].(string)

		for _, rawAttr := range asList(schema["attributes"]) {
			attr := asMap(rawAttr)
			name, _ := attr["name"].(string)
			multiValued, _ := attr["multiValued"].(bool)
			subAttrs := asList(attr["subAttributes"])

			// Skip complex attributes without sub-attributes, except emails.
			complex := strings.EqualFold(fmt.Sprint(attr["type"]), "complex")
			if len(subAttrs) == 0 && inProfile(attr) && (!complex || name == "emails") {
				attrs = append(attrs, newProfileAttribute(schemaID, name, "", attr, false))
			}
			for _, rawSub := range subAttrs {
				subAttr := asMap(rawSub)
				subName, _ := subAttr["name"].(string)
				if inProfile(subAttr) {
					attrs = append(attrs, newProfileAttribute(schemaID, name, subName, subAttr, multiValued))
				}
			}
		}
	}

	// Names in more than one schema are prefixed with the schema.
	nameCount := map[string]int{}
	for _, a := range attrs {
		nameCount[a.Name]++
	}
	kept := attrs[:0]
	for _, a := range attrs {
		if isProtected(a.attr) || isProtected(a.sub) {
			continue
		}
		if nameCount[a.Name] > 1 {
			a.Name = a.schema + ":" + a.Name
		}
		kept = append(kept, a)
	}
	return kept, nil
}

func newProfileAttribute(schema, attr, sub string, spec map[string]interface{}, typed bool) profileAttribute {
	name := attr
	if sub != "" {
		name += "." + sub
	}
	a := profileAttribute{
		Name:     name,
		Type:     spec["type"],
		Editable: isEditable(mutability(spec)),
		schema:   schema,
		attr:     attr,
		sub:      sub,
		typed:    typed,
	}
	a.DisplayName, _ = spec["displayName"].(string)
	a.MultiValued, _ = spec["multiValued"].(bool)
	a.CanonicalValues = asList(spec["canonicalValues"])
	if required, ok := asMap(asMap(spec["profiles"])["endUser"])["required"].(bool); ok {
		a.Required = required
	} else {
		a.Required, _ = spec["required"].(bool)
	}
	if len(a.CanonicalValues) == 0 {
		a.CanonicalValues = nil
	}
	return a
}

// inProfile reports whether an attribute is in the end-user profile.
func inProfile(attr map[string]interface{}) bool {
	supported, ok := asMap(asMap(attr["profiles"])["endUser"])["supportedByDefault"]
	if !ok {
		supported = attr["supportedByDefault"]
	}
	switch val := supported.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true")
	}
	return false
}

// mutability returns the end-user profile's mutability, or the global one.
func mutability(attr map[string]interface{}) string {
	if m, ok := asMap(asMap(attr["profiles"])["endUser"])["mutability"].(string); ok && m != "" {
		return m
	}
	m, _ := attr["mutability"].(string)
	return m
}

// isEditable accepts both READ_WRITE and readWrite.
func isEditable(m string) bool {
	switch strings.ToLower(strings.ReplaceAll(m, "_", "")) {
	case "readwrite", "writeonly":
		return true
	}
	return false
}

func schemaList(resp interface{}) []interface{} {
	if m := asMap(resp); m != nil {
		return asList(m["Resources"])
	}
	return asList(resp)
}

// isProtected reports whether the agent must not see or change an attribute.
func isProtected(name string) bool {
	lower := strings.ToLower(name)
	return protectedAttributes[lower] || strings.Contains(lower, "password") || strings.Contains(lower, "failed")
}

// updateProfile validates the attributes and patches them in one request.
func updateProfile(ctx context.Context, c *sdk.Client, values map[string]interface{}) (interface{}, error) {
	attrs, err := loadProfileAttributes(ctx, c)
	if err != nil {
		return nil, err
	}
	attrsByName := map[string]profileAttribute{}
	for _, a := range attrs {
		attrsByName[a.Name] = a
	}

	var ops []user.PatchOperationModel
	for _, name := range slices.Sorted(maps.Keys(values)) {
		a, ok := attrsByName[name]
		switch {
		case contactAttributes[name]:
			return nil, fmt.Errorf("change %s with update_email or update_mobile", name)
		case !ok:
			return nil, fmt.Errorf("%s isn't an attribute in the user's profile; use a name from list_profile_attributes", name)
		case !a.Editable:
			return nil, fmt.Errorf("%s is read-only in the user's profile", name)
		case isEmpty(values[name]) && a.Required:
			return nil, fmt.Errorf("%s is required and can't be cleared", name)
		}
		ops = append(ops, patchOperation(a, values[name]))
	}
	return patchMe(ctx, c, ops)
}

// patchOperation builds the SCIM operation for one attribute.
func patchOperation(a profileAttribute, value interface{}) user.PatchOperationModel {
	if value == nil {
		value = ""
	}
	var patchValue map[string]interface{}

	switch {
	case a.typed && isEmpty(value):
		path := fmt.Sprintf(`%s[type eq "%s"]`, a.attr, a.sub)
		if a.schema != coreUserSchema {
			path = a.schema + ":" + path
		}
		return patchOp(user.PatchOpRemove, path, nil)
	case a.typed:
		field := "value"
		if a.attr == "addresses" {
			field = "formatted"
		}
		patchValue = map[string]interface{}{a.attr: []interface{}{map[string]interface{}{"type": a.sub, field: value}}}
		return patchOp(user.PatchOpAdd, "", inSchema(a.schema, patchValue))
	case a.sub != "":
		patchValue = map[string]interface{}{a.attr: map[string]interface{}{a.sub: value}}
	default:
		patchValue = map[string]interface{}{a.attr: value}
	}
	return patchOp(user.PatchOpReplace, "", inSchema(a.schema, patchValue))
}

func patchOp(op, path string, value interface{}) user.PatchOperationModel {
	return user.PatchOperationModel{Op: op, Path: path, Value: value}
}

// inSchema nests an extension attribute under its schema.
func inSchema(schema string, attrs map[string]interface{}) map[string]interface{} {
	if schema == coreUserSchema {
		return attrs
	}
	return map[string]interface{}{schema: attrs}
}

// getProfile returns the user's profile as a plain map.
func getProfile(ctx context.Context, c *sdk.Client, attributes string) (map[string]interface{}, error) {
	profile, err := c.User.GetMe(ctx, attributes)
	return profile, err
}

// patchMe updates the profile and returns it as a plain map.
func patchMe(ctx context.Context, c *sdk.Client, ops []user.PatchOperationModel) (map[string]interface{}, error) {
	profile, err := c.User.PatchMe(ctx, ops)
	return profile, err
}

func isEmpty(v interface{}) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}:
		return len(val) == 0
	}
	return false
}

// contactType describes how the user's emails or mobile numbers are stored.
type contactType struct {
	kind         string
	listAttr     string
	verifiedAttr string
	primary      func(me map[string]interface{}) string
	pending      func(extensionAttrs map[string]interface{}) string
	clearPending map[string]interface{}
	setPrimary   func(value string) user.PatchOperationModel
}

var emailContact = contactType{
	kind:         "email",
	listAttr:     "emailAddresses",
	verifiedAttr: "verifiedEmailAddresses",
	primary: func(me map[string]interface{}) string {
		for _, item := range asList(me["emails"]) {
			if address, ok := item.(string); ok {
				return address
			}
			if email := asMap(item); email["primary"] == true {
				address, _ := email["value"].(string)
				return address
			}
		}
		return ""
	},
	pending: func(extensionAttrs map[string]interface{}) string {
		for _, item := range asList(extensionAttrs["pendingEmails"]) {
			address, _ := asMap(item)["value"].(string)
			return address
		}
		return ""
	},
	clearPending: map[string]interface{}{"pendingEmails": []interface{}{map[string]interface{}{"type": "value", "value": ""}}},
	setPrimary: func(value string) user.PatchOperationModel {
		return patchOp(user.PatchOpAdd, "", map[string]interface{}{
			"emails": []interface{}{map[string]interface{}{"value": value, "primary": true}},
		})
	},
}

var mobileContact = contactType{
	kind:         "mobile",
	listAttr:     "mobileNumbers",
	verifiedAttr: "verifiedMobileNumbers",
	primary: func(me map[string]interface{}) string {
		for _, item := range asList(me["phoneNumbers"]) {
			if phone := asMap(item); phone["type"] == "mobile" {
				number, _ := phone["value"].(string)
				return number
			}
		}
		return ""
	},
	pending: func(extensionAttrs map[string]interface{}) string {
		number, _ := extensionAttrs["pendingMobileNumber"].(string)
		return number
	},
	clearPending: map[string]interface{}{"pendingMobileNumber": ""},
	setPrimary: func(value string) user.PatchOperationModel {
		return patchOp(user.PatchOpAdd, "", map[string]interface{}{
			"phoneNumbers": []interface{}{map[string]interface{}{"type": "mobile", "value": value}},
		})
	},
}

// updateContact adds, removes or makes primary an email address or mobile number.
func updateContact(ctx context.Context, c *sdk.Client, ct contactType, action, value string) (interface{}, error) {
	me, err := getProfile(ctx, c, "")
	if err != nil {
		return nil, err
	}
	attrs, err := loadProfileAttributes(ctx, c)
	if err != nil {
		return nil, err
	}
	// The list exists when multiple values are allowed.
	hasList := slices.ContainsFunc(attrs, func(a profileAttribute) bool { return a.Name == ct.listAttr && a.Editable })
	extensionAttrs := asMap(me[extensionSchema])
	current := stringList(extensionAttrs[ct.listAttr])
	primary := ct.primary(me)

	var ops []user.PatchOperationModel
	switch action {
	case "add":
		if !hasList {
			return nil, fmt.Errorf("this account has one %[1]s; use make_primary to change it", ct.kind)
		}
		if slices.Contains(current, value) {
			return nil, fmt.Errorf("%s is already added", value)
		}
		ops = append(ops, replaceInExtension(map[string]interface{}{ct.listAttr: append(current, value)}))
		if len(current) == 0 {
			ops = append(ops, ct.setPrimary(value))
		}
	case "make_primary":
		if value == primary {
			return nil, fmt.Errorf("%s is already the primary %s", value, ct.kind)
		}
		if hasList && !slices.Contains(current, value) {
			return nil, fmt.Errorf("add %s before making it primary", value)
		}
		ops = append(ops, ct.setPrimary(value))
	case "remove":
		if value == primary {
			return nil, fmt.Errorf("%s is the primary %s; make another one primary first", value, ct.kind)
		}
		// Removing a pending value cancels its verification.
		isPending := ct.pending(extensionAttrs) == value
		if !slices.Contains(current, value) && !isPending {
			return nil, fmt.Errorf("%s isn't one of the user's %ss", value, ct.kind)
		}
		changes := map[string]interface{}{}
		if slices.Contains(current, value) {
			changes[ct.listAttr] = without(current, value)
		}
		if verified := stringList(extensionAttrs[ct.verifiedAttr]); slices.Contains(verified, value) {
			changes[ct.verifiedAttr] = without(verified, value)
		}
		if isPending {
			maps.Copy(changes, ct.clearPending)
		}
		ops = append(ops, replaceInExtension(changes))
	default:
		return nil, fmt.Errorf("action must be add, remove or make_primary")
	}

	profile, err := patchMe(ctx, c, ops)
	if err != nil {
		return nil, err
	}
	if action != "remove" && ct.pending(asMap(profile[extensionSchema])) == value {
		return map[string]interface{}{
			"profile":             profile,
			"pendingVerification": value,
			"note":                "This takes effect after the user verifies it in the portal.",
		}, nil
	}
	return profile, nil
}

// replaceInExtension replaces attributes in the extension schema.
func replaceInExtension(attrs map[string]interface{}) user.PatchOperationModel {
	return patchOp(user.PatchOpReplace, "", inSchema(extensionSchema, attrs))
}

func stringList(v interface{}) []string {
	out := []string{}
	for _, item := range asList(v) {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func without(list []string, value string) []string {
	return slices.DeleteFunc(slices.Clone(list), func(s string) bool { return s == value })
}
