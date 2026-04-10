// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// ResourceSync represents a Komodo resource sync (Resource<ResourceSyncConfig>).
type ResourceSync struct {
	ID     OID                `json:"_id"`
	Name   string             `json:"name"`
	Config ResourceSyncConfig `json:"config"`
}

// ResourceSyncConfig is the Komodo resource sync configuration.
type ResourceSyncConfig struct {
	// Git-based source
	LinkedRepo  string `json:"linked_repo"`
	GitProvider string `json:"git_provider"`
	GitHttps    bool   `json:"git_https"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Commit      string `json:"commit"`
	GitAccount  string `json:"git_account"`

	// Files
	FilesOnHost  bool     `json:"files_on_host"`
	ResourcePath []string `json:"resource_path"`
	FileContents string   `json:"file_contents"`

	// Webhook
	WebhookEnabled bool   `json:"webhook_enabled"`
	WebhookSecret  string `json:"webhook_secret"`

	// Sync behaviour
	Managed           bool     `json:"managed"`
	Delete            bool     `json:"delete"`
	IncludeResources  bool     `json:"include_resources"`
	MatchTags         []string `json:"match_tags"`
	IncludeVariables  bool     `json:"include_variables"`
	IncludeUserGroups bool     `json:"include_user_groups"`
	PendingAlert      bool     `json:"pending_alert"`
}

// PartialResourceSyncConfig holds optional fields for Create/Update.
type PartialResourceSyncConfig struct {
	LinkedRepo        *string  `json:"linked_repo,omitempty"`
	GitProvider       *string  `json:"git_provider,omitempty"`
	GitHttps          *bool    `json:"git_https,omitempty"`
	Repo              *string  `json:"repo,omitempty"`
	Branch            *string  `json:"branch,omitempty"`
	Commit            *string  `json:"commit,omitempty"`
	GitAccount        *string  `json:"git_account,omitempty"`
	FilesOnHost       *bool    `json:"files_on_host,omitempty"`
	ResourcePath      []string `json:"resource_path,omitempty"`
	FileContents      *string  `json:"file_contents,omitempty"`
	WebhookEnabled    *bool    `json:"webhook_enabled,omitempty"`
	WebhookSecret     *string  `json:"webhook_secret,omitempty"`
	Managed           *bool    `json:"managed,omitempty"`
	Delete            *bool    `json:"delete,omitempty"`
	IncludeResources  *bool    `json:"include_resources,omitempty"`
	MatchTags         []string `json:"match_tags,omitempty"`
	IncludeVariables  *bool    `json:"include_variables,omitempty"`
	IncludeUserGroups *bool    `json:"include_user_groups,omitempty"`
	PendingAlert      *bool    `json:"pending_alert,omitempty"`
}

// CreateResourceSyncRequest is the payload for CreateResourceSync.
type CreateResourceSyncRequest struct {
	Name   string                    `json:"name"`
	Config PartialResourceSyncConfig `json:"config"`
}

// UpdateResourceSyncRequest is the payload for UpdateResourceSync.
type UpdateResourceSyncRequest struct {
	ID     string                    `json:"id"`
	Config PartialResourceSyncConfig `json:"config"`
}

// RunSyncRequest is the request body for the RunSync execute action.
type RunSyncRequest struct {
	Sync string `json:"sync"`
}
