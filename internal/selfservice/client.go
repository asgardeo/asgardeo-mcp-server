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

package selfservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/asgardeo/go/pkg/common"
	"github.com/asgardeo/go/pkg/config"
	"github.com/asgardeo/go/pkg/sdk"
	internal_config "github.com/asgardeo/mcp/internal/config"
)

// NewClient initializes an SDK client, using the context token when token is empty.
func NewClient(token string) (*sdk.Client, error) {
	baseURL := strings.TrimRight(os.Getenv(internal_config.BASE_URL_PARAM), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("missing required environment variable BASE_URL")
	}
	certPath := os.Getenv(internal_config.CERTIFICATE_PATH_PARAM)

	cfg := config.DefaultClientConfig().
		WithBaseURL(baseURL).
		WithTimeout(15 * time.Second).
		WithCertificatePath(&certPath)
	if token != "" {
		cfg = cfg.WithToken(token)
	} else {
		cfg = cfg.WithContextToken()
	}
	return sdk.New(cfg)
}

// WithAuthorization sets the bearer token of an Authorization header on the context.
func WithAuthorization(ctx context.Context, header string) context.Context {
	if token := BearerToken(header); token != "" {
		return config.ContextWithToken(ctx, token)
	}
	return ctx
}

// BearerToken returns the token of a Bearer Authorization header.
func BearerToken(header string) string {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

// ErrorText describes a failed call with a hint for its status.
func ErrorText(err error) string {
	var apiErr *common.APIError
	if !errors.As(err, &apiErr) {
		return err.Error()
	}
	hint := ""
	switch apiErr.StatusCode {
	case http.StatusUnauthorized:
		hint = "the access token is missing, expired or revoked; the user needs to sign in again"
	case http.StatusForbidden:
		hint = "this user or application is not permitted to do this"
	case http.StatusNotFound:
		hint = "not found, or not available on this server"
	}
	msg := strings.TrimSpace(strings.Join([]string{hint, errorMessage(apiErr.Body)}, ". "))
	return fmt.Sprintf("request failed with status %d: %s", apiErr.StatusCode, strings.Trim(msg, ". "))
}

// errorMessage returns the error text from an error response.
func errorMessage(data []byte) string {
	var body map[string]interface{}
	if json.Unmarshal(data, &body) != nil {
		return ""
	}
	for _, key := range []string{"detail", "description", "message"} {
		if v, ok := body[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// IsStatus reports whether err has one of the statuses.
func IsStatus(err error, statuses ...int) bool {
	var apiErr *common.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, s := range statuses {
		if apiErr.StatusCode == s {
			return true
		}
	}
	return false
}
