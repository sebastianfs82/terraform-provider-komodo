// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

func (c *Client) AddUserToUserGroup(ctx context.Context, req AddUserToUserGroupRequest) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "AddUserToUserGroup",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

func (c *Client) RemoveUserFromUserGroup(ctx context.Context, req RemoveUserFromUserGroupRequest) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "RemoveUserFromUserGroup",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

func (c *Client) SetEveryoneUserGroup(ctx context.Context, req SetEveryoneUserGroupRequest) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "SetEveryoneUserGroup",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

// UserGroup CRUD.
func (c *Client) CreateUserGroup(ctx context.Context, req CreateUserGroupRequest) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "CreateUserGroup",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

func (c *Client) ListUserGroups(ctx context.Context) ([]UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "ListUserGroups",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var groups []UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return groups, nil
}

func (c *Client) GetUserGroup(ctx context.Context, name string) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "GetUserGroup",
		"params": map[string]interface{}{"user_group": name},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

func (c *Client) RenameUserGroup(ctx context.Context, oldName, newName string) (*UserGroup, error) {
	payload := map[string]interface{}{
		"type":   "RenameUserGroup",
		"params": RenameUserGroupRequest{ID: oldName, Name: newName},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var group UserGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &group, nil
}

func (c *Client) DeleteUserGroup(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteUserGroup",
		"params": DeleteUserGroupRequest{ID: id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(strings.ToLower(string(body)), "no usergroup found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type Client struct {
	endpoint   string
	username   string
	password   string
	apiKey     string
	apiSecret  string
	httpClient *http.Client

	// JWT token management
	mu    sync.RWMutex
	token string
}

// Endpoint returns the base URL of the Komodo API endpoint.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// NewClient creates a new Komodo API client using username/password authentication.
func NewClient(endpoint, username, password string) *Client {
	return &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		username:   username,
		password:   password,
		httpClient: http.DefaultClient,
	}
}

// NewClientWithApiKey creates a new Komodo API client using API key/secret authentication.
// The key and secret are sent directly as X-API-KEY and X-API-SECRET headers on every request.
func NewClientWithApiKey(endpoint, apiKey, apiSecret string) *Client {
	return &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: http.DefaultClient,
	}
}

// LoginResponse represents the JWT data inside a successful login response.
type LoginResponse struct {
	JWT string `json:"jwt"`
}

// loginLocalUserResponse is the tagged-union response from the Komodo 2.x
// /auth/login/LoginLocalUser endpoint: {"type":"Jwt","data":{"jwt":"..."}}.
type loginLocalUserResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ApiKey represents a Komodo API key.
type ApiKey struct {
	Key       string `json:"key"`
	Secret    string `json:"secret"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Expires   int64  `json:"expires"`
}

// CreateApiKeyRequest represents the request to create an API key.
type CreateApiKeyRequest struct {
	Name    string `json:"name"`
	Expires int64  `json:"expires"`
}

// DeleteApiKeyRequest represents the request to delete an API key.
type DeleteApiKeyRequest struct {
	Key string `json:"key"`
}

// CreateApiKeyForServiceUserRequest represents the request to create an API key for a service user.
type CreateApiKeyForServiceUserRequest struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Expires int64  `json:"expires"`
}

// DeleteApiKeyForServiceUserRequest represents the request to delete an API key for a service user.
type DeleteApiKeyForServiceUserRequest struct {
	Key string `json:"key"`
}

// ListApiKeysRequest represents the request to list API keys.
type ListApiKeysRequest struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// login authenticates with the Komodo server and obtains a JWT token.
func (c *Client) login(ctx context.Context) error {
	loginReq := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/auth/login/LoginLocalUser", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var outer loginLocalUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&outer); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if outer.Type != "Jwt" {
		return fmt.Errorf("login requires additional authentication (%s) which is not supported by the Terraform provider; use an API key instead", outer.Type)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(outer.Data, &loginResp); err != nil {
		return fmt.Errorf("failed to decode JWT from login response: %w", err)
	}

	c.mu.Lock()
	c.token = loginResp.JWT
	c.mu.Unlock()

	return nil
}

// getToken returns the current JWT token, logging in if necessary.
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	if token == "" {
		if err := c.login(ctx); err != nil {
			return "", err
		}
		c.mu.RLock()
		token = c.token
		c.mu.RUnlock()
	}

	return token, nil
}

// doRequest makes an authenticated HTTP request to the Komodo API.
func (c *Client) doRequest(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		// API key authentication: pass key and secret directly as headers
		req.Header.Set("X-API-KEY", c.apiKey)
		req.Header.Set("X-API-SECRET", c.apiSecret)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		return resp, nil
	}

	// JWT authentication
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// If we get a 401, try to re-login and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()

		token, err = c.getToken(ctx)
		if err != nil {
			return nil, err
		}

		// Recreate request body if needed
		if body != nil {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
	}

	return resp, nil
}

// ListApiKeys lists all API keys for the authenticated user.
func (c *Client) ListApiKeys(ctx context.Context) ([]ApiKey, error) {
	listReq := ListApiKeysRequest{
		Type:   "ListApiKeys",
		Params: map[string]interface{}{},
	}

	resp, err := c.doRequest(ctx, "/read", listReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var keys []ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return keys, nil
}

// GetApiKey gets information about a specific API key by its key ID.
func (c *Client) GetApiKey(ctx context.Context, keyID string) (*ApiKey, error) {
	keys, err := c.ListApiKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Key == keyID {
			return &key, nil
		}
	}

	return nil, nil
}

// CreateApiKey creates a new API key.
func (c *Client) CreateApiKey(ctx context.Context, req CreateApiKeyRequest) (*ApiKey, error) {
	resp, err := c.doRequest(ctx, "/auth/manage/CreateApiKey", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var key ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// The API only returns key and secret, so populate the request values
	key.Name = req.Name
	key.Expires = req.Expires

	// Fetch the full key details to get user_id and created_at
	fullKey, err := c.GetApiKey(ctx, key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full key details: %w", err)
	}

	// If we found the full details, merge them with our create response
	if fullKey != nil {
		key.UserID = fullKey.UserID
		key.CreatedAt = fullKey.CreatedAt
		// Use the name and expires from the list if they're populated, otherwise keep request values
		if fullKey.Name != "" {
			key.Name = fullKey.Name
		}
		if fullKey.Expires != 0 {
			key.Expires = fullKey.Expires
		}
	}

	return &key, nil
}

// DeleteApiKey deletes an API key.
func (c *Client) DeleteApiKey(ctx context.Context, req DeleteApiKeyRequest) error {
	resp, err := c.doRequest(ctx, "/auth/manage/DeleteApiKey", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(strings.ToLower(string(body)), "no api key") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Variable CRUD

// CreateVariable creates a new variable using the Komodo API.
func (c *Client) CreateVariable(ctx context.Context, req CreateVariableRequest) (*Variable, error) {
	// The Komodo API expects a write request with type CreateVariable
	payload := map[string]interface{}{
		"type":   "CreateVariable",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var variable Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &variable, nil
}

// GetVariable gets information about a specific variable by its ID using the Komodo API.
func (c *Client) GetVariable(ctx context.Context, id string) (*Variable, error) {
	// Look up variable by name only (case-insensitive)
	variables, err := c.ListVariables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list variables: %w", err)
	}
	var found *Variable
	for _, v := range variables {
		if strings.EqualFold(v.Name, id) {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("variable with name %s not found", id)
	}
	return found, nil
}

// UpdateVariable updates an existing variable using the Komodo API.
func (c *Client) UpdateVariable(ctx context.Context, id string, req CreateVariableRequest) (*Variable, error) {
	// Look up variable by name only (case-insensitive)
	variables, err := c.ListVariables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list variables: %w", err)
	}
	var found *Variable
	for _, v := range variables {
		if strings.EqualFold(v.Name, id) {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("variable with name %s not found", id)
	}

	// Update value
	valuePayload := map[string]interface{}{
		"type":   "UpdateVariableValue",
		"params": map[string]interface{}{"name": found.Name, "value": req.Value},
	}
	resp, err := c.doRequest(ctx, "/write", valuePayload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Update description
	descPayload := map[string]interface{}{
		"type":   "UpdateVariableDescription",
		"params": map[string]interface{}{"name": found.Name, "description": req.Description},
	}
	resp2, err := c.doRequest(ctx, "/write", descPayload)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp2.StatusCode, string(body))
	}

	// Update is_secret
	secretPayload := map[string]interface{}{
		"type":   "UpdateVariableIsSecret",
		"params": map[string]interface{}{"name": found.Name, "is_secret": req.IsSecret},
	}
	resp3, err := c.doRequest(ctx, "/write", secretPayload)
	if err != nil {
		return nil, err
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp3.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp3.StatusCode, string(body))
	}

	// Return the updated variable
	return c.GetVariable(ctx, id)
}

// DeleteVariable deletes a variable using the Komodo API.
func (c *Client) DeleteVariable(ctx context.Context, req DeleteVariableRequest) error {
	// Look up variable by name only (case-insensitive)
	variables, err := c.ListVariables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list variables: %w", err)
	}
	var found *Variable
	for _, v := range variables {
		if strings.EqualFold(v.Name, req.ID) {
			found = &v
			break
		}
	}
	if found == nil {
		return nil // already deleted
	}
	payload := map[string]interface{}{
		"type":   "DeleteVariable",
		"params": map[string]interface{}{"name": found.Name},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListVariables lists all variables for the authenticated user.
func (c *Client) ListVariables(ctx context.Context) ([]Variable, error) {
	listReq := map[string]interface{}{
		"type":   "ListVariables",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", listReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var variables []Variable
	if err := json.NewDecoder(resp.Body).Decode(&variables); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return variables, nil
}

// UpdateResourceMetaRequest is the payload for the UpdateResourceMeta write API.
type UpdateResourceMetaRequest struct {
	Target ResourceTarget `json:"target"`
	Tags   *[]string      `json:"tags,omitempty"`
}

// UpdateResourceMeta updates the meta fields (tags, description, template) of a resource.
func (c *Client) UpdateResourceMeta(ctx context.Context, req UpdateResourceMetaRequest) error {
	payload := map[string]interface{}{
		"type":   "UpdateResourceMeta",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Tag represents a Komodo tag.
type Tag struct {
	ID    OID    `json:"_id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Color string `json:"color"`
}

// Tag CRUD

func (c *Client) CreateTag(ctx context.Context, req CreateTagRequest) (*Tag, error) {
	payload := map[string]interface{}{
		"type":   "CreateTag",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &tag, nil
}

func (c *Client) GetTag(ctx context.Context, name string) (*Tag, error) {
	// List all tags and find by name (case-insensitive)
	tags, err := c.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	var matches []*Tag
	for i, tag := range tags {
		if strings.EqualFold(tag.Name, name) {
			matches = append(matches, &tags[i])
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no tag found matching %q", name)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple tags found matching %q (case-insensitive); tag names must be unique (case-insensitive) for Terraform management", name)
	}
	return matches[0], nil
}

// ListTags lists all tags for the authenticated user.
func (c *Client) ListTags(ctx context.Context) ([]Tag, error) {
	listReq := map[string]interface{}{
		"type":   "ListTags",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", listReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return tags, nil
}

func (c *Client) UpdateTag(ctx context.Context, name string, req CreateTagRequest) (*Tag, error) {
	// Rename if needed
	if !strings.EqualFold(name, req.Name) {
		// Look up tag ID by the old name so we can pass it to the RenameTag API.
		existing, lookupErr := c.GetTag(ctx, name)
		if lookupErr != nil {
			return nil, fmt.Errorf("failed to look up tag %q before rename: %w", name, lookupErr)
		}
		renamePayload := map[string]interface{}{
			"type": "RenameTag",
			"params": RenameTagRequest{
				ID:      existing.ID.OID,
				OldName: name,
				Name:    req.Name,
			},
		}
		resp, err := c.doRequest(ctx, "/write", renamePayload)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		name = req.Name
	}
	// Update color only when one is provided.
	if req.Color != "" {
		colorPayload := map[string]interface{}{
			"type":   "UpdateTagColor",
			"params": UpdateTagColorRequest{Tag: name, Color: req.Color},
		}
		resp2, err := c.doRequest(ctx, "/write", colorPayload)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp2.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp2.StatusCode, string(body))
		}
	}
	return c.GetTag(ctx, name)
}

func (c *Client) DeleteTag(ctx context.Context, name string) error {
	// Accept either ObjectId or name, try both
	var tag *Tag
	if len(name) == 24 && isHex(name) {
		tags, listErr := c.ListTags(ctx)
		if listErr != nil {
			return fmt.Errorf("failed to list tags: %w", listErr)
		}
		for _, t := range tags {
			if t.ID.OID == name {
				tag = &t
				break
			}
		}
	} else {
		// Always search by name (case-insensitive)
		tags, listErr := c.ListTags(ctx)
		if listErr != nil {
			return fmt.Errorf("failed to list tags: %w", listErr)
		}
		for _, t := range tags {
			if strings.EqualFold(t.Name, name) {
				tag = &t
				break
			}
		}
	}
	if tag == nil || tag.ID.OID == "" {
		return nil // already deleted
	}
	payload := map[string]interface{}{
		"type":   "DeleteTag",
		"params": DeleteTagRequest{ID: tag.ID.OID},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil

}

// User management

func (c *Client) CreateLocalUser(ctx context.Context, req CreateLocalUserRequest) (*User, error) {
	payload := map[string]interface{}{
		"type":   "CreateLocalUser",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &user, nil
}

func (c *Client) FindUser(ctx context.Context, user string) (*User, error) {
	payload := map[string]interface{}{
		"type":   "FindUser",
		"params": FindUserRequest{User: user},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(bodyStr, "no user found") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &u, nil
}

func (c *Client) UpdateUserBasePermissions(ctx context.Context, req UpdateUserBasePermissionsRequest) error {
	payload := map[string]interface{}{
		"type":   "UpdateUserBasePermissions",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) UpdateUserAdmin(ctx context.Context, req UpdateUserAdminRequest) error {
	payload := map[string]interface{}{
		"type":   "UpdateUserAdmin",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteUser(ctx context.Context, user string) error {
	payload := map[string]interface{}{
		"type":   "DeleteUser",
		"params": DeleteUserRequest{User: user},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(strings.ToLower(string(body)), "no user found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	payload := map[string]interface{}{
		"type":   "ListUsers",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return users, nil
}

func (c *Client) ListServiceUsers(ctx context.Context) ([]User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	var serviceUsers []User
	for _, u := range users {
		if u.Config.Type == "Service" {
			serviceUsers = append(serviceUsers, u)
		}
	}
	if serviceUsers == nil {
		serviceUsers = []User{}
	}
	return serviceUsers, nil
}

func (c *Client) CreateServiceUser(ctx context.Context, req CreateServiceUserRequest) (*User, error) {
	payload := map[string]interface{}{
		"type":   "CreateServiceUser",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &user, nil
}

func (c *Client) UpdateServiceUserDescription(ctx context.Context, req UpdateServiceUserDescriptionRequest) error {
	payload := map[string]interface{}{
		"type":   "UpdateServiceUserDescription",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListApiKeysForServiceUser lists all API keys belonging to a service user.
func (c *Client) ListApiKeysForServiceUser(ctx context.Context, userID string) ([]ApiKey, error) {
	payload := map[string]interface{}{
		"type":   "ListApiKeysForServiceUser",
		"params": map[string]string{"user": userID},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var keys []ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return keys, nil
}

// GetApiKeyForServiceUser retrieves a specific API key for a service user by key ID.
func (c *Client) GetApiKeyForServiceUser(ctx context.Context, userID, keyID string) (*ApiKey, error) {
	keys, err := c.ListApiKeysForServiceUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if key.Key == keyID {
			return &key, nil
		}
	}
	return nil, nil
}

// CreateApiKeyForServiceUser creates an API key for a service user.
func (c *Client) CreateApiKeyForServiceUser(ctx context.Context, req CreateApiKeyForServiceUserRequest) (*ApiKey, error) {
	payload := map[string]interface{}{
		"type":   "CreateApiKeyForServiceUser",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var key ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	key.Name = req.Name
	key.Expires = req.Expires
	key.UserID = req.UserID

	fullKey, err := c.GetApiKeyForServiceUser(ctx, req.UserID, key.Key)
	if err == nil && fullKey != nil {
		key.CreatedAt = fullKey.CreatedAt
		if fullKey.Name != "" {
			key.Name = fullKey.Name
		}
		if fullKey.Expires != 0 {
			key.Expires = fullKey.Expires
		}
	}

	return &key, nil
}

// DeleteApiKeyForServiceUser deletes an API key belonging to a service user.
func (c *Client) DeleteApiKeyForServiceUser(ctx context.Context, req DeleteApiKeyForServiceUserRequest) error {
	payload := map[string]interface{}{
		"type":   "DeleteApiKeyForServiceUser",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// isHex returns true if s is a valid hex string (for ObjectId detection).
func isHex(s string) bool {
	if len(s) != 24 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// GitProviderAccount CRUD

func (c *Client) CreateGitProviderAccount(ctx context.Context, req CreateGitProviderAccountRequest) (*GitProviderAccount, error) {
	payload := map[string]interface{}{
		"type":   "CreateGitProviderAccount",
		"params": map[string]interface{}{"account": req},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var account GitProviderAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) GetGitProviderAccount(ctx context.Context, id string) (*GitProviderAccount, error) {
	payload := map[string]interface{}{
		"type":   "GetGitProviderAccount",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var account GitProviderAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) ListGitProviderAccounts(ctx context.Context) ([]GitProviderAccount, error) {
	payload := map[string]interface{}{
		"type":   "ListGitProviderAccounts",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var accounts []GitProviderAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return accounts, nil
}

func (c *Client) UpdateGitProviderAccount(ctx context.Context, id string, req CreateGitProviderAccountRequest) (*GitProviderAccount, error) {
	payload := map[string]interface{}{
		"type":   "UpdateGitProviderAccount",
		"params": map[string]interface{}{"id": id, "account": req},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var account GitProviderAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) DeleteGitProviderAccount(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteGitProviderAccount",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(strings.ToLower(string(body)), "no account found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ResolveGitAccountUsername resolves a git account identifier to a username.
// If accountID is a 24-char hex ObjectId it fetches the GitProviderAccount and
// returns its Username. Otherwise the value is already a username and returned as-is.
func (c *Client) ResolveGitAccountUsername(ctx context.Context, accountID string) (string, error) {
	if accountID == "" {
		return "", nil
	}
	if len(accountID) == 24 && isHex(accountID) {
		account, err := c.GetGitProviderAccount(ctx, accountID)
		if err != nil {
			return "", fmt.Errorf("failed to resolve git provider account %q: %w", accountID, err)
		}
		return account.Username, nil
	}
	return accountID, nil
}

// ResolveGitAccountFull resolves a git account identifier to the full GitProviderAccount.
// Returns nil when accountID is empty, not a 24-char hex ObjectId, or the lookup fails.
func (c *Client) ResolveGitAccountFull(ctx context.Context, accountID string) *GitProviderAccount {
	if len(accountID) != 24 || !isHex(accountID) {
		return nil
	}
	account, err := c.GetGitProviderAccount(ctx, accountID)
	if err != nil {
		return nil
	}
	return account
}

// ResolveGitAccountID resolves a git account username (as returned by the API) back
// to the provider account ObjectId. It lists all provider accounts and tries an exact
// (domain + username) match first, then falls back to username-only to handle cases
// where the domain stored by the API does not exactly match the account's registered
// domain (e.g. custom-hosted git providers).
func (c *Client) ResolveGitAccountID(ctx context.Context, domain, username string) string {
	if username == "" {
		return ""
	}
	accounts, err := c.ListGitProviderAccounts(ctx)
	if err != nil {
		return ""
	}
	for _, a := range accounts {
		if a.Username == username && a.Domain == domain {
			return a.ID.OID
		}
	}
	// Fallback: match by username alone so custom-domain accounts are found even when
	// git_provider returned by the API differs from the account's registered domain.
	for _, a := range accounts {
		if a.Username == username {
			return a.ID.OID
		}
	}
	return ""
}

// ResolveDockerRegistryAccountID resolves a docker registry account domain+username
// back to its ObjectId. It lists all accounts and tries an exact domain+username match
// first, then falls back to username-only.
func (c *Client) ResolveDockerRegistryAccountID(ctx context.Context, domain, username string) string {
	if username == "" {
		return ""
	}
	accounts, err := c.ListDockerRegistryAccounts(ctx)
	if err != nil {
		return ""
	}
	for _, a := range accounts {
		if a.Username == username && a.Domain == domain {
			return a.ID.OID
		}
	}
	for _, a := range accounts {
		if a.Username == username {
			return a.ID.OID
		}
	}
	return ""
}

// DockerRegistryAccount CRUD

func (c *Client) CreateDockerRegistryAccount(ctx context.Context, req CreateDockerRegistryAccountRequest) (*DockerRegistryAccount, error) {
	payload := map[string]interface{}{
		"type":   "CreateDockerRegistryAccount",
		"params": map[string]interface{}{"account": req},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var account DockerRegistryAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) GetDockerRegistryAccount(ctx context.Context, id string) (*DockerRegistryAccount, error) {
	payload := map[string]interface{}{
		"type":   "GetDockerRegistryAccount",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var account DockerRegistryAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) ListDockerRegistryAccounts(ctx context.Context) ([]DockerRegistryAccount, error) {
	payload := map[string]interface{}{
		"type":   "ListDockerRegistryAccounts",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var accounts []DockerRegistryAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return accounts, nil
}

func (c *Client) UpdateDockerRegistryAccount(ctx context.Context, id string, req CreateDockerRegistryAccountRequest) (*DockerRegistryAccount, error) {
	payload := map[string]interface{}{
		"type":   "UpdateDockerRegistryAccount",
		"params": map[string]interface{}{"id": id, "account": req},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var account DockerRegistryAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &account, nil
}

func (c *Client) DeleteDockerRegistryAccount(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteDockerRegistryAccount",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(strings.ToLower(string(body)), "no account found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GitRepository CRUD

func (c *Client) CreateGitRepository(ctx context.Context, req CreateGitRepositoryRequest) (*GitRepository, error) {
	payload := map[string]interface{}{
		"type":   "CreateRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var repo GitRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &repo, nil
}

func (c *Client) GetGitRepository(ctx context.Context, idOrName string) (*GitRepository, error) {
	payload := map[string]interface{}{
		"type":   "GetRepo",
		"params": map[string]interface{}{"repo": idOrName},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var repo GitRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &repo, nil
}

func (c *Client) UpdateGitRepository(ctx context.Context, req UpdateGitRepositoryRequest) (*GitRepository, error) {
	payload := map[string]interface{}{
		"type":   "UpdateRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var repo GitRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &repo, nil
}

func (c *Client) RenameGitRepository(ctx context.Context, req RenameGitRepositoryRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteGitRepository(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteRepo",
		"params": DeleteGitRepositoryRequest{ID: id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) ListGitRepositories(ctx context.Context) ([]GitRepository, error) {
	payload := map[string]interface{}{
		"type":   "ListFullRepos",
		"params": map[string]interface{}{"query": map[string]interface{}{}},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var repos []GitRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return repos, nil
}

// BuildRepo builds the target repo using its attached builder.
func (c *Client) BuildRepo(ctx context.Context, req BuildRepoRequest) error {
	payload := map[string]interface{}{
		"type":   "BuildRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CloneRepo clones the target repo onto its attached server.
func (c *Client) CloneRepo(ctx context.Context, req CloneRepoRequest) error {
	payload := map[string]interface{}{
		"type":   "CloneRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PullRepo pulls the latest changes for the target repo on its attached server.
func (c *Client) PullRepo(ctx context.Context, req PullRepoRequest) error {
	payload := map[string]interface{}{
		"type":   "PullRepo",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Stack CRUD

func (c *Client) CreateStack(ctx context.Context, req CreateStackRequest) (*Stack, error) {
	payload := map[string]interface{}{
		"type":   "CreateStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var stack Stack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &stack, nil
}

func (c *Client) GetStack(ctx context.Context, nameOrID string) (*Stack, error) {
	payload := map[string]interface{}{
		"type":   "GetStack",
		"params": map[string]interface{}{"stack": nameOrID},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var stack Stack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &stack, nil
}

func (c *Client) UpdateStack(ctx context.Context, req UpdateStackRequest) (*Stack, error) {
	payload := map[string]interface{}{
		"type":   "UpdateStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var stack Stack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &stack, nil
}

func (c *Client) RenameStack(ctx context.Context, req RenameStackRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteStack(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteStack",
		"params": DeleteStackRequest{ID: id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	return nil
}

func (c *Client) ListStacks(ctx context.Context) ([]Stack, error) {
	payload := map[string]interface{}{
		"type":   "ListFullStacks",
		"params": map[string]interface{}{"query": map[string]interface{}{}},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var stacks []Stack
	if err := json.NewDecoder(resp.Body).Decode(&stacks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return stacks, nil
}

// Stack execute actions

// StartStack starts the target stack (docker compose start).
func (c *Client) StartStack(ctx context.Context, req StartStackRequest) error {
	payload := map[string]interface{}{
		"type":   "StartStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// StopStack stops the target stack (docker compose stop).
func (c *Client) StopStack(ctx context.Context, req StopStackRequest) error {
	payload := map[string]interface{}{
		"type":   "StopStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PauseStack pauses the target stack (docker compose pause).
func (c *Client) PauseStack(ctx context.Context, req PauseStackRequest) error {
	payload := map[string]interface{}{
		"type":   "PauseStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeployStack deploys the target stack (docker compose up).
func (c *Client) DeployStack(ctx context.Context, req DeployStackRequest) error {
	payload := map[string]interface{}{
		"type":   "DeployStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DestroyStack destroys the target stack (docker compose down).
func (c *Client) DestroyStackAction(ctx context.Context, req DestroyStackActionRequest) error {
	payload := map[string]interface{}{
		"type":   "DestroyStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Server read methods

func (c *Client) GetServer(ctx context.Context, nameOrID string) (*Server, error) {
	payload := map[string]interface{}{
		"type":   "GetServer",
		"params": map[string]interface{}{"server": nameOrID},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var server Server
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &server, nil
}

func (c *Client) ListServers(ctx context.Context) ([]Server, error) {
	payload := map[string]interface{}{
		"type":   "ListFullServers",
		"params": map[string]interface{}{"query": map[string]interface{}{}},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var servers []Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return servers, nil
}

// CreateServer creates a new server.
func (c *Client) CreateServer(ctx context.Context, req CreateServerRequest) (*Server, error) {
	payload := map[string]interface{}{
		"type":   "CreateServer",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var server Server
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &server, nil
}

// UpdateServer updates an existing server's config.
func (c *Client) UpdateServer(ctx context.Context, req UpdateServerRequest) (*Server, error) {
	payload := map[string]interface{}{
		"type":   "UpdateServer",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var server Server
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &server, nil
}

// UpdateServerPublicKey updates the public key for the given server.
func (c *Client) UpdateServerPublicKey(ctx context.Context, serverID, publicKey string) error {
	payload := map[string]interface{}{
		"type": "UpdateServerPublicKey",
		"params": UpdateServerPublicKeyRequest{
			Server:    serverID,
			PublicKey: publicKey,
		},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// RenameServer renames the server with the given ID.
func (c *Client) RenameServer(ctx context.Context, req RenameServerRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameServer",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteServer deletes the server with the given ID.
func (c *Client) DeleteServer(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteServer",
		"params": DeleteServerRequest{ID: id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	return nil
}

// Server execute actions

// PruneBuildx prunes the docker buildx cache on the target server.
func (c *Client) PruneBuildx(ctx context.Context, req PruneBuildxRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneBuildx",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneContainers prunes the docker containers on the target server.
func (c *Client) PruneContainers(ctx context.Context, req PruneContainersRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneContainers",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneDockerBuilders prunes the docker builders on the target server.
func (c *Client) PruneDockerBuilders(ctx context.Context, req PruneDockerBuildersRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneDockerBuilders",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneImages prunes the docker images on the target server.
func (c *Client) PruneImages(ctx context.Context, req PruneImagesRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneImages",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneNetworks prunes the docker networks on the target server.
func (c *Client) PruneNetworks(ctx context.Context, req PruneNetworksRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneNetworks",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneSystem prunes the docker system on the target server, including volumes.
func (c *Client) PruneSystem(ctx context.Context, req PruneSystemRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneSystem",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PruneVolumes prunes the docker volumes on the target server.
func (c *Client) PruneVolumes(ctx context.Context, req PruneVolumesRequest) error {
	payload := map[string]interface{}{
		"type":   "PruneVolumes",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Deployment execute actions

// StartDeployment starts the target deployment.
func (c *Client) StartDeployment(ctx context.Context, req StartDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "StartDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PullDeployment pulls the latest image for the target deployment.
func (c *Client) PullDeployment(ctx context.Context, req PullDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "PullDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Action execute actions

// RunAction runs the target action.
func (c *Client) RunAction(ctx context.Context, req RunActionRequest) error {
	payload := map[string]interface{}{
		"type":   "RunAction",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Build execute actions

// RunBuild runs the target build.
func (c *Client) RunBuild(ctx context.Context, req RunBuildRequest) error {
	payload := map[string]interface{}{
		"type":   "RunBuild",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Procedure execute actions

// RunProcedure runs the target procedure.
func (c *Client) RunProcedure(ctx context.Context, req RunProcedureRequest) error {
	payload := map[string]interface{}{
		"type":   "RunProcedure",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Deploy (re)deploys the container for the target deployment.
func (c *Client) Deploy(ctx context.Context, req DeployRequest) error {
	payload := map[string]interface{}{
		"type":   "Deploy",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// StopDeployment stops the container for the target deployment.
func (c *Client) StopDeployment(ctx context.Context, req StopDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "StopDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DestroyDeployment stops and removes the container for the target deployment.
func (c *Client) DestroyDeployment(ctx context.Context, req DestroyDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "DestroyDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// RestartDeployment restarts the container for the target deployment.
func (c *Client) RestartDeployment(ctx context.Context, req RestartDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "RestartDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PauseDeployment pauses the container for the target deployment.
func (c *Client) PauseDeployment(ctx context.Context, req PauseDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "PauseDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// UnpauseDeployment unpauses the container for the target deployment.
func (c *Client) UnpauseDeployment(ctx context.Context, req UnpauseDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "UnpauseDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// RestartStack restarts the target stack (docker compose restart).
func (c *Client) RestartStack(ctx context.Context, req RestartStackRequest) error {
	payload := map[string]interface{}{
		"type":   "RestartStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// UnpauseStack unpauses the target stack (docker compose unpause).
func (c *Client) UnpauseStack(ctx context.Context, req UnpauseStackRequest) error {
	payload := map[string]interface{}{
		"type":   "UnpauseStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PullStack pulls images for the target stack (docker compose pull).
func (c *Client) PullStack(ctx context.Context, req PullStackRequest) error {
	payload := map[string]interface{}{
		"type":   "PullStack",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeployStackIfChanged checks deployed contents vs latest and only deploys if changed.
func (c *Client) DeployStackIfChanged(ctx context.Context, req DeployStackIfChangedRequest) error {
	payload := map[string]interface{}{
		"type":   "DeployStackIfChanged",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// RunStackService runs a one-time command against a service using docker compose run.
func (c *Client) RunStackService(ctx context.Context, req RunStackServiceRequest) error {
	payload := map[string]interface{}{
		"type":   "RunStackService",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// TestAlerter tests the alerter's ability to reach the configured endpoint.
func (c *Client) TestAlerter(ctx context.Context, req TestAlerterRequest) error {
	payload := map[string]interface{}{
		"type":   "TestAlerter",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Builder CRUD

func (c *Client) CreateBuilder(ctx context.Context, req CreateBuilderRequest) (*Builder, error) {
	payload := map[string]interface{}{
		"type":   "CreateBuilder",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var builder Builder
	if err := json.NewDecoder(resp.Body).Decode(&builder); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &builder, nil
}

func (c *Client) GetBuilder(ctx context.Context, idOrName string) (*Builder, error) {
	payload := map[string]interface{}{
		"type":   "GetBuilder",
		"params": map[string]interface{}{"builder": idOrName},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var builder Builder
	if err := json.NewDecoder(resp.Body).Decode(&builder); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &builder, nil
}

func (c *Client) UpdateBuilder(ctx context.Context, req UpdateBuilderRequest) (*Builder, error) {
	payload := map[string]interface{}{
		"type":   "UpdateBuilder",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var builder Builder
	if err := json.NewDecoder(resp.Body).Decode(&builder); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &builder, nil
}

func (c *Client) RenameBuilder(ctx context.Context, req RenameBuilderRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameBuilder",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteBuilder(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteBuilder",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) ListBuilders(ctx context.Context) ([]Builder, error) {
	payload := map[string]interface{}{
		"type":   "ListFullBuilders",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var builders []Builder
	if err := json.NewDecoder(resp.Body).Decode(&builders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return builders, nil
}

func (c *Client) CreateAlerter(ctx context.Context, req CreateAlerterRequest) (*Alerter, error) {
	payload := map[string]interface{}{
		"type":   "CreateAlerter",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var alerter Alerter
	if err := json.NewDecoder(resp.Body).Decode(&alerter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &alerter, nil
}

func (c *Client) GetAlerter(ctx context.Context, idOrName string) (*Alerter, error) {
	payload := map[string]interface{}{
		"type":   "GetAlerter",
		"params": map[string]interface{}{"alerter": idOrName},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var alerter Alerter
	if err := json.NewDecoder(resp.Body).Decode(&alerter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &alerter, nil
}

func (c *Client) UpdateAlerter(ctx context.Context, req UpdateAlerterRequest) (*Alerter, error) {
	payload := map[string]interface{}{
		"type":   "UpdateAlerter",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var alerter Alerter
	if err := json.NewDecoder(resp.Body).Decode(&alerter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &alerter, nil
}

func (c *Client) RenameAlerter(ctx context.Context, req RenameAlerterRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameAlerter",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteAlerter(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteAlerter",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) ListAlerters(ctx context.Context) ([]Alerter, error) {
	payload := map[string]interface{}{
		"type":   "ListAlerters",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var alerters []Alerter
	if err := json.NewDecoder(resp.Body).Decode(&alerters); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return alerters, nil
}

// Action CRUD

func (c *Client) CreateAction(ctx context.Context, req CreateActionRequest) (*Action, error) {
	payload := map[string]interface{}{
		"type":   "CreateAction",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var action Action
	if err := json.NewDecoder(resp.Body).Decode(&action); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &action, nil
}

func (c *Client) GetAction(ctx context.Context, idOrName string) (*Action, error) {
	payload := map[string]interface{}{
		"type":   "GetAction",
		"params": map[string]interface{}{"action": idOrName},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var action Action
	if err := json.NewDecoder(resp.Body).Decode(&action); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &action, nil
}

func (c *Client) UpdateAction(ctx context.Context, req UpdateActionRequest) (*Action, error) {
	payload := map[string]interface{}{
		"type":   "UpdateAction",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var action Action
	if err := json.NewDecoder(resp.Body).Decode(&action); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &action, nil
}

func (c *Client) RenameAction(ctx context.Context, req RenameActionRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameAction",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteAction(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteAction",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) ListActions(ctx context.Context) ([]Action, error) {
	payload := map[string]interface{}{
		"type":   "ListActions",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var actions []Action
	if err := json.NewDecoder(resp.Body).Decode(&actions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return actions, nil
}

// CreateBuild creates a new build and returns the created build.
func (c *Client) CreateBuild(ctx context.Context, req CreateBuildRequest) (*Build, error) {
	payload := map[string]interface{}{
		"type":   "CreateBuild",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var build Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &build, nil
}

// GetBuild retrieves a build by id or name. Returns nil if not found.
func (c *Client) GetBuild(ctx context.Context, idOrName string) (*Build, error) {
	payload := map[string]interface{}{
		"type": "GetBuild",
		"params": map[string]interface{}{
			"build": idOrName,
		},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var build Build
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &build, nil
}

// UpdateBuild updates a build and returns the updated build.
func (c *Client) UpdateBuild(ctx context.Context, req UpdateBuildRequest) (*Build, error) {
	payload := map[string]interface{}{
		"type":   "UpdateBuild",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var build Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &build, nil
}

// RenameBuild renames a build by id.
func (c *Client) RenameBuild(ctx context.Context, req RenameBuildRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameBuild",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteBuild deletes a build by id.
func (c *Client) DeleteBuild(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type": "DeleteBuild",
		"params": map[string]interface{}{
			"id": id,
		},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListBuilds returns all builds.
func (c *Client) ListBuilds(ctx context.Context) ([]Build, error) {
	payload := map[string]interface{}{
		"type":   "ListBuilds",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var builds []Build
	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return builds, nil
}

// CreateDeployment creates a new deployment and returns the created deployment.
func (c *Client) CreateDeployment(ctx context.Context, req CreateDeploymentRequest) (*Deployment, error) {
	payload := map[string]interface{}{
		"type":   "CreateDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var deployment Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &deployment, nil
}

// GetDeployment retrieves a deployment by id or name. Returns nil if not found.
func (c *Client) GetDeployment(ctx context.Context, idOrName string) (*Deployment, error) {
	payload := map[string]interface{}{
		"type": "GetDeployment",
		"params": map[string]interface{}{
			"deployment": idOrName,
		},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var deployment Deployment
	if err := json.Unmarshal(body, &deployment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &deployment, nil
}

// UpdateDeployment updates a deployment and returns the updated deployment.
func (c *Client) UpdateDeployment(ctx context.Context, req UpdateDeploymentRequest) (*Deployment, error) {
	payload := map[string]interface{}{
		"type":   "UpdateDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var deployment Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &deployment, nil
}

// RenameDeployment renames a deployment by id.
func (c *Client) RenameDeployment(ctx context.Context, req RenameDeploymentRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameDeployment",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteDeployment deletes a deployment by id.
func (c *Client) DeleteDeployment(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type": "DeleteDeployment",
		"params": map[string]interface{}{
			"id": id,
		},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListDeployments returns all deployments.
func (c *Client) ListDeployments(ctx context.Context) ([]Deployment, error) {
	payload := map[string]interface{}{
		"type":   "ListDeployments",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var deployments []Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return deployments, nil
}

// CreateProcedure creates a new procedure.
func (c *Client) CreateProcedure(ctx context.Context, req CreateProcedureRequest) (*Procedure, error) {
	payload := map[string]interface{}{
		"type":   "CreateProcedure",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var proc Procedure
	if err := json.NewDecoder(resp.Body).Decode(&proc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &proc, nil
}

// GetProcedure retrieves a procedure by ID or name.
func (c *Client) GetProcedure(ctx context.Context, idOrName string) (*Procedure, error) {
	payload := map[string]interface{}{
		"type": "GetProcedure",
		"params": map[string]interface{}{
			"procedure": idOrName,
		},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var proc Procedure
	if err := json.NewDecoder(resp.Body).Decode(&proc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &proc, nil
}

// UpdateProcedure updates an existing procedure's configuration.
func (c *Client) UpdateProcedure(ctx context.Context, req UpdateProcedureRequest) (*Procedure, error) {
	payload := map[string]interface{}{
		"type":   "UpdateProcedure",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var proc Procedure
	if err := json.NewDecoder(resp.Body).Decode(&proc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &proc, nil
}

// RenameProcedure renames a procedure by ID.
func (c *Client) RenameProcedure(ctx context.Context, req RenameProcedureRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameProcedure",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteProcedure deletes a procedure by ID.
func (c *Client) DeleteProcedure(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type": "DeleteProcedure",
		"params": map[string]interface{}{
			"id": id,
		},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListProcedures returns all procedures.
func (c *Client) ListProcedures(ctx context.Context) ([]Procedure, error) {
	payload := map[string]interface{}{
		"type":   "ListProcedures",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var procedures []Procedure
	if err := json.NewDecoder(resp.Body).Decode(&procedures); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return procedures, nil
}

// ResourceSync CRUD

// CreateResourceSync creates a new resource sync.
func (c *Client) CreateResourceSync(ctx context.Context, req CreateResourceSyncRequest) (*ResourceSync, error) {
	payload := map[string]interface{}{
		"type":   "CreateResourceSync",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var rs ResourceSync
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &rs, nil
}

// GetResourceSync retrieves a resource sync by id or name.
func (c *Client) GetResourceSync(ctx context.Context, id string) (*ResourceSync, error) {
	payload := map[string]interface{}{
		"type":   "GetResourceSync",
		"params": map[string]interface{}{"sync": id},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var rs ResourceSync
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &rs, nil
}

// ListResourceSyncs returns all resource syncs.
func (c *Client) ListResourceSyncs(ctx context.Context) ([]ResourceSync, error) {
	payload := map[string]interface{}{
		"type":   "ListResourceSyncs",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var syncs []ResourceSync
	if err := json.NewDecoder(resp.Body).Decode(&syncs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return syncs, nil
}

// UpdateResourceSync updates an existing resource sync.
func (c *Client) UpdateResourceSync(ctx context.Context, req UpdateResourceSyncRequest) (*ResourceSync, error) {
	payload := map[string]interface{}{
		"type":   "UpdateResourceSync",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var rs ResourceSync
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &rs, nil
}

// RenameResourceSync renames a resource sync by id or name.
func (c *Client) RenameResourceSync(ctx context.Context, req RenameResourceSyncRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameResourceSync",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteResourceSync deletes a resource sync by id or name.
func (c *Client) DeleteResourceSync(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteResourceSync",
		"params": map[string]interface{}{"id": id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Swarm CRUD

func (c *Client) CreateSwarm(ctx context.Context, req CreateSwarmRequest) (*Swarm, error) {
	payload := map[string]interface{}{
		"type":   "CreateSwarm",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var swarm Swarm
	if err := json.NewDecoder(resp.Body).Decode(&swarm); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &swarm, nil
}

func (c *Client) GetSwarm(ctx context.Context, nameOrID string) (*Swarm, error) {
	payload := map[string]interface{}{
		"type":   "GetSwarm",
		"params": map[string]interface{}{"swarm": nameOrID},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	var swarm Swarm
	if err := json.NewDecoder(resp.Body).Decode(&swarm); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &swarm, nil
}

func (c *Client) UpdateSwarm(ctx context.Context, req UpdateSwarmRequest) (*Swarm, error) {
	payload := map[string]interface{}{
		"type":   "UpdateSwarm",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var swarm Swarm
	if err := json.NewDecoder(resp.Body).Decode(&swarm); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &swarm, nil
}

func (c *Client) RenameSwarm(ctx context.Context, req RenameSwarmRequest) error {
	payload := map[string]interface{}{
		"type":   "RenameSwarm",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) DeleteSwarm(ctx context.Context, id string) error {
	payload := map[string]interface{}{
		"type":   "DeleteSwarm",
		"params": DeleteSwarmRequest{ID: id},
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if strings.Contains(strings.ToLower(bodyStr), "did not find") || strings.Contains(strings.ToLower(bodyStr), "not found") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, bodyStr)
	}
	return nil
}

func (c *Client) ListSwarms(ctx context.Context) ([]Swarm, error) {
	payload := map[string]interface{}{
		"type":   "ListFullSwarms",
		"params": map[string]interface{}{"query": map[string]interface{}{}},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var swarms []Swarm
	if err := json.NewDecoder(resp.Body).Decode(&swarms); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return swarms, nil
}

// RunSync executes a resource sync.
func (c *Client) RunSync(ctx context.Context, req RunSyncRequest) error {
	payload := map[string]interface{}{
		"type":   "RunSync",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/execute", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CreateNetwork creates a docker network on the specified server.
func (c *Client) CreateNetwork(ctx context.Context, req CreateNetworkRequest) error {
	payload := map[string]interface{}{
		"type":   "CreateNetwork",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListDockerNetworks lists docker networks on the specified server.
func (c *Client) ListDockerNetworks(ctx context.Context, server string) ([]NetworkListItem, error) {
	payload := map[string]interface{}{
		"type":   "ListDockerNetworks",
		"params": map[string]interface{}{"server": server},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var networks []NetworkListItem
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return networks, nil
}

// OnboardingKey CRUD

// ListOnboardingKeys returns all onboarding keys.
func (c *Client) ListOnboardingKeys(ctx context.Context) ([]OnboardingKey, error) {
	payload := map[string]interface{}{
		"type":   "ListOnboardingKeys",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var keys []OnboardingKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return keys, nil
}

// GetOnboardingKey retrieves an onboarding key by its public key, using list+filter.
func (c *Client) GetOnboardingKey(ctx context.Context, publicKey string) (*OnboardingKey, error) {
	keys, err := c.ListOnboardingKeys(ctx)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if key.PublicKey == publicKey {
			return &key, nil
		}
	}
	return nil, nil
}

// CreateOnboardingKey creates a new onboarding key.
func (c *Client) CreateOnboardingKey(ctx context.Context, req CreateOnboardingKeyRequest) (*CreateOnboardingKeyResponse, error) {
	payload := map[string]interface{}{
		"type":   "CreateOnboardingKey",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var createResp CreateOnboardingKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &createResp, nil
}

// UpdateOnboardingKey updates an existing onboarding key.
func (c *Client) UpdateOnboardingKey(ctx context.Context, req UpdateOnboardingKeyRequest) (*OnboardingKey, error) {
	payload := map[string]interface{}{
		"type":   "UpdateOnboardingKey",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var key OnboardingKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &key, nil
}

// DeleteOnboardingKey deletes an onboarding key by its public key.
func (c *Client) DeleteOnboardingKey(ctx context.Context, req DeleteOnboardingKeyRequest) error {
	payload := map[string]interface{}{
		"type":   "DeleteOnboardingKey",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "not found") || strings.Contains(strings.ToLower(bodyStr), "did not find") {
			return nil
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetVersion returns the version of the Komodo Core API.
func (c *Client) GetVersion(ctx context.Context) (*GetVersionResponse, error) {
	payload := map[string]interface{}{
		"type":   "GetVersion",
		"params": map[string]interface{}{},
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var result GetVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

// Terminal operations.

// CreateTerminal creates a terminal session for a server.
func (c *Client) CreateTerminal(ctx context.Context, req CreateTerminalRequest) (*Terminal, error) {
	payload := map[string]interface{}{
		"type":   "CreateTerminal",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var terminal Terminal
	if err := json.NewDecoder(resp.Body).Decode(&terminal); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &terminal, nil
}

// ListTerminals lists terminals, optionally filtered by target.
func (c *Client) ListTerminals(ctx context.Context, req ListTerminalsRequest) ([]Terminal, error) {
	payload := map[string]interface{}{
		"type":   "ListTerminals",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/read", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var terminals []Terminal
	if err := json.NewDecoder(resp.Body).Decode(&terminals); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return terminals, nil
}

// DeleteTerminal deletes the named terminal on the given target.
func (c *Client) DeleteTerminal(ctx context.Context, req DeleteTerminalRequest) error {
	payload := map[string]interface{}{
		"type":   "DeleteTerminal",
		"params": req,
	}
	resp, err := c.doRequest(ctx, "/write", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
