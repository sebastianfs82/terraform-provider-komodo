// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
)

// ResourceTarget is a discriminated-union reference to a specific Komodo resource.
// It serialises as {"type": "<Variant>", "id": "<name>"}.
type ResourceTarget struct {
	Type string
	ID   string
}

func (r ResourceTarget) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{Type: r.Type, ID: r.ID})
}

func (r *ResourceTarget) UnmarshalJSON(data []byte) error {
	var v struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to decode ResourceTarget: %w", err)
	}
	r.Type = v.Type
	r.ID = v.ID
	return nil
}

// Alerter represents a Komodo alerter resource (Resource<AlerterConfig>).
type Alerter struct {
	ID     OID           `json:"_id"`
	Name   string        `json:"name"`
	Config AlerterConfig `json:"config"`
}

// AlerterConfig is the Komodo alerter configuration.
type AlerterConfig struct {
	Enabled            bool                `json:"enabled"`
	Endpoint           AlerterEndpoint     `json:"endpoint"`
	AlertTypes         []string            `json:"alert_types"`
	Resources          []ResourceTarget    `json:"resources"`
	ExceptResources    []ResourceTarget    `json:"except_resources"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

// AlerterEndpoint is a discriminated union of alerter endpoint types,
// serialised by the Komodo API as {"type": "<Variant>", "params": {...}}.
type AlerterEndpoint struct {
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

// GetCustomParams decodes the Params field as CustomAlerterEndpoint.
// Returns nil, nil when Type is not "Custom".
func (e *AlerterEndpoint) GetCustomParams() (*CustomAlerterEndpoint, error) {
	if e.Type != "Custom" {
		return nil, nil
	}
	var p CustomAlerterEndpoint
	if err := json.Unmarshal(e.Params, &p); err != nil {
		return nil, fmt.Errorf("failed to decode Custom alerter endpoint: %w", err)
	}
	return &p, nil
}

// GetSlackParams decodes the Params field as SlackAlerterEndpoint.
// Returns nil, nil when Type is not "Slack".
func (e *AlerterEndpoint) GetSlackParams() (*SlackAlerterEndpoint, error) {
	if e.Type != "Slack" {
		return nil, nil
	}
	var p SlackAlerterEndpoint
	if err := json.Unmarshal(e.Params, &p); err != nil {
		return nil, fmt.Errorf("failed to decode Slack alerter endpoint: %w", err)
	}
	return &p, nil
}

// GetDiscordParams decodes the Params field as DiscordAlerterEndpoint.
// Returns nil, nil when Type is not "Discord".
func (e *AlerterEndpoint) GetDiscordParams() (*DiscordAlerterEndpoint, error) {
	if e.Type != "Discord" {
		return nil, nil
	}
	var p DiscordAlerterEndpoint
	if err := json.Unmarshal(e.Params, &p); err != nil {
		return nil, fmt.Errorf("failed to decode Discord alerter endpoint: %w", err)
	}
	return &p, nil
}

// GetNtfyParams decodes the Params field as NtfyAlerterEndpoint.
// Returns nil, nil when Type is not "Ntfy".
func (e *AlerterEndpoint) GetNtfyParams() (*NtfyAlerterEndpoint, error) {
	if e.Type != "Ntfy" {
		return nil, nil
	}
	var p NtfyAlerterEndpoint
	if err := json.Unmarshal(e.Params, &p); err != nil {
		return nil, fmt.Errorf("failed to decode Ntfy alerter endpoint: %w", err)
	}
	return &p, nil
}

// GetPushoverParams decodes the Params field as PushoverAlerterEndpoint.
// Returns nil, nil when Type is not "Pushover".
func (e *AlerterEndpoint) GetPushoverParams() (*PushoverAlerterEndpoint, error) {
	if e.Type != "Pushover" {
		return nil, nil
	}
	var p PushoverAlerterEndpoint
	if err := json.Unmarshal(e.Params, &p); err != nil {
		return nil, fmt.Errorf("failed to decode Pushover alerter endpoint: %w", err)
	}
	return &p, nil
}

// CustomAlerterEndpoint is the configuration for a Custom alerter endpoint.
type CustomAlerterEndpoint struct {
	URL string `json:"url"`
}

// SlackAlerterEndpoint is the configuration for a Slack alerter endpoint.
type SlackAlerterEndpoint struct {
	URL string `json:"url"`
}

// DiscordAlerterEndpoint is the configuration for a Discord alerter endpoint.
type DiscordAlerterEndpoint struct {
	URL string `json:"url"`
}

// NtfyAlerterEndpoint is the configuration for a Ntfy alerter endpoint.
type NtfyAlerterEndpoint struct {
	URL   string  `json:"url"`
	Email *string `json:"email"`
}

// PushoverAlerterEndpoint is the configuration for a Pushover alerter endpoint.
type PushoverAlerterEndpoint struct {
	URL string `json:"url"`
}

// AlerterEndpointInput is used for write operations.
type AlerterEndpointInput struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

// PartialAlerterConfigInput is sent in create and update requests.
type PartialAlerterConfigInput struct {
	Enabled            *bool                 `json:"enabled,omitempty"`
	Endpoint           *AlerterEndpointInput `json:"endpoint,omitempty"`
	AlertTypes         *[]string             `json:"alert_types,omitempty"`
	Resources          *[]ResourceTarget     `json:"resources,omitempty"`
	ExceptResources    *[]ResourceTarget     `json:"except_resources,omitempty"`
	MaintenanceWindows *[]MaintenanceWindow  `json:"maintenance_windows,omitempty"`
}

// CreateAlerterRequest is the params for the CreateAlerter API call.
type CreateAlerterRequest struct {
	Name   string                    `json:"name"`
	Config PartialAlerterConfigInput `json:"config"`
}

// UpdateAlerterRequest is the params for the UpdateAlerter API call.
type UpdateAlerterRequest struct {
	ID     string                    `json:"id"`
	Config PartialAlerterConfigInput `json:"config"`
}

// RenameAlerterRequest is the params for the RenameAlerter API call.
type RenameAlerterRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TestAlerterRequest is the request body for the TestAlerter execute action.
type TestAlerterRequest struct {
	Alerter string `json:"alerter"`
}
