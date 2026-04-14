// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// Swarm represents a Komodo Docker Swarm resource.
type Swarm struct {
	ID     OID         `json:"_id"`
	Name   string      `json:"name"`
	Tags   []string    `json:"tags"`
	Config SwarmConfig `json:"config"`
}

// SwarmConfig is the configuration for a Komodo swarm.
type SwarmConfig struct {
	ServerIDs          []string            `json:"server_ids"`
	Links              []string            `json:"links"`
	AlertsEnabled      bool                `json:"send_unhealthy_alerts"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

// PartialSwarmConfig mirrors _PartialSwarmConfig on the API side.
// All fields are pointers with omitempty so only explicitly-set fields are
// sent in JSON — omitted fields are left untouched by the server.
type PartialSwarmConfig struct {
	ServerIDs          *[]string            `json:"server_ids,omitempty"`
	Links              *[]string            `json:"links,omitempty"`
	AlertsEnabled      *bool                `json:"send_unhealthy_alerts,omitempty"`
	MaintenanceWindows *[]MaintenanceWindow `json:"maintenance_windows,omitempty"`
}

// CreateSwarmRequest is the request body for CreateSwarm.
type CreateSwarmRequest struct {
	Name   string             `json:"name"`
	Config PartialSwarmConfig `json:"config"`
}

// UpdateSwarmRequest is the request body for UpdateSwarm.
type UpdateSwarmRequest struct {
	ID     string             `json:"id"`
	Config PartialSwarmConfig `json:"config"`
}

// RenameSwarmRequest is the payload for the RenameSwarm write API.
type RenameSwarmRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DeleteSwarmRequest is the request body for DeleteSwarm.
type DeleteSwarmRequest struct {
	ID string `json:"id"`
}
