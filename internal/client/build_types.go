// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

// BuildVersion represents a semantic version for a Build.
type BuildVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// ImageRegistryConfig holds the registry configuration for a Build image.
type ImageRegistryConfig struct {
	Domain       string `json:"domain"`
	Account      string `json:"account"`
	Organization string `json:"organization"`
}

// Build represents a Komodo build resource (Resource<BuildConfig, BuildInfo>).
type Build struct {
	ID     OID         `json:"_id"`
	Name   string      `json:"name"`
	Config BuildConfig `json:"config"`
}

// BuildConfig is the Komodo build configuration.
type BuildConfig struct {
	BuilderID            string                `json:"builder_id"`
	Version              BuildVersion          `json:"version"`
	AutoIncrementVersion bool                  `json:"auto_increment_version"`
	ImageName            string                `json:"image_name"`
	ImageTag             string                `json:"image_tag"`
	IncludeLatestTag     bool                  `json:"include_latest_tag"`
	IncludeVersionTags   bool                  `json:"include_version_tags"`
	IncludeCommitTag     bool                  `json:"include_commit_tag"`
	Links                []string              `json:"links"`
	LinkedRepo           string                `json:"linked_repo"`
	GitProvider          string                `json:"git_provider"`
	GitHttps             bool                  `json:"git_https"`
	GitAccount           string                `json:"git_account"`
	Repo                 string                `json:"repo"`
	Branch               string                `json:"branch"`
	Commit               string                `json:"commit"`
	WebhookEnabled       bool                  `json:"webhook_enabled"`
	WebhookSecret        string                `json:"webhook_secret"`
	FilesOnHost          bool                  `json:"files_on_host"`
	BuildPath            string                `json:"build_path"`
	DockerfilePath       string                `json:"dockerfile_path"`
	ImageRegistry        []ImageRegistryConfig `json:"image_registry"`
	SkipSecretInterp     bool                  `json:"skip_secret_interp"`
	UseBuildx            bool                  `json:"use_buildx"`
	ExtraArgs            []string              `json:"extra_args"`
	PreBuild             SystemCommand         `json:"pre_build"`
	Dockerfile           string                `json:"dockerfile"`
	BuildArgs            string                `json:"build_args"`
	SecretArgs           string                `json:"secret_args"`
	Labels               string                `json:"labels"`
}

// PartialBuildConfig holds optional config fields for Create/Update.
// Pointer-to-slice types (Links, ExtraArgs, ImageRegistry) allow sending an
// explicit empty list to the API (clearing entries) without being omitted.
type PartialBuildConfig struct {
	BuilderID            *string                `json:"builder_id,omitempty"`
	Version              *BuildVersion          `json:"version,omitempty"`
	AutoIncrementVersion *bool                  `json:"auto_increment_version,omitempty"`
	ImageName            *string                `json:"image_name,omitempty"`
	ImageTag             *string                `json:"image_tag,omitempty"`
	IncludeLatestTag     *bool                  `json:"include_latest_tag,omitempty"`
	IncludeVersionTags   *bool                  `json:"include_version_tags,omitempty"`
	IncludeCommitTag     *bool                  `json:"include_commit_tag,omitempty"`
	Links                *[]string              `json:"links,omitempty"`
	LinkedRepo           *string                `json:"linked_repo,omitempty"`
	GitProvider          *string                `json:"git_provider,omitempty"`
	GitHttps             *bool                  `json:"git_https,omitempty"`
	GitAccount           *string                `json:"git_account,omitempty"`
	Repo                 *string                `json:"repo,omitempty"`
	Branch               *string                `json:"branch,omitempty"`
	Commit               *string                `json:"commit,omitempty"`
	WebhookEnabled       *bool                  `json:"webhook_enabled,omitempty"`
	WebhookSecret        *string                `json:"webhook_secret,omitempty"`
	FilesOnHost          *bool                  `json:"files_on_host,omitempty"`
	BuildPath            *string                `json:"build_path,omitempty"`
	DockerfilePath       *string                `json:"dockerfile_path,omitempty"`
	ImageRegistry        *[]ImageRegistryConfig `json:"image_registry,omitempty"`
	SkipSecretInterp     *bool                  `json:"skip_secret_interp,omitempty"`
	UseBuildx            *bool                  `json:"use_buildx,omitempty"`
	ExtraArgs            *[]string              `json:"extra_args,omitempty"`
	PreBuild             *SystemCommand         `json:"pre_build,omitempty"`
	Dockerfile           *string                `json:"dockerfile,omitempty"`
	BuildArgs            *string                `json:"build_args,omitempty"`
	SecretArgs           *string                `json:"secret_args,omitempty"`
	Labels               *string                `json:"labels,omitempty"`
}

// CreateBuildRequest is the payload for the CreateBuild write API.
type CreateBuildRequest struct {
	Name   string             `json:"name"`
	Config PartialBuildConfig `json:"config"`
}

// UpdateBuildRequest is the payload for the UpdateBuild write API.
type UpdateBuildRequest struct {
	ID     string             `json:"id"`
	Config PartialBuildConfig `json:"config"`
}

// RunBuildRequest is the request body for the RunBuild execute action.
type RunBuildRequest struct {
	Build string `json:"build"`
}
