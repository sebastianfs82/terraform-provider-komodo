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

	// Maintenance windows
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

// PartialServerConfig mirrors _PartialServerConfig on the API side.
// All fields are pointers with omitempty so only explicitly-set fields are
// sent in JSON — omitted fields are left untouched by the server.
type PartialServerConfig struct {
	// Connection
	Address         *string `json:"address,omitempty"`
	InsecureTLS     *bool   `json:"insecure_tls,omitempty"`
	ExternalAddress *string `json:"external_address,omitempty"`
	Region          *string `json:"region,omitempty"`

	// Behaviour
	Enabled         *bool `json:"enabled,omitempty"`
	AutoRotateKeys  *bool `json:"auto_rotate_keys,omitempty"`
	AutoPrune       *bool `json:"auto_prune,omitempty"`
	StatsMonitoring *bool `json:"stats_monitoring,omitempty"`

	// Mounts / links
	IgnoreMounts *[]string `json:"ignore_mounts,omitempty"`
	Links        *[]string `json:"links,omitempty"`

	// Alert flags
	SendUnreachableAlerts     *bool `json:"send_unreachable_alerts,omitempty"`
	SendCPUAlerts             *bool `json:"send_cpu_alerts,omitempty"`
	SendMemAlerts             *bool `json:"send_mem_alerts,omitempty"`
	SendDiskAlerts            *bool `json:"send_disk_alerts,omitempty"`
	SendVersionMismatchAlerts *bool `json:"send_version_mismatch_alerts,omitempty"`

	// Alert thresholds
	CPUWarning   *float64 `json:"cpu_warning,omitempty"`
	CPUCritical  *float64 `json:"cpu_critical,omitempty"`
	MemWarning   *float64 `json:"mem_warning,omitempty"`
	MemCritical  *float64 `json:"mem_critical,omitempty"`
	DiskWarning  *float64 `json:"disk_warning,omitempty"`
	DiskCritical *float64 `json:"disk_critical,omitempty"`

	// Maintenance windows
	MaintenanceWindows *[]MaintenanceWindow `json:"maintenance_windows,omitempty"`
}

// MaintenanceWindow represents a scheduled maintenance window.
type MaintenanceWindow struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ScheduleType    string `json:"schedule_type"`
	DayOfWeek       string `json:"day_of_week"`
	Date            string `json:"date"`
	Hour            int64  `json:"hour"`
	Minute          int64  `json:"minute"`
	DurationMinutes int64  `json:"duration_minutes"`
	Timezone        string `json:"timezone"`
	Enabled         bool   `json:"enabled"`
}

// BoolPtr is a convenience helper for *bool literals.
func BoolPtr(b bool) *bool { return &b }

// Float64Ptr is a convenience helper for *float64 literals.
func Float64Ptr(f float64) *float64 { return &f }

// StringPtr is a convenience helper for *string literals.
func StringPtr(s string) *string { return &s }

// CreateServerRequest is the request body for CreateServer.
type CreateServerRequest struct {
	Name      string              `json:"name"`
	Config    PartialServerConfig `json:"config"`
	PublicKey *string             `json:"public_key,omitempty"`
}

// UpdateServerPublicKeyRequest is the request body for UpdateServerPublicKey.
type UpdateServerPublicKeyRequest struct {
	Server    string `json:"server"`
	PublicKey string `json:"public_key"`
}

// UpdateServerRequest is the request body for UpdateServer.
type UpdateServerRequest struct {
	ID     string              `json:"id"`
	Config PartialServerConfig `json:"config"`
}

// RenameServerRequest is the payload for the RenameServer write API.
type RenameServerRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
