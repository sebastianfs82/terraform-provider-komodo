// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// TerminalTargetParams holds the target-specific identifiers for a terminal.
// Which fields are populated depends on TerminalTarget.Type:
//   - Server:     Server only
//   - Container:  Server (host) + Container (name)
//   - Stack:      Stack + Service
//   - Deployment: Deployment only
type TerminalTargetParams struct {
	Server     *string `json:"server,omitempty"`
	Container  *string `json:"container,omitempty"`
	Stack      *string `json:"stack,omitempty"`
	Deployment *string `json:"deployment,omitempty"`
	Service    *string `json:"service,omitempty"`
}

// TerminalTarget is the target for a terminal session.
// Supported types: Server, Container, Stack, Deployment.
type TerminalTarget struct {
	Type   string               `json:"type"`
	Params TerminalTargetParams `json:"params"`
}

// NewTerminalTarget creates a TerminalTarget from its constituent fields.
//   - targetType: "Server", "Container", "Stack", or "Deployment"
//   - targetID:   server ID (Server/Container), stack ID (Stack), deployment ID (Deployment)
//   - container:  container name (Container only)
//   - service:    service name (Stack only)
func NewTerminalTarget(targetType, targetID, container, service string) TerminalTarget {
	params := TerminalTargetParams{}
	switch targetType {
	case "Container":
		params.Server = &targetID
		if container != "" {
			params.Container = &container
		}
	case "Stack":
		params.Stack = &targetID
		if service != "" {
			params.Service = &service
		}
	case "Deployment":
		params.Deployment = &targetID
	default: // "Server"
		params.Server = &targetID
	}
	return TerminalTarget{Type: targetType, Params: params}
}

// TargetID returns the primary resource ID from the target params.
// For Container this is the host server; for Stack the stack; etc.
func (t TerminalTarget) TargetID() string {
	switch t.Type {
	case "Container":
		if t.Params.Server != nil {
			return *t.Params.Server
		}
	case "Stack":
		if t.Params.Stack != nil {
			return *t.Params.Stack
		}
	case "Deployment":
		if t.Params.Deployment != nil {
			return *t.Params.Deployment
		}
	default: // "Server"
		if t.Params.Server != nil {
			return *t.Params.Server
		}
	}
	return ""
}

// ContainerName returns the container name, or empty string when not set.
func (t TerminalTarget) ContainerName() string {
	if t.Params.Container != nil {
		return *t.Params.Container
	}
	return ""
}

// ServiceName returns the stack service name, or empty string when not set.
func (t TerminalTarget) ServiceName() string {
	if t.Params.Service != nil {
		return *t.Params.Service
	}
	return ""
}

// Terminal is the API response for a terminal resource.
type Terminal struct {
	Name         string         `json:"name"`
	Command      string         `json:"command"`
	CreatedAt    int64          `json:"created_at"`
	StoredSizeKB float64        `json:"stored_size_kb"`
	Target       TerminalTarget `json:"target"`
}

// CreateTerminalRequest is the request payload for CreateTerminal.
type CreateTerminalRequest struct {
	Target   TerminalTarget `json:"target"`
	Command  *string        `json:"command,omitempty"`
	Mode     *string        `json:"mode,omitempty"`
	Name     *string        `json:"name,omitempty"`
	Recreate string         `json:"recreate"`
}

// DeleteTerminalRequest is the request payload for DeleteTerminal.
type DeleteTerminalRequest struct {
	Target   TerminalTarget `json:"target"`
	Terminal string         `json:"terminal"`
}

// ListTerminalsRequest is the request payload for ListTerminals.
type ListTerminalsRequest struct {
	Target   *TerminalTarget `json:"target,omitempty"`
	UseNames bool            `json:"use_names"`
}
