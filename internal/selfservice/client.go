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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/asgardeo/mcp/internal/config"
)

// Client calls the self-service REST APIs as the signed-in user.
type Client struct {
	baseURL  string
	token    string
	clientID string
	http     *http.Client
}

// APIError is a non-2xx response from the identity server.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	hint := ""
	switch e.Status {
	case http.StatusUnauthorized:
		hint = "the access token is missing, expired or revoked; the user needs to sign in again"
	case http.StatusForbidden:
		hint = "this user or application is not permitted to do this"
	case http.StatusNotFound:
		hint = "not found, or not available on this server"
	}
	msg := strings.TrimSpace(strings.Join([]string{hint, e.Message}, ". "))
	return fmt.Sprintf("request failed with status %d: %s", e.Status, strings.Trim(msg, ". "))
}

type authorizationKey struct{}

// WithAuthorization sets the Authorization header sent to the identity server.
func WithAuthorization(ctx context.Context, header string) context.Context {
	return context.WithValue(ctx, authorizationKey{}, header)
}

// NewClient initializes a self-service client from environment variables.
func NewClient() (*Client, error) {
	baseURL := strings.TrimRight(os.Getenv(config.BASE_URL_PARAM), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("missing required environment variable BASE_URL")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if certPath := os.Getenv(config.CERTIFICATE_PATH_PARAM); certPath != "" {
		pem, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("reading CERT_PATH: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("CERT_PATH does not contain a PEM certificate")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool}
	}

	return &Client{
		baseURL:  baseURL,
		token:    os.Getenv(config.ACCESS_TOKEN_PARAM),
		clientID: os.Getenv(config.CLIENT_ID_PARAM),
		http:     &http.Client{Timeout: 15 * time.Second, Transport: transport},
	}, nil
}

// Do sends a request and returns the decoded JSON body.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body interface{}, contentType string) (interface{}, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, err
	}
	authorization, _ := ctx.Value(authorizationKey{}).(string)
	if authorization == "" && c.token != "" {
		authorization = "Bearer " + c.token
	}
	if authorization == "" {
		return nil, &APIError{Status: http.StatusUnauthorized}
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the identity server: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, &APIError{Status: resp.StatusCode, Message: errorMessage(data)}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var out interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unexpected response from the identity server")
	}
	return out, nil
}

func (c *Client) Get(ctx context.Context, path string, query url.Values) (interface{}, error) {
	return c.Do(ctx, http.MethodGet, path, query, nil, "")
}

func (c *Client) Delete(ctx context.Context, path string) (interface{}, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, nil, "")
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

// IsStatus reports whether err is an APIError with one of the statuses.
func IsStatus(err error, statuses ...int) bool {
	apiErr, ok := err.(*APIError)
	if !ok {
		return false
	}
	for _, s := range statuses {
		if apiErr.Status == s {
			return true
		}
	}
	return false
}

// ClientID returns the client ID of the application the token came from, or "" if not set.
func (c *Client) ClientID() string {
	return c.clientID
}
