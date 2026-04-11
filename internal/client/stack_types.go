// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// Stack represents a Komodo stack resource.
type Stack struct {
	ID     OID         `json:"_id"`
	Name   string      `json:"name"`
	Config StackConfig `json:"config"`
}

// StackConfig is the configuration for a Komodo stack.
type StackConfig struct {
	ServerID                 string        `json:"server_id"`
	SwarmID                  string        `json:"swarm_id"`
	Links                    []string      `json:"links"`
	ProjectName              string        `json:"project_name"`
	AutoPull                 bool          `json:"auto_pull"`
	RunBuild                 bool          `json:"run_build"`
	PollForUpdates           bool          `json:"poll_for_updates"`
	AutoUpdate               bool          `json:"auto_update"`
	AutoUpdateAllServices    bool          `json:"auto_update_all_services"`
	DestroyBeforeDeploy      bool          `json:"destroy_before_deploy"`
	LinkedRepo               string        `json:"linked_repo"`
	GitProvider              string        `json:"git_provider"`
	GitHttps                 bool          `json:"git_https"`
	GitAccount               string        `json:"git_account"`
	Repo                     string        `json:"repo"`
	Branch                   string        `json:"branch"`
	Commit                   string        `json:"commit"`
	Reclone                  bool          `json:"reclone"`
	WebhookEnabled           bool          `json:"webhook_enabled"`
	WebhookSecret            string        `json:"webhook_secret"`
	WebhookForceDeploy       bool          `json:"webhook_force_deploy"`
	FilesOnHost              bool          `json:"files_on_host"`
	RunDirectory             string        `json:"run_directory"`
	FilePaths                []string      `json:"file_paths"`
	EnvFilePath              string        `json:"env_file_path"`
	Environment              string        `json:"environment"`
	FileContents             string        `json:"file_contents"`
	SendAlerts               bool          `json:"send_alerts"`
	RegistryProvider         string        `json:"registry_provider"`
	RegistryAccount          string        `json:"registry_account"`
	PreDeploy                SystemCommand `json:"pre_deploy"`
	PostDeploy               SystemCommand `json:"post_deploy"`
	ExtraArgs                []string      `json:"extra_args"`
	BuildExtraArgs           []string      `json:"build_extra_args"`
	ComposeCmdWrapper        string        `json:"compose_cmd_wrapper"`
	ComposeCmdWrapperInclude []string      `json:"compose_cmd_wrapper_include"`
	IgnoreServices           []string      `json:"ignore_services"`
}

// CreateStackRequest is the request body for CreateStack.
type CreateStackRequest struct {
	Name   string      `json:"name"`
	Config StackConfig `json:"config"`
}

// UpdateStackRequest is the request body for UpdateStack.
type UpdateStackRequest struct {
	ID     string      `json:"id"`
	Config StackConfig `json:"config"`
}

// DeleteStackRequest is the request body for DeleteStack.
type DeleteStackRequest struct {
	ID string `json:"id"`
}

// Execute action request types.

// StartStackRequest is the request body for the StartStack execute action.
type StartStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
}

// StopStackRequest is the request body for the StopStack execute action.
type StopStackRequest struct {
	Stack    string   `json:"stack"`
	StopTime *int64   `json:"stop_time,omitempty"`
	Services []string `json:"services"`
}

// PauseStackRequest is the request body for the PauseStack execute action.
type PauseStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
}

// DeployStackRequest is the request body for the DeployStack execute action.
type DeployStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
	StopTime *int64   `json:"stop_time,omitempty"`
}

// DestroyStackActionRequest is the request body for the DestroyStack execute action.
type DestroyStackActionRequest struct {
	Stack         string   `json:"stack"`
	Services      []string `json:"services"`
	RemoveOrphans bool     `json:"remove_orphans"`
	StopTime      *int64   `json:"stop_time,omitempty"`
}

// RestartStackRequest is the request body for the RestartStack execute action.
type RestartStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
}

// UnpauseStackRequest is the request body for the UnpauseStack execute action.
type UnpauseStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
}

// PullStackRequest is the request body for the PullStack execute action.
type PullStackRequest struct {
	Stack    string   `json:"stack"`
	Services []string `json:"services"`
}

// DeployStackIfChangedRequest is the request body for the DeployStackIfChanged execute action.
type DeployStackIfChangedRequest struct {
	Stack    string `json:"stack"`
	StopTime *int64 `json:"stop_time,omitempty"`
}

// RunStackServiceRequest is the request body for the RunStackService execute action.
type RunStackServiceRequest struct {
	Stack        string            `json:"stack"`
	Service      string            `json:"service"`
	Command      []string          `json:"command,omitempty"`
	NoTty        *bool             `json:"no_tty,omitempty"`
	NoDeps       *bool             `json:"no_deps,omitempty"`
	Detach       *bool             `json:"detach,omitempty"`
	ServicePorts *bool             `json:"service_ports,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Workdir      *string           `json:"workdir,omitempty"`
	User         *string           `json:"user,omitempty"`
	Entrypoint   *string           `json:"entrypoint,omitempty"`
	Pull         *bool             `json:"pull,omitempty"`
}
