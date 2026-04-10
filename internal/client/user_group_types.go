// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

type UserGroup struct {
	ID        OID                    `json:"_id"` // Use OID struct
	Name      string                 `json:"name"`
	Everyone  bool                   `json:"everyone"`
	Users     []string               `json:"users"`
	All       map[string]interface{} `json:"all"`
	UpdatedAt int64                  `json:"updated_at"`
}

type CreateUserGroupRequest struct {
	Name     string                 `json:"name"`
	Everyone bool                   `json:"everyone"`
	Users    []string               `json:"users"`
	All      map[string]interface{} `json:"all"`
}

type RenameUserGroupRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DeleteUserGroupRequest struct {
	ID string `json:"id"`
}
