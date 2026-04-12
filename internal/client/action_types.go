// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// Action represents a Komodo action resource (Resource<ActionConfig, NoData>).
type Action struct {
	ID     OID          `json:"_id"`
	Name   string       `json:"name"`
	Config ActionConfig `json:"config"`
}

// ActionConfig is the Komodo action configuration.
type ActionConfig struct {
	RunAtStartup     bool   `json:"run_at_startup"`
	ScheduleFormat   string `json:"schedule_format"`
	Schedule         string `json:"schedule"`
	ScheduleEnabled  bool   `json:"schedule_enabled"`
	ScheduleTimezone string `json:"schedule_timezone"`
	ScheduleAlert    bool   `json:"schedule_alert"`
	FailureAlert     bool   `json:"failure_alert"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	WebhookSecret    string `json:"webhook_secret"`
	ReloadDenoDeps   bool   `json:"reload_deno_deps"`
	FileContents     string `json:"file_contents"`
	ArgumentsFormat  string `json:"arguments_format"`
	Arguments        string `json:"arguments"`
}

// CreateActionRequest is the payload for the CreateAction write API.
type CreateActionRequest struct {
	Name   string              `json:"name"`
	Config PartialActionConfig `json:"config"`
}

// UpdateActionRequest is the payload for the UpdateAction write API.
type UpdateActionRequest struct {
	ID     string              `json:"id"`
	Config PartialActionConfig `json:"config"`
}

// RenameActionRequest is the payload for the RenameAction write API.
type RenameActionRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PartialActionConfig holds optional config fields for Create/Update.
type PartialActionConfig struct {
	RunAtStartup     *bool   `json:"run_at_startup,omitempty"`
	ScheduleFormat   *string `json:"schedule_format,omitempty"`
	Schedule         *string `json:"schedule,omitempty"`
	ScheduleEnabled  *bool   `json:"schedule_enabled,omitempty"`
	ScheduleTimezone *string `json:"schedule_timezone,omitempty"`
	ScheduleAlert    *bool   `json:"schedule_alert,omitempty"`
	FailureAlert     *bool   `json:"failure_alert,omitempty"`
	WebhookEnabled   *bool   `json:"webhook_enabled,omitempty"`
	WebhookSecret    *string `json:"webhook_secret,omitempty"`
	ReloadDenoDeps   *bool   `json:"reload_deno_deps,omitempty"`
	FileContents     *string `json:"file_contents,omitempty"`
	ArgumentsFormat  *string `json:"arguments_format,omitempty"`
	Arguments        *string `json:"arguments,omitempty"`
}

// RunActionRequest is the request body for the RunAction execute action.
type RunActionRequest struct {
	Action string `json:"action"`
}
