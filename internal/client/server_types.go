// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// Server represents a Komodo server resource.
type Server struct {
	ID     OID          `json:"_id"`
	Name   string       `json:"name"`
	Config ServerConfig `json:"config"`
}

// ServerConfig is the configuration for a Komodo server.
type ServerConfig struct {
	// Connection
	Address         string `json:"address"`
	InsecureTLS     bool   `json:"insecure_tls"`
	ExternalAddress string `json:"external_address"`
	Region          string `json:"region"`

	// Behaviour
	Enabled         bool   `json:"enabled"`
	AutoRotateKeys  bool   `json:"auto_rotate_keys"`
	Passkey         string `json:"passkey"`
	AutoPrune       bool   `json:"auto_prune"`
	StatsMonitoring bool   `json:"stats_monitoring"`

	// Mounts / links
	IgnoreMounts []string `json:"ignore_mounts"`
	Links        []string `json:"links"`

	// Alert flags
	SendUnreachableAlerts     bool `json:"send_unreachable_alerts"`
	SendCPUAlerts             bool `json:"send_cpu_alerts"`
	SendMemAlerts             bool `json:"send_mem_alerts"`
	SendDiskAlerts            bool `json:"send_disk_alerts"`
	SendVersionMismatchAlerts bool `json:"send_version_mismatch_alerts"`

	// Alert thresholds
	CPUWarning   float64 `json:"cpu_warning"`
	CPUCritical  float64 `json:"cpu_critical"`
	MemWarning   float64 `json:"mem_warning"`
	MemCritical  float64 `json:"mem_critical"`
	DiskWarning  float64 `json:"disk_warning"`
	DiskCritical float64 `json:"disk_critical"`
}

// CreateServerRequest is the request body for CreateServer.
type CreateServerRequest struct {
	Name   string       `json:"name"`
	Config ServerConfig `json:"config"`
}

// UpdateServerRequest is the request body for UpdateServer.
type UpdateServerRequest struct {
	ID     string       `json:"id"`
	Config ServerConfig `json:"config"`
}

// DeleteServerRequest is the request body for DeleteServer.
type DeleteServerRequest struct {
	ID string `json:"id"`
}

// PruneBuildxRequest is the request body for the PruneBuildx execute action.
type PruneBuildxRequest struct {
	Server string `json:"server"`
}

// PruneContainersRequest is the request body for the PruneContainers execute action.
type PruneContainersRequest struct {
	Server string `json:"server"`
}

// PruneDockerBuildersRequest is the request body for the PruneDockerBuilders execute action.
type PruneDockerBuildersRequest struct {
	Server string `json:"server"`
}

// PruneImagesRequest is the request body for the PruneImages execute action.
type PruneImagesRequest struct {
	Server string `json:"server"`
}

// PruneNetworksRequest is the request body for the PruneNetworks execute action.
type PruneNetworksRequest struct {
	Server string `json:"server"`
}

// PruneSystemRequest is the request body for the PruneSystem execute action.
type PruneSystemRequest struct {
	Server string `json:"server"`
}

// PruneVolumesRequest is the request body for the PruneVolumes execute action.
type PruneVolumesRequest struct {
	Server string `json:"server"`
}
