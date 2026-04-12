// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

import "encoding/json"

// Procedure represents a Komodo procedure resource (Resource<ProcedureConfig>).
type Procedure struct {
	ID     OID             `json:"_id"`
	Name   string          `json:"name"`
	Tags   []string        `json:"tags"`
	Config ProcedureConfig `json:"config"`
}

// ProcedureConfig is the Komodo procedure configuration (full, as returned by the API).
type ProcedureConfig struct {
	// Stages is stored as raw JSON (array of ProcedureStage) to avoid enumerating
	// all Execution enum variants. Terraform represents this as a JSON string.
	Stages           json.RawMessage `json:"stages"`
	ScheduleFormat   string          `json:"schedule_format"`
	Schedule         string          `json:"schedule"`
	ScheduleEnabled  bool            `json:"schedule_enabled"`
	ScheduleTimezone string          `json:"schedule_timezone"`
	ScheduleAlert    bool            `json:"schedule_alert"`
	FailureAlert     bool            `json:"failure_alert"`
	WebhookEnabled   bool            `json:"webhook_enabled"`
	WebhookSecret    string          `json:"webhook_secret"`
}

// PartialProcedureConfig holds optional config fields for Create/Update.
// Stages is json.RawMessage so that nil is omitted and a set value is passed through.
type PartialProcedureConfig struct {
	Stages           json.RawMessage `json:"stages,omitempty"`
	ScheduleFormat   *string         `json:"schedule_format,omitempty"`
	Schedule         *string         `json:"schedule,omitempty"`
	ScheduleEnabled  *bool           `json:"schedule_enabled,omitempty"`
	ScheduleTimezone *string         `json:"schedule_timezone,omitempty"`
	ScheduleAlert    *bool           `json:"schedule_alert,omitempty"`
	FailureAlert     *bool           `json:"failure_alert,omitempty"`
	WebhookEnabled   *bool           `json:"webhook_enabled,omitempty"`
	WebhookSecret    *string         `json:"webhook_secret,omitempty"`
}

// CreateProcedureRequest is the payload for the CreateProcedure write API.
type CreateProcedureRequest struct {
	Name   string                 `json:"name"`
	Config PartialProcedureConfig `json:"config"`
}

// UpdateProcedureRequest is the payload for the UpdateProcedure write API.
type UpdateProcedureRequest struct {
	ID     string                 `json:"id"`
	Config PartialProcedureConfig `json:"config"`
}

// RenameProcedureRequest is the payload for the RenameProcedure write API.
type RenameProcedureRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RunProcedureRequest is the request body for the RunProcedure execute action.
type RunProcedureRequest struct {
	Procedure string `json:"procedure"`
}
