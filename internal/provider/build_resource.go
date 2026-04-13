// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &BuildResource{}
var _ resource.ResourceWithImportState = &BuildResource{}
var _ resource.ResourceWithConfigValidators = &BuildResource{}

func NewBuildResource() resource.Resource {
	return &BuildResource{}
}

type BuildResource struct {
	client *client.Client
}

// BuildVersionModel is the Terraform model for the version block of a build.
type BuildVersionModel struct {
	Value                types.String `tfsdk:"value"`
	AutoIncrementEnabled types.Bool   `tfsdk:"auto_increment_enabled"`
}

// BuildImageModel is the Terraform model for the image block of a build.
type BuildImageModel struct {
	Name               types.String               `tfsdk:"name"`
	Tag                types.String               `tfsdk:"tag"`
	IncludeLatestTag   types.Bool                 `tfsdk:"include_latest_tag_enabled"`
	IncludeVersionTags types.Bool                 `tfsdk:"include_version_tags_enabled"`
	IncludeCommitTag   types.Bool                 `tfsdk:"include_commit_tag_enabled"`
	Registries         []ImageRegistryConfigModel `tfsdk:"registry"`
}

// DockerBuildModel is the Terraform model for the build block of a build.
type DockerBuildModel struct {
	Path       types.String `tfsdk:"path"`
	ExtraArgs  types.List   `tfsdk:"extra_args"`
	Args       types.String `tfsdk:"args"`
	SecretArgs types.String `tfsdk:"secret_args"`
}

// BuildSourceModel is the Terraform model for the source block of a build.
type BuildSourceModel struct {
	RepoID    types.String `tfsdk:"repo_id"`
	URL       types.String `tfsdk:"url"`
	AccountID types.String `tfsdk:"account_id"`
	Path      types.String `tfsdk:"path"`
	Branch    types.String `tfsdk:"branch"`
	Commit    types.String `tfsdk:"commit"`
}

// ImageRegistryConfigModel is the Terraform model for an image_registry list entry.
type ImageRegistryConfigModel struct {
	Account      types.String `tfsdk:"account_id"`
	Organization types.String `tfsdk:"organization"`
}

// BuildResourceModel is the Terraform resource model for komodo_build.
type BuildResourceModel struct {
	ID               types.String        `tfsdk:"id"`
	Name             types.String        `tfsdk:"name"`
	Tags             types.List          `tfsdk:"tags"`
	BuilderID        types.String        `tfsdk:"builder_id"`
	Version          *BuildVersionModel  `tfsdk:"version"`
	Image            *BuildImageModel    `tfsdk:"image"`
	Links            types.List          `tfsdk:"links"`
	Source           *BuildSourceModel   `tfsdk:"source"`
	Webhook          *WebhookModel       `tfsdk:"webhook"`
	FilesOnHost      types.Bool          `tfsdk:"files_on_host"`
	Build            *DockerBuildModel   `tfsdk:"build"`
	DockerfilePath   types.String        `tfsdk:"dockerfile_path"`
	SkipSecretInterp types.Bool          `tfsdk:"skip_secret_interp"`
	UseBuildx        types.Bool          `tfsdk:"use_buildx"`
	PreBuild         *SystemCommandModel `tfsdk:"pre_build"`
	Dockerfile       types.String        `tfsdk:"dockerfile"`
	Labels           types.String        `tfsdk:"labels"`
}

func (r *BuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (r *BuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The working directory for the command.",
		},
		"command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The shell command to run.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo build resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The build identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the build.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of tag IDs to attach to this resource. Use `komodo_tag.<name>.id` to reference tags.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"builder_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the builder to use. Leave empty to use the default builder.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Quick links associated with this build.",
			},
			"files_on_host": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use files on the host filesystem for the build context instead of a git repository.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dockerfile_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Path to the Dockerfile relative to `build_path`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_secret_interp": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to skip secret interpolation in build args.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"use_buildx": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use `docker buildx` for multi-platform builds.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dockerfile": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Inline Dockerfile contents. Overrides `dockerfile_path` when set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Docker image labels in `KEY=VALUE` format, newline-separated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"version": schema.SingleNestedBlock{
				MarkdownDescription: "Semantic version and auto-increment settings for the built image.",
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Semantic version for the built image, e.g. `1.0.0`.",
					},
					"auto_increment_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether to automatically increment the patch version on each build. Defaults to true.",
					},
				},
			},
			"image": schema.SingleNestedBlock{
				MarkdownDescription: "Image configuration for the build output.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Override for the image name. Defaults to the build name.",
					},
					"tag": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "An extra tag suffix to apply to the image.",
					},
					"include_latest_tag_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether to push a `:latest` tag alongside the versioned tag.",
					},
					"include_version_tags_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether to push individual semver component tags (e.g. `:1`, `:1.2`).",
					},
					"include_commit_tag_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether to push a tag with the git commit hash.",
					},
				},
				Blocks: map[string]schema.Block{
					"registry": schema.ListNestedBlock{
						MarkdownDescription: "Image registry configurations to push the built image to.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"account_id": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "The ID of a `komodo_registry_account` resource to push the image to.",
								},
								"organization": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Optional organization name within the registry account.",
								},
							},
						},
					},
				},
			},
			"source": schema.SingleNestedBlock{
				MarkdownDescription: "Git source configuration for repo-based builds.",
				Attributes: map[string]schema.Attribute{
					"repo_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Id or name of a linked `komodo_repo` resource.",
					},
					"url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The URL of the git provider, e.g. `https://github.com`.",
					},
					"account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Git account for private repositories.",
					},
					"path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The repository path, e.g. `owner/repo`.",
					},
					"branch": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The branch to check out.",
					},
					"commit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "A specific commit hash to check out.",
					},
				},
			},
			"webhook": schema.SingleNestedBlock{
				MarkdownDescription: "Webhook configuration for the build.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to allow triggering this build via webhook.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Override the default webhook secret for this build.",
					},
				},
			},
			"build": schema.SingleNestedBlock{
				MarkdownDescription: "Docker build configuration.",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("."),
						MarkdownDescription: "Path to the Docker build context directory. Defaults to `.`.",
					},
					"extra_args": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Additional arguments to pass to the `docker build` command.",
					},
					"args": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Docker build arguments in `KEY=VALUE` format, newline-separated.",
					},
					"secret_args": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Docker secret build arguments. These are not stored in the image layers.",
					},
				},
			},
			"pre_build": schema.SingleNestedBlock{
				MarkdownDescription: "A command to run before the Docker build.",
				Attributes:          systemCommandAttrs,
			},
		},
	}
}

func (r *BuildResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *BuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating build", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := client.CreateBuildRequest{
		Name:   data.Name.ValueString(),
		Config: partialBuildConfigFromModel(ctx, r.client, &data),
	}
	b, err := r.client.CreateBuild(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create build, got error: %s", err))
		return
	}
	if b.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Build creation failed: missing ID",
			"The Komodo API did not return a build ID. Resource cannot be tracked in state.",
		)
		return
	}
	plannedTags := data.Tags
	buildToModel(ctx, r.client, b, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Build", ID: b.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on build, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	tflog.Trace(ctx, "Created build resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	b, err := r.client.GetBuild(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read build, got error: %s", err))
		return
	}
	if b == nil {
		// Resource may have been externally recreated with a new ID — try name lookup before removing from state.
		b, err = r.client.GetBuild(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read build by name, got error: %s", err))
			return
		}
		if b == nil {
			tflog.Debug(ctx, "Build not found by ID or name, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Debug(ctx, "Build adopted by name after ID lookup failed", map[string]interface{}{"name": b.Name, "new_id": b.ID.OID})
	}
	buildToModel(ctx, r.client, b, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuildResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state BuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameBuild(ctx, client.RenameBuildRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename build, got error: %s", err))
			return
		}
	}

	updateReq := client.UpdateBuildRequest{
		ID:     data.ID.ValueString(),
		Config: partialBuildConfigFromModel(ctx, r.client, &data),
	}
	b, err := r.client.UpdateBuild(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update build, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	buildToModel(ctx, r.client, b, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Build", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on build, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuildResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting build", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteBuild(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete build, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted build resource")
}

func (r *BuildResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *BuildResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		buildGitRepoConflictsValidator{},
	}
}

type buildGitRepoConflictsValidator struct{}

func (v buildGitRepoConflictsValidator) Description(_ context.Context) string {
	return "`source.repo_id` cannot be set together with `source.url`, `source.account_id`, `source.path`, `source.branch`, or `source.commit`"
}

func (v buildGitRepoConflictsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v buildGitRepoConflictsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data BuildResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Source == nil || data.Source.RepoID.IsNull() || data.Source.RepoID.IsUnknown() {
		return
	}
	conflicts := map[string]bool{
		"url":        !data.Source.URL.IsNull() && !data.Source.URL.IsUnknown(),
		"account_id": !data.Source.AccountID.IsNull() && !data.Source.AccountID.IsUnknown(),
		"path":       !data.Source.Path.IsNull() && !data.Source.Path.IsUnknown(),
		"branch":     !data.Source.Branch.IsNull() && !data.Source.Branch.IsUnknown(),
		"commit":     !data.Source.Commit.IsNull() && !data.Source.Commit.IsUnknown(),
	}
	for field, set := range conflicts {
		if set {
			resp.Diagnostics.AddAttributeError(
				path.Root("source").AtName("repo_id"),
				"source.repo_id conflicts with other source fields",
				fmt.Sprintf("`source.repo_id` cannot be set together with `source.%s`. Use either a linked `komodo_repo` (`repo_id`) or direct source fields, not both.", field),
			)
		}
	}
}

func partialBuildConfigFromModel(ctx context.Context, c *client.Client, data *BuildResourceModel) client.PartialBuildConfig {
	cfg := client.PartialBuildConfig{}

	if !data.BuilderID.IsNull() && !data.BuilderID.IsUnknown() {
		v := data.BuilderID.ValueString()
		cfg.BuilderID = &v
	}
	if data.Version != nil {
		if v := data.Version.Value.ValueString(); v != "" {
			parts := strings.SplitN(v, ".", 3)
			for len(parts) < 3 {
				parts = append(parts, "0")
			}
			major, _ := strconv.Atoi(parts[0])
			minor, _ := strconv.Atoi(parts[1])
			patch, _ := strconv.Atoi(parts[2])
			cfg.Version = &client.BuildVersion{Major: major, Minor: minor, Patch: patch}
		}
		if !data.Version.AutoIncrementEnabled.IsNull() && !data.Version.AutoIncrementEnabled.IsUnknown() {
			v := data.Version.AutoIncrementEnabled.ValueBool()
			cfg.AutoIncrementVersion = &v
		}
	} else {
		// Explicitly reset version to 0.0.0 and enable auto-increment when block is removed.
		cfg.Version = &client.BuildVersion{Major: 0, Minor: 0, Patch: 0}
		t := true
		cfg.AutoIncrementVersion = &t
	}
	if data.Image != nil {
		if !data.Image.Name.IsNull() && !data.Image.Name.IsUnknown() {
			v := data.Image.Name.ValueString()
			cfg.ImageName = &v
		}
		if !data.Image.Tag.IsNull() && !data.Image.Tag.IsUnknown() {
			v := data.Image.Tag.ValueString()
			cfg.ImageTag = &v
		}
		if !data.Image.IncludeLatestTag.IsNull() && !data.Image.IncludeLatestTag.IsUnknown() {
			v := data.Image.IncludeLatestTag.ValueBool()
			cfg.IncludeLatestTag = &v
		}
		if !data.Image.IncludeVersionTags.IsNull() && !data.Image.IncludeVersionTags.IsUnknown() {
			v := data.Image.IncludeVersionTags.ValueBool()
			cfg.IncludeVersionTags = &v
		}
		if !data.Image.IncludeCommitTag.IsNull() && !data.Image.IncludeCommitTag.IsUnknown() {
			v := data.Image.IncludeCommitTag.ValueBool()
			cfg.IncludeCommitTag = &v
		}
		if data.Image.Registries != nil {
			regs := make([]client.ImageRegistryConfig, len(data.Image.Registries))
			for i, r := range data.Image.Registries {
				var domain, username string
				if acc, err := c.GetDockerRegistryAccount(ctx, r.Account.ValueString()); err == nil && acc != nil {
					domain = acc.Domain
					username = acc.Username
				}
				regs[i] = client.ImageRegistryConfig{
					Domain:       domain,
					Account:      username,
					Organization: r.Organization.ValueString(),
				}
			}
			cfg.ImageRegistry = &regs
		}
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		data.Links.ElementsAs(ctx, &links, false)
		if links == nil {
			links = []string{}
		}
		cfg.Links = &links
	}
	if data.Source != nil {
		repoID := data.Source.RepoID.ValueString()
		cfg.LinkedRepo = &repoID
		url := data.Source.URL.ValueString()
		var provider string
		https := true
		if strings.HasPrefix(url, "https://") {
			provider = strings.TrimPrefix(url, "https://")
		} else if strings.HasPrefix(url, "http://") {
			provider = strings.TrimPrefix(url, "http://")
			https = false
		} else if url != "" {
			provider = url
		} else if acc := c.ResolveGitAccountFull(ctx, data.Source.AccountID.ValueString()); acc != nil {
			// No URL set: derive the provider domain from the account's registered domain
			// so the API stores the correct domain instead of defaulting to "github.com".
			provider = acc.Domain
		}
		if provider != "" {
			cfg.GitProvider = &provider
			cfg.GitHttps = &https
		}
		account, err := c.ResolveGitAccountUsername(ctx, data.Source.AccountID.ValueString())
		if err != nil {
			account = data.Source.AccountID.ValueString()
		}
		cfg.GitAccount = &account
		p := data.Source.Path.ValueString()
		cfg.Repo = &p
		b := data.Source.Branch.ValueString()
		cfg.Branch = &b
		co := data.Source.Commit.ValueString()
		cfg.Commit = &co
	}
	if data.Webhook != nil {
		if !data.Webhook.Enabled.IsNull() && !data.Webhook.Enabled.IsUnknown() {
			v := data.Webhook.Enabled.ValueBool()
			cfg.WebhookEnabled = &v
		}
		if !data.Webhook.Secret.IsNull() && !data.Webhook.Secret.IsUnknown() {
			v := data.Webhook.Secret.ValueString()
			cfg.WebhookSecret = &v
		}
	} else {
		f, s := false, ""
		cfg.WebhookEnabled = &f
		cfg.WebhookSecret = &s
	}
	if !data.FilesOnHost.IsNull() && !data.FilesOnHost.IsUnknown() {
		v := data.FilesOnHost.ValueBool()
		cfg.FilesOnHost = &v
	}
	if data.Build != nil {
		if !data.Build.Path.IsNull() && !data.Build.Path.IsUnknown() {
			v := data.Build.Path.ValueString()
			cfg.BuildPath = &v
		}
		if !data.Build.ExtraArgs.IsNull() && !data.Build.ExtraArgs.IsUnknown() {
			var args []string
			data.Build.ExtraArgs.ElementsAs(ctx, &args, false)
			if args == nil {
				args = []string{}
			}
			cfg.ExtraArgs = &args
		}
		if !data.Build.Args.IsNull() && !data.Build.Args.IsUnknown() {
			v := data.Build.Args.ValueString()
			cfg.BuildArgs = &v
		}
		if !data.Build.SecretArgs.IsNull() && !data.Build.SecretArgs.IsUnknown() {
			v := data.Build.SecretArgs.ValueString()
			cfg.SecretArgs = &v
		}
	}
	if !data.DockerfilePath.IsNull() && !data.DockerfilePath.IsUnknown() {
		v := data.DockerfilePath.ValueString()
		cfg.DockerfilePath = &v
	}
	if !data.SkipSecretInterp.IsNull() && !data.SkipSecretInterp.IsUnknown() {
		v := data.SkipSecretInterp.ValueBool()
		cfg.SkipSecretInterp = &v
	}
	if !data.UseBuildx.IsNull() && !data.UseBuildx.IsUnknown() {
		v := data.UseBuildx.ValueBool()
		cfg.UseBuildx = &v
	}
	if data.PreBuild != nil {
		cfg.PreBuild = &client.SystemCommand{
			Path:    data.PreBuild.Path.ValueString(),
			Command: data.PreBuild.Command.ValueString(),
		}
	} else {
		// Explicitly clear pre_build on the API when the user removes the block.
		empty := ""
		cfg.PreBuild = &client.SystemCommand{Path: empty, Command: empty}
	}
	if !data.Dockerfile.IsNull() && !data.Dockerfile.IsUnknown() {
		v := data.Dockerfile.ValueString()
		cfg.Dockerfile = &v
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		v := data.Labels.ValueString()
		cfg.Labels = &v
	}

	return cfg
}

// buildToModel populates a BuildResourceModel from a Build API response.
func buildToModel(ctx context.Context, c *client.Client, b *client.Build, data *BuildResourceModel) {
	data.ID = types.StringValue(b.ID.OID)
	data.Name = types.StringValue(b.Name)
	tagsSlice := b.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	data.Tags = tags
	data.BuilderID = types.StringValue(b.Config.BuilderID)

	// Populate version block only when it was previously set.
	if data.Version != nil {
		v := b.Config.Version
		valueStr := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
		data.Version = &BuildVersionModel{
			Value:                types.StringValue(valueStr),
			AutoIncrementEnabled: types.BoolValue(b.Config.AutoIncrementVersion),
		}
	}
	// image block: only populate when the block was already set.
	if data.Image != nil {
		var regs []ImageRegistryConfigModel
		if len(b.Config.ImageRegistry) > 0 {
			regs = make([]ImageRegistryConfigModel, len(b.Config.ImageRegistry))
			for i, r := range b.Config.ImageRegistry {
				accountID := c.ResolveDockerRegistryAccountID(ctx, r.Domain, r.Account)
				regs[i] = ImageRegistryConfigModel{
					Account:      types.StringValue(accountID),
					Organization: types.StringValue(r.Organization),
				}
			}
		} else if data.Image != nil && data.Image.Registries != nil && len(data.Image.Registries) == 0 {
			regs = []ImageRegistryConfigModel{}
		}
		data.Image = &BuildImageModel{
			Name:               types.StringValue(b.Config.ImageName),
			Tag:                types.StringValue(b.Config.ImageTag),
			IncludeLatestTag:   types.BoolValue(b.Config.IncludeLatestTag),
			IncludeVersionTags: types.BoolValue(b.Config.IncludeVersionTags),
			IncludeCommitTag:   types.BoolValue(b.Config.IncludeCommitTag),
			Registries:         regs,
		}
	}

	links, _ := types.ListValueFrom(ctx, types.StringType, b.Config.Links)
	data.Links = links

	strOrNull := func(s string) types.String {
		if s != "" {
			return types.StringValue(s)
		}
		return types.StringNull()
	}

	// Source block: keep nil when all git fields are empty/default and caller didn't set it.
	if data.Source != nil || b.Config.Repo != "" || b.Config.Branch != "" ||
		b.Config.GitProvider != "" || b.Config.GitAccount != "" || b.Config.Commit != "" ||
		b.Config.LinkedRepo != "" {
		// When repo_id is set, URL/account_id/path/branch/commit are derived from the
		// linked repo by the API and must not be persisted, to avoid permanent diffs.
		var urlVal types.String
		accountIDVal := types.StringNull()
		pathVal := types.StringNull()
		branchVal := types.StringNull()
		commitVal := types.StringNull()
		if b.Config.LinkedRepo == "" {
			// Only reconstruct url if the prior state/plan had it explicitly set.
			// If user omitted url (relying on account_id for domain), keep it null.
			priorURLSet := data.Source != nil && !data.Source.URL.IsNull()
			if b.Config.GitProvider != "" && priorURLSet {
				if b.Config.GitHttps {
					urlVal = types.StringValue("https://" + b.Config.GitProvider)
				} else {
					urlVal = types.StringValue("http://" + b.Config.GitProvider)
				}
			} else {
				urlVal = types.StringNull()
			}
			id := c.ResolveGitAccountID(ctx, b.Config.GitProvider, b.Config.GitAccount)
			if id != "" {
				accountIDVal = types.StringValue(id)
			}
			pathVal = strOrNull(b.Config.Repo)
			branchVal = strOrNull(b.Config.Branch)
			commitVal = strOrNull(b.Config.Commit)
		} else {
			urlVal = types.StringNull()
		}
		data.Source = &BuildSourceModel{
			RepoID:    strOrNull(b.Config.LinkedRepo),
			URL:       urlVal,
			AccountID: accountIDVal,
			Path:      pathVal,
			Branch:    branchVal,
			Commit:    commitVal,
		}
	} else {
		data.Source = nil
	}
	webhookSecret := types.StringNull()
	if b.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(b.Config.WebhookSecret)
	}
	if b.Config.WebhookEnabled || b.Config.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(b.Config.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
	data.FilesOnHost = types.BoolValue(b.Config.FilesOnHost)
	if data.Build != nil {
		extraArgs, _ := types.ListValueFrom(ctx, types.StringType, b.Config.ExtraArgs)
		data.Build = &DockerBuildModel{
			Path:       types.StringValue(b.Config.BuildPath),
			ExtraArgs:  extraArgs,
			Args:       types.StringValue(b.Config.BuildArgs),
			SecretArgs: types.StringValue(b.Config.SecretArgs),
		}
	}
	data.DockerfilePath = types.StringValue(b.Config.DockerfilePath)

	data.SkipSecretInterp = types.BoolValue(b.Config.SkipSecretInterp)
	data.UseBuildx = types.BoolValue(b.Config.UseBuildx)

	// pre_build: only show when non-empty.
	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: types.StringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else if data.PreBuild != nil {
		// Preserve user-set block even if both fields were empty.
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: types.StringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else {
		data.PreBuild = nil
	}

	data.Dockerfile = types.StringValue(b.Config.Dockerfile)
	data.Labels = types.StringValue(b.Config.Labels)
}
