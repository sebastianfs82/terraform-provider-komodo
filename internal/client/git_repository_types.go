// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// GitRepository represents a Komodo git repository resource.
type GitRepository struct {
	ID     OID                 `json:"_id"`
	Name   string              `json:"name"`
	Tags   []string            `json:"tags"`
	Config GitRepositoryConfig `json:"config"`
}

// SystemCommand represents a shell command with an optional working directory.
type SystemCommand struct {
	Path    string `json:"path"`
	Command string `json:"command"`
}

// GitRepositoryConfig is the configuration for a Komodo git repository.
type GitRepositoryConfig struct {
	ServerID         string        `json:"server_id"`
	BuilderID        string        `json:"builder_id"`
	GitProvider      string        `json:"git_provider"`
	GitHttps         bool          `json:"git_https"`
	GitAccount       string        `json:"git_account"`
	Repo             string        `json:"repo"`
	Branch           string        `json:"branch"`
	Commit           string        `json:"commit"`
	Path             string        `json:"path"`
	WebhookEnabled   bool          `json:"webhook_enabled"`
	WebhookSecret    string        `json:"webhook_secret"`
	OnClone          SystemCommand `json:"on_clone"`
	OnPull           SystemCommand `json:"on_pull"`
	Links            []string      `json:"links"`
	Environment      string        `json:"environment"`
	EnvFilePath      string        `json:"env_file_path"`
	SkipSecretInterp bool          `json:"skip_secret_interp"`
}

// CreateGitRepositoryRequest is the request body for CreateRepo.
type CreateGitRepositoryRequest struct {
	Name   string              `json:"name"`
	Config GitRepositoryConfig `json:"config"`
}

// UpdateGitRepositoryRequest is the request body for UpdateRepo.
type UpdateGitRepositoryRequest struct {
	ID     string              `json:"id"`
	Config GitRepositoryConfig `json:"config"`
}

// RenameGitRepositoryRequest is the payload for the RenameRepo write API.
type RenameGitRepositoryRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DeleteGitRepositoryRequest is the request body for DeleteRepo.
type DeleteGitRepositoryRequest struct {
	ID string `json:"id"`
}

// BuildRepoRequest is the request body for the BuildRepo execute action.
type BuildRepoRequest struct {
	Repo string `json:"repo"`
}

// CloneRepoRequest is the request body for the CloneRepo execute action.
type CloneRepoRequest struct {
	Repo string `json:"repo"`
}

// PullRepoRequest is the request body for the PullRepo execute action.
type PullRepoRequest struct {
	Repo string `json:"repo"`
}
