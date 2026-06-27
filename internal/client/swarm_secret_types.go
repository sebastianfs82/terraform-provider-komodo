// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// CreateSwarmSecretRequest is the request body for CreateSwarmSecret execute action.
type CreateSwarmSecretRequest struct {
	Swarm          string   `json:"swarm"`
	Name           string   `json:"name"`
	Data           string   `json:"data"`
	Driver         *string  `json:"driver,omitempty"`
	Labels         []string `json:"labels"`
	TemplateDriver *string  `json:"template_driver,omitempty"`
}

// RotateSwarmSecretRequest is the request body for RotateSwarmSecret execute action.
type RotateSwarmSecretRequest struct {
	Swarm  string `json:"swarm"`
	Secret string `json:"secret"`
	Data   string `json:"data"`
}

// RemoveSwarmSecretsRequest is the request body for RemoveSwarmSecrets execute action.
type RemoveSwarmSecretsRequest struct {
	Swarm   string   `json:"swarm"`
	Secrets []string `json:"secrets"`
}

// ListSwarmSecretsRequest is the request body for ListSwarmSecrets read action.
type ListSwarmSecretsRequest struct {
	Swarm string `json:"swarm"`
}

// SwarmSecretListItem represents one item from ListSwarmSecrets response.
type SwarmSecretListItem struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}
