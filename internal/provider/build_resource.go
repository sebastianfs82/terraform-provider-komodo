// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &BuildResource{}
var _ resource.ResourceWithImportState = &BuildResource{}

func NewBuildResource() resource.Resource {
	return &BuildResource{}
}

type BuildResource struct {
	client *client.Client
}

// BuildVersionModel is the Terraform model for the build version block.
type BuildVersionModel struct {
	Major types.Int64 `tfsdk:"major"`
	Minor types.Int64 `tfsdk:"minor"`
	Patch types.Int64 `tfsdk:"patch"`
}

// ImageRegistryConfigModel is the Terraform model for an image_registry list entry.
type ImageRegistryConfigModel struct {
	Domain       types.String `tfsdk:"domain"`
	Account      types.String `tfsdk:"account"`
	Organization types.String `tfsdk:"organization"`
}

// BuildResourceModel is the Terraform resource model for komodo_build.
type BuildResourceModel struct {
	ID                   types.String               `tfsdk:"id"`
	Name                 types.String               `tfsdk:"name"`
	Tags                 types.List                 `tfsdk:"tags"`
	BuilderID            types.String               `tfsdk:"builder_id"`
	Version              *BuildVersionModel         `tfsdk:"version"`
	AutoIncrementVersion types.Bool                 `tfsdk:"auto_increment_version"`
	ImageName            types.String               `tfsdk:"image_name"`
	ImageTag             types.String               `tfsdk:"image_tag"`
	IncludeLatestTag     types.Bool                 `tfsdk:"include_latest_tag"`
	IncludeVersionTags   types.Bool                 `tfsdk:"include_version_tags"`
	IncludeCommitTag     types.Bool                 `tfsdk:"include_commit_tag"`
	Links                types.List                 `tfsdk:"links"`
	LinkedRepo           types.String               `tfsdk:"linked_repo"`
	GitProvider          types.String               `tfsdk:"git_provider"`
	GitHttps             types.Bool                 `tfsdk:"git_https"`
	GitAccount           types.String               `tfsdk:"git_account"`
	Repo                 types.String               `tfsdk:"repo"`
	Branch               types.String               `tfsdk:"branch"`
	Commit               types.String               `tfsdk:"commit"`
	Webhook              *WebhookModel              `tfsdk:"webhook"`
	FilesOnHost          types.Bool                 `tfsdk:"files_on_host"`
	BuildPath            types.String               `tfsdk:"build_path"`
	DockerfilePath       types.String               `tfsdk:"dockerfile_path"`
	ImageRegistry        []ImageRegistryConfigModel `tfsdk:"image_registry"`
	SkipSecretInterp     types.Bool                 `tfsdk:"skip_secret_interp"`
	UseBuildx            types.Bool                 `tfsdk:"use_buildx"`
	ExtraArgs            types.List                 `tfsdk:"extra_args"`
	PreBuild             *SystemCommandModel        `tfsdk:"pre_build"`
	Dockerfile           types.String               `tfsdk:"dockerfile"`
	BuildArgs            types.String               `tfsdk:"build_args"`
	SecretArgs           types.String               `tfsdk:"secret_args"`
	Labels               types.String               `tfsdk:"labels"`
}

func (r *BuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (r *BuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The working directory for the command.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"command": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The shell command to run.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
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
			"version": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Semantic version for the built image. Managed automatically when `auto_increment_version` is true.",
				Attributes: map[string]schema.Attribute{
					"major": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Major version component.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"minor": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Minor version component.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"patch": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Patch version component.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"auto_increment_version": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to automatically increment the patch version on each build. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"image_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Override for the image name. Defaults to the build name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_tag": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "An extra tag suffix to apply to the image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"include_latest_tag": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to push a `:latest` tag alongside the versioned tag.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"include_version_tags": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to push individual semver component tags (e.g. `:1`, `:1.2`).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"include_commit_tag": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to push a tag with the git commit hash.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links associated with this build.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"linked_repo": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of a Komodo Repo resource to link for source code.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_provider": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Git provider domain, e.g. `github.com`. Defaults to `github.com`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_https": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use HTTPS for git access. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"git_account": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Git account to use for private repositories.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Repository path, e.g. `owner/repo`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Branch to build from.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"commit": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Specific commit hash to build. Overrides branch.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook": schema.SingleNestedAttribute{
				Optional:            true,
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
			"files_on_host": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use files on the host filesystem for the build context instead of a git repository.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"build_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Path to the Docker build context directory. Defaults to `.`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"image_registry": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Image registry configurations to push the built image to.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Registry provider domain, e.g. `docker.io`.",
						},
						"account": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Account to use with this registry.",
						},
						"organization": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Optional organization name within the registry account.",
						},
					},
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
			"extra_args": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional arguments to pass to the `docker build` command.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"pre_build": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A command to run before the Docker build.",
				Attributes:          systemCommandAttrs,
			},
			"dockerfile": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Inline Dockerfile contents. Overrides `dockerfile_path` when set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_args": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Docker build arguments in `KEY=VALUE` format, newline-separated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_args": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Docker secret build arguments. These are not stored in the image layers.",
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
		Config: partialBuildConfigFromModel(ctx, &data),
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
	buildToModel(ctx, b, &data)
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
		tflog.Debug(ctx, "Build not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	buildToModel(ctx, b, &data)
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
		Config: partialBuildConfigFromModel(ctx, &data),
	}
	b, err := r.client.UpdateBuild(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update build, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	buildToModel(ctx, b, &data)
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

// partialBuildConfigFromModel converts the Terraform model into a PartialBuildConfig.
func partialBuildConfigFromModel(ctx context.Context, data *BuildResourceModel) client.PartialBuildConfig {
	cfg := client.PartialBuildConfig{}

	if !data.BuilderID.IsNull() && !data.BuilderID.IsUnknown() {
		v := data.BuilderID.ValueString()
		cfg.BuilderID = &v
	}
	if data.Version != nil {
		cfg.Version = &client.BuildVersion{
			Major: int(data.Version.Major.ValueInt64()),
			Minor: int(data.Version.Minor.ValueInt64()),
			Patch: int(data.Version.Patch.ValueInt64()),
		}
	}
	if !data.AutoIncrementVersion.IsNull() && !data.AutoIncrementVersion.IsUnknown() {
		v := data.AutoIncrementVersion.ValueBool()
		cfg.AutoIncrementVersion = &v
	}
	if !data.ImageName.IsNull() && !data.ImageName.IsUnknown() {
		v := data.ImageName.ValueString()
		cfg.ImageName = &v
	}
	if !data.ImageTag.IsNull() && !data.ImageTag.IsUnknown() {
		v := data.ImageTag.ValueString()
		cfg.ImageTag = &v
	}
	if !data.IncludeLatestTag.IsNull() && !data.IncludeLatestTag.IsUnknown() {
		v := data.IncludeLatestTag.ValueBool()
		cfg.IncludeLatestTag = &v
	}
	if !data.IncludeVersionTags.IsNull() && !data.IncludeVersionTags.IsUnknown() {
		v := data.IncludeVersionTags.ValueBool()
		cfg.IncludeVersionTags = &v
	}
	if !data.IncludeCommitTag.IsNull() && !data.IncludeCommitTag.IsUnknown() {
		v := data.IncludeCommitTag.ValueBool()
		cfg.IncludeCommitTag = &v
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		data.Links.ElementsAs(ctx, &links, false)
		if links == nil {
			links = []string{}
		}
		cfg.Links = &links
	}
	if !data.LinkedRepo.IsNull() && !data.LinkedRepo.IsUnknown() {
		v := data.LinkedRepo.ValueString()
		cfg.LinkedRepo = &v
	}
	if !data.GitProvider.IsNull() && !data.GitProvider.IsUnknown() {
		v := data.GitProvider.ValueString()
		cfg.GitProvider = &v
	}
	if !data.GitHttps.IsNull() && !data.GitHttps.IsUnknown() {
		v := data.GitHttps.ValueBool()
		cfg.GitHttps = &v
	}
	if !data.GitAccount.IsNull() && !data.GitAccount.IsUnknown() {
		v := data.GitAccount.ValueString()
		cfg.GitAccount = &v
	}
	if !data.Repo.IsNull() && !data.Repo.IsUnknown() {
		v := data.Repo.ValueString()
		cfg.Repo = &v
	}
	if !data.Branch.IsNull() && !data.Branch.IsUnknown() {
		v := data.Branch.ValueString()
		cfg.Branch = &v
	}
	if !data.Commit.IsNull() && !data.Commit.IsUnknown() {
		v := data.Commit.ValueString()
		cfg.Commit = &v
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
	if !data.BuildPath.IsNull() && !data.BuildPath.IsUnknown() {
		v := data.BuildPath.ValueString()
		cfg.BuildPath = &v
	}
	if !data.DockerfilePath.IsNull() && !data.DockerfilePath.IsUnknown() {
		v := data.DockerfilePath.ValueString()
		cfg.DockerfilePath = &v
	}
	if data.ImageRegistry != nil {
		regs := make([]client.ImageRegistryConfig, len(data.ImageRegistry))
		for i, r := range data.ImageRegistry {
			regs[i] = client.ImageRegistryConfig{
				Domain:       r.Domain.ValueString(),
				Account:      r.Account.ValueString(),
				Organization: r.Organization.ValueString(),
			}
		}
		cfg.ImageRegistry = &regs
	}
	if !data.SkipSecretInterp.IsNull() && !data.SkipSecretInterp.IsUnknown() {
		v := data.SkipSecretInterp.ValueBool()
		cfg.SkipSecretInterp = &v
	}
	if !data.UseBuildx.IsNull() && !data.UseBuildx.IsUnknown() {
		v := data.UseBuildx.ValueBool()
		cfg.UseBuildx = &v
	}
	if !data.ExtraArgs.IsNull() && !data.ExtraArgs.IsUnknown() {
		var args []string
		data.ExtraArgs.ElementsAs(ctx, &args, false)
		if args == nil {
			args = []string{}
		}
		cfg.ExtraArgs = &args
	}
	if data.PreBuild != nil {
		cfg.PreBuild = &client.SystemCommand{
			Path:    data.PreBuild.Path.ValueString(),
			Command: data.PreBuild.Command.ValueString(),
		}
	}
	if !data.Dockerfile.IsNull() && !data.Dockerfile.IsUnknown() {
		v := data.Dockerfile.ValueString()
		cfg.Dockerfile = &v
	}
	if !data.BuildArgs.IsNull() && !data.BuildArgs.IsUnknown() {
		v := data.BuildArgs.ValueString()
		cfg.BuildArgs = &v
	}
	if !data.SecretArgs.IsNull() && !data.SecretArgs.IsUnknown() {
		v := data.SecretArgs.ValueString()
		cfg.SecretArgs = &v
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		v := data.Labels.ValueString()
		cfg.Labels = &v
	}

	return cfg
}

// buildToModel populates a BuildResourceModel from a Build API response.
func buildToModel(ctx context.Context, b *client.Build, data *BuildResourceModel) {
	data.ID = types.StringValue(b.ID.OID)
	data.Name = types.StringValue(b.Name)
	tagsSlice := b.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	data.Tags = tags
	data.BuilderID = types.StringValue(b.Config.BuilderID)

	// Only populate version block when any component is non-zero, or when the user
	// had already set it (data.Version != nil from prior state).
	if data.Version != nil || b.Config.Version.Major != 0 || b.Config.Version.Minor != 0 || b.Config.Version.Patch != 0 {
		data.Version = &BuildVersionModel{
			Major: types.Int64Value(int64(b.Config.Version.Major)),
			Minor: types.Int64Value(int64(b.Config.Version.Minor)),
			Patch: types.Int64Value(int64(b.Config.Version.Patch)),
		}
	}

	data.AutoIncrementVersion = types.BoolValue(b.Config.AutoIncrementVersion)
	data.ImageName = types.StringValue(b.Config.ImageName)
	data.ImageTag = types.StringValue(b.Config.ImageTag)
	data.IncludeLatestTag = types.BoolValue(b.Config.IncludeLatestTag)
	data.IncludeVersionTags = types.BoolValue(b.Config.IncludeVersionTags)
	data.IncludeCommitTag = types.BoolValue(b.Config.IncludeCommitTag)

	links, _ := types.ListValueFrom(ctx, types.StringType, b.Config.Links)
	data.Links = links

	data.LinkedRepo = types.StringValue(b.Config.LinkedRepo)
	data.GitProvider = types.StringValue(b.Config.GitProvider)
	data.GitHttps = types.BoolValue(b.Config.GitHttps)
	data.GitAccount = types.StringValue(b.Config.GitAccount)
	data.Repo = types.StringValue(b.Config.Repo)
	data.Branch = types.StringValue(b.Config.Branch)
	data.Commit = types.StringValue(b.Config.Commit)
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
	data.BuildPath = types.StringValue(b.Config.BuildPath)
	data.DockerfilePath = types.StringValue(b.Config.DockerfilePath)

	// image_registry: populate only if there are entries or user had set the block.
	if len(b.Config.ImageRegistry) > 0 {
		regs := make([]ImageRegistryConfigModel, len(b.Config.ImageRegistry))
		for i, r := range b.Config.ImageRegistry {
			regs[i] = ImageRegistryConfigModel{
				Domain:       types.StringValue(r.Domain),
				Account:      types.StringValue(r.Account),
				Organization: types.StringValue(r.Organization),
			}
		}
		data.ImageRegistry = regs
	} else if data.ImageRegistry != nil && len(data.ImageRegistry) == 0 {
		// Preserve explicit empty list set by user.
		data.ImageRegistry = []ImageRegistryConfigModel{}
	} else {
		data.ImageRegistry = nil
	}

	data.SkipSecretInterp = types.BoolValue(b.Config.SkipSecretInterp)
	data.UseBuildx = types.BoolValue(b.Config.UseBuildx)

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, b.Config.ExtraArgs)
	data.ExtraArgs = extraArgs

	// pre_build: only show when non-empty.
	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: types.StringValue(b.Config.PreBuild.Command),
		}
	} else if data.PreBuild != nil {
		// Preserve user-set block even if both fields were empty.
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: types.StringValue(b.Config.PreBuild.Command),
		}
	} else {
		data.PreBuild = nil
	}

	data.Dockerfile = types.StringValue(b.Config.Dockerfile)
	data.BuildArgs = types.StringValue(b.Config.BuildArgs)
	data.SecretArgs = types.StringValue(b.Config.SecretArgs)
	data.Labels = types.StringValue(b.Config.Labels)
}
