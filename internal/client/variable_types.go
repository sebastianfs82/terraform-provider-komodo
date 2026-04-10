// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

type Variable struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	IsSecret    bool   `json:"is_secret"`
}

type CreateVariableRequest struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	IsSecret    bool   `json:"is_secret"`
}

type DeleteVariableRequest struct {
	ID string `json:"id"`
}

// ListVariablesRequest and other CRUD request/response types can be added as needed.
