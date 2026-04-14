// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
	Owner string `json:"owner,omitempty"`
}

type DeleteTagRequest struct {
	ID string `json:"id"`
}

type RenameTagRequest struct {
	ID      string `json:"id"`
	OldName string `json:"old_name,omitempty"`
	Name    string `json:"name"`
}

type UpdateTagColorRequest struct {
	Tag   string `json:"tag"`
	Color string `json:"color"`
}
