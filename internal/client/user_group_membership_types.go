// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

type AddUserToUserGroupRequest struct {
	UserGroup string `json:"user_group"`
	User      string `json:"user"`
}

type RemoveUserFromUserGroupRequest struct {
	UserGroup string `json:"user_group"`
	User      string `json:"user"`
}

type SetEveryoneUserGroupRequest struct {
	UserGroup string `json:"user_group"`
	Everyone  bool   `json:"everyone"`
}
