// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// UserConfig holds the discriminated user configuration returned by the API.
type UserConfig struct {
	Type string `json:"type"`
}

type User struct {
	ID            OID        `json:"_id"`
	Username      string     `json:"username"`
	Enabled       bool       `json:"enabled"`
	Admin         bool       `json:"admin"`
	CreateServers bool       `json:"create_server_permissions"`
	CreateBuilds  bool       `json:"create_build_permissions"`
	UpdatedAt     int64      `json:"updated_at"`
	Config        UserConfig `json:"config"`
}

type CreateLocalUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DeleteUserRequest struct {
	User string `json:"user"`
}

type FindUserRequest struct {
	User string `json:"user"`
}

type UpdateUserAdminRequest struct {
	UserID string `json:"user_id"`
	Admin  bool   `json:"admin"`
}

type UpdateUserBasePermissionsRequest struct {
	UserID        string `json:"user_id"`
	Enabled       *bool  `json:"enabled,omitempty"`
	CreateServers *bool  `json:"create_servers,omitempty"`
	CreateBuilds  *bool  `json:"create_builds,omitempty"`
}

type CreateServiceUserRequest struct {
	Username    string `json:"username"`
	Description string `json:"description"`
}

type UpdateServiceUserDescriptionRequest struct {
	Username    string `json:"username"`
	Description string `json:"description"`
}
