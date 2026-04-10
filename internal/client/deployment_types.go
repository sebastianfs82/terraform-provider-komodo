// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
)

// DeploymentImageExternal is the "Image" variant of DeploymentImage.
type DeploymentImageExternal struct {
	Image string `json:"image"`
}

// DeploymentImageBuild is the "Build" variant of DeploymentImage.
type DeploymentImageBuild struct {
	BuildID string       `json:"build_id"`
	Version BuildVersion `json:"version"`
}

// DeploymentImage represents the tagged union for a deployment's image source.
// The API uses the format: {"type": "Image", "params": {"image": "nginx:latest"}}.
type DeploymentImage struct {
	Image *DeploymentImageExternal
	Build *DeploymentImageBuild
}

// MarshalJSON encodes DeploymentImage in the Komodo API format:
// {"type": "Image", "params": {...}} or {"type": "Build", "params": {...}}.
func (d DeploymentImage) MarshalJSON() ([]byte, error) {
	if d.Build != nil {
		return json.Marshal(struct {
			Type   string               `json:"type"`
			Params DeploymentImageBuild `json:"params"`
		}{Type: "Build", Params: *d.Build})
	}
	if d.Image != nil {
		return json.Marshal(struct {
			Type   string                  `json:"type"`
			Params DeploymentImageExternal `json:"params"`
		}{Type: "Image", Params: *d.Image})
	}
	// Default to empty Image type
	return json.Marshal(struct {
		Type   string                  `json:"type"`
		Params DeploymentImageExternal `json:"params"`
	}{Type: "Image", Params: DeploymentImageExternal{}})
}

// UnmarshalJSON decodes the Komodo API tagged-union format.
func (d *DeploymentImage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type   string          `json:"type"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch raw.Type {
	case "Build":
		var b DeploymentImageBuild
		if err := json.Unmarshal(raw.Params, &b); err != nil {
			return err
		}
		d.Build = &b
		d.Image = nil
	case "Image", "":
		var e DeploymentImageExternal
		if len(raw.Params) > 0 {
			if err := json.Unmarshal(raw.Params, &e); err != nil {
				return err
			}
		}
		d.Image = &e
		d.Build = nil
	default:
		return fmt.Errorf("unknown DeploymentImage type: %q", raw.Type)
	}
	return nil
}

// Deployment represents a Komodo deployment resource (Resource<DeploymentConfig, DeploymentInfo>).
type Deployment struct {
	ID     OID              `json:"_id"`
	Name   string           `json:"name"`
	Config DeploymentConfig `json:"config"`
}

// DeploymentConfig is the Komodo deployment configuration (full, as returned by the API).
type DeploymentConfig struct {
	SwarmID              string          `json:"swarm_id"`
	ServerID             string          `json:"server_id"`
	Image                DeploymentImage `json:"image"`
	ImageRegistryAccount string          `json:"image_registry_account"`
	SkipSecretInterp     bool            `json:"skip_secret_interp"`
	RedeployOnBuild      bool            `json:"redeploy_on_build"`
	PollForUpdates       bool            `json:"poll_for_updates"`
	AutoUpdate           bool            `json:"auto_update"`
	SendAlerts           bool            `json:"send_alerts"`
	Links                []string        `json:"links"`
	Network              string          `json:"network"`
	Restart              string          `json:"restart"`
	Command              string          `json:"command"`
	Replicas             int             `json:"replicas"`
	TerminationSignal    string          `json:"termination_signal"`
	TerminationTimeout   int             `json:"termination_timeout"`
	ExtraArgs            []string        `json:"extra_args"`
	TermSignalLabels     string          `json:"term_signal_labels"`
	Ports                string          `json:"ports"`
	Volumes              string          `json:"volumes"`
	Environment          string          `json:"environment"`
	Labels               string          `json:"labels"`
}

// PartialDeploymentConfig holds optional config fields for Create/Update.
// Pointer-to-slice types allow sending an explicit empty list without omitempty suppression.
type PartialDeploymentConfig struct {
	SwarmID              *string          `json:"swarm_id,omitempty"`
	ServerID             *string          `json:"server_id,omitempty"`
	Image                *DeploymentImage `json:"image,omitempty"`
	ImageRegistryAccount *string          `json:"image_registry_account,omitempty"`
	SkipSecretInterp     *bool            `json:"skip_secret_interp,omitempty"`
	RedeployOnBuild      *bool            `json:"redeploy_on_build,omitempty"`
	PollForUpdates       *bool            `json:"poll_for_updates,omitempty"`
	AutoUpdate           *bool            `json:"auto_update,omitempty"`
	SendAlerts           *bool            `json:"send_alerts,omitempty"`
	Links                *[]string        `json:"links,omitempty"`
	Network              *string          `json:"network,omitempty"`
	Restart              *string          `json:"restart,omitempty"`
	Command              *string          `json:"command,omitempty"`
	Replicas             *int             `json:"replicas,omitempty"`
	TerminationSignal    *string          `json:"termination_signal,omitempty"`
	TerminationTimeout   *int             `json:"termination_timeout,omitempty"`
	ExtraArgs            *[]string        `json:"extra_args,omitempty"`
	TermSignalLabels     *string          `json:"term_signal_labels,omitempty"`
	Ports                *string          `json:"ports,omitempty"`
	Volumes              *string          `json:"volumes,omitempty"`
	Environment          *string          `json:"environment,omitempty"`
	Labels               *string          `json:"labels,omitempty"`
}

// CreateDeploymentRequest is the payload for the CreateDeployment write API.
type CreateDeploymentRequest struct {
	Name   string                  `json:"name"`
	Config PartialDeploymentConfig `json:"config"`
}

// UpdateDeploymentRequest is the payload for the UpdateDeployment write API.
type UpdateDeploymentRequest struct {
	ID     string                  `json:"id"`
	Config PartialDeploymentConfig `json:"config"`
}

// StartDeploymentRequest is the request body for the StartDeployment execute action.
type StartDeploymentRequest struct {
	Deployment string `json:"deployment"`
}

// PullDeploymentRequest is the request body for the PullDeployment execute action.
type PullDeploymentRequest struct {
	Deployment string `json:"deployment"`
}

// DeployRequest is the request body for the Deploy execute action (full redeploy).
type DeployRequest struct {
	Deployment string `json:"deployment"`
	StopSignal string `json:"stop_signal,omitempty"`
	StopTime   *int64 `json:"stop_time,omitempty"`
}

// StopDeploymentRequest is the request body for the StopDeployment execute action.
type StopDeploymentRequest struct {
	Deployment string `json:"deployment"`
	Signal     string `json:"signal,omitempty"`
	Time       *int64 `json:"time,omitempty"`
}

// DestroyDeploymentRequest is the request body for the DestroyDeployment execute action.
type DestroyDeploymentRequest struct {
	Deployment string `json:"deployment"`
	Signal     string `json:"signal,omitempty"`
	Time       *int64 `json:"time,omitempty"`
}

// RestartDeploymentRequest is the request body for the RestartDeployment execute action.
type RestartDeploymentRequest struct {
	Deployment string `json:"deployment"`
}

// PauseDeploymentRequest is the request body for the PauseDeployment execute action.
type PauseDeploymentRequest struct {
	Deployment string `json:"deployment"`
}

// UnpauseDeploymentRequest is the request body for the UnpauseDeployment execute action.
type UnpauseDeploymentRequest struct {
	Deployment string `json:"deployment"`
}
