// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
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
	Dockerfile         *DockerfileModel           `tfsdk:"dockerfile"`
}

// BuildArgumentModel holds a single build argument for a Docker build.
// When SecretEnabled is true the argument is passed as a Docker secret
// (--secret id=NAME,env=VALUE) instead of a plain build-arg.
type BuildArgumentModel struct {
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	SecretEnabled types.Bool   `tfsdk:"secret_enabled"`
}

// DockerBuildModel is the Terraform model for the build block of a build.
type DockerBuildModel struct {
	Path           types.String         `tfsdk:"path"`
	ExtraArguments types.List           `tfsdk:"extra_arguments"`
	Arguments      []BuildArgumentModel `tfsdk:"argument"`
	UseBuildx      types.Bool           `tfsdk:"buildx_enabled"`
}

// DockerfileModel is the Terraform model for the dockerfile block of a build.
type DockerfileModel struct {
	Contents TrimmedStringValue `tfsdk:"contents"`
	Path     types.String       `tfsdk:"path"`
}

// BuildSourceModel is the Terraform model for the source block of a build.
type BuildSourceModel struct {
	RepoID      types.String `tfsdk:"repo_id"`
	URL         types.String `tfsdk:"url"`
	AccountID   types.String `tfsdk:"account_id"`
	Path        types.String `tfsdk:"path"`
	Branch      types.String `tfsdk:"branch"`
	Commit      types.String `tfsdk:"commit"`
	FilesOnHost types.Bool   `tfsdk:"on_host_enabled"`
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
	Build            *DockerBuildModel   `tfsdk:"build"`
	SkipSecretInterp types.Bool          `tfsdk:"skip_secret_interpolation_enabled"`
	PreBuild         *SystemCommandModel `tfsdk:"pre_build"`
	Labels           TrimmedStringValue  `tfsdk:"labels"`
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
			CustomType:          TrimmedStringType{},
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
			"skip_secret_interpolation_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to skip secret interpolation in build args.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				CustomType:          TrimmedStringType{},
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
					"dockerfile": schema.SingleNestedBlock{
						MarkdownDescription: "Dockerfile configuration.",
						Attributes: map[string]schema.Attribute{
							"contents": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								CustomType:          TrimmedStringType{},
								MarkdownDescription: "Inline Dockerfile contents. Overrides `path` when set.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"path": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Path to the Dockerfile relative to `build.path`.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
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
					"on_host_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Whether to use files on the host filesystem for the build context instead of a git repository.",
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
					"extra_arguments": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Additional arguments to pass to the `docker build` command.",
					},
					"buildx_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to use `docker buildx` for multi-platform builds.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"argument": schema.ListNestedBlock{
						MarkdownDescription: "Docker build argument. Set `secret_enabled = true` to pass it as a Docker secret (`--secret id=NAME,env=VALUE`) instead of a plain build-arg (`--build-arg NAME=VALUE`).",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The build argument name.",
								},
								"value": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The build argument value.",
								},
								"secret_enabled": schema.BoolAttribute{
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
									MarkdownDescription: "When `true`, passes this argument as a Docker secret instead of a plain build-arg. Defaults to `false`.",
								},
							},
						},
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
	} else {
		// Explicitly clear all git source fields when source block is removed.
		empty := ""
		f := false
		cfg.LinkedRepo = &empty
		cfg.GitProvider = &empty
		cfg.GitHttps = &f
		cfg.GitAccount = &empty
		cfg.Repo = &empty
		cfg.Branch = &empty
		cfg.Commit = &empty
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
	if data.Source != nil && !data.Source.FilesOnHost.IsNull() && !data.Source.FilesOnHost.IsUnknown() {
		v := data.Source.FilesOnHost.ValueBool()
		cfg.FilesOnHost = &v
	}
	if !data.SkipSecretInterp.IsNull() && !data.SkipSecretInterp.IsUnknown() {
		v := data.SkipSecretInterp.ValueBool()
		cfg.SkipSecretInterp = &v
	}
	if data.Build != nil {
		if !data.Build.Path.IsNull() && !data.Build.Path.IsUnknown() {
			v := data.Build.Path.ValueString()
			cfg.BuildPath = &v
		}
		var extraArgSlice []string
		if !data.Build.ExtraArguments.IsNull() && !data.Build.ExtraArguments.IsUnknown() {
			data.Build.ExtraArguments.ElementsAs(ctx, &extraArgSlice, false)
		}
		if extraArgSlice == nil {
			extraArgSlice = []string{}
		}
		cfg.ExtraArgs = &extraArgSlice
		var plainBuildArgs, secretBuildArgs []BuildArgumentModel
		for _, a := range data.Build.Arguments {
			if a.SecretEnabled.ValueBool() {
				secretBuildArgs = append(secretBuildArgs, a)
			} else {
				plainBuildArgs = append(plainBuildArgs, a)
			}
		}
		{
			v := buildArgsToString(plainBuildArgs)
			cfg.BuildArgs = &v
		}
		{
			v := buildArgsToString(secretBuildArgs)
			cfg.SecretArgs = &v
		}
		if !data.Build.UseBuildx.IsNull() && !data.Build.UseBuildx.IsUnknown() {
			v := data.Build.UseBuildx.ValueBool()
			cfg.UseBuildx = &v
		}
	}
	if data.PreBuild != nil {
		cfg.PreBuild = &client.SystemCommand{
			Path:    data.PreBuild.Path.ValueString(),
			Command: data.PreBuild.Command.ValueString(),
		}
	} else {
		emptyCmd := client.SystemCommand{Path: "", Command: ""}
		cfg.PreBuild = &emptyCmd
	}
	if data.Image != nil && data.Image.Dockerfile != nil {
		if !data.Image.Dockerfile.Path.IsNull() && !data.Image.Dockerfile.Path.IsUnknown() {
			v := data.Image.Dockerfile.Path.ValueString()
			cfg.DockerfilePath = &v
		}
		if !data.Image.Dockerfile.Contents.IsNull() && !data.Image.Dockerfile.Contents.IsUnknown() {
			v := data.Image.Dockerfile.Contents.ValueString()
			cfg.Dockerfile = &v
		}
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
		var docfile *DockerfileModel
		if data.Image.Dockerfile != nil {
			docfile = &DockerfileModel{
				Path:     types.StringValue(b.Config.DockerfilePath),
				Contents: NewTrimmedStringValue(strings.TrimRight(b.Config.Dockerfile, "\n\r")),
			}
		}
		data.Image = &BuildImageModel{
			Name:               types.StringValue(b.Config.ImageName),
			Tag:                types.StringValue(b.Config.ImageTag),
			IncludeLatestTag:   types.BoolValue(b.Config.IncludeLatestTag),
			IncludeVersionTags: types.BoolValue(b.Config.IncludeVersionTags),
			IncludeCommitTag:   types.BoolValue(b.Config.IncludeCommitTag),
			Registries:         regs,
			Dockerfile:         docfile,
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
	if data.Source != nil || b.Config.Repo != "" || b.Config.LinkedRepo != "" || b.Config.FilesOnHost {
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
			RepoID:      strOrNull(b.Config.LinkedRepo),
			URL:         urlVal,
			AccountID:   accountIDVal,
			Path:        pathVal,
			Branch:      branchVal,
			Commit:      commitVal,
			FilesOnHost: types.BoolValue(b.Config.FilesOnHost),
		}
	} else {
		data.Source = nil
	}
	webhookSecret := types.StringNull()
	if b.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(b.Config.WebhookSecret)
	}
	if b.Config.WebhookEnabled || b.Config.WebhookSecret != "" || data.Webhook != nil {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(b.Config.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
	if data.Build != nil {
		var extraArgs types.List
		if len(b.Config.ExtraArgs) > 0 {
			extraArgs, _ = types.ListValueFrom(ctx, types.StringType, b.Config.ExtraArgs)
		} else {
			extraArgs = types.ListNull(types.StringType)
		}
		plain := parseBuildArguments(b.Config.BuildArgs, false)
		secret := parseBuildArguments(b.Config.SecretArgs, true)
		allArgs := append(plain, secret...)
		sort.Slice(allArgs, func(i, j int) bool {
			return allArgs[i].Name.ValueString() < allArgs[j].Name.ValueString()
		})
		allArgs = matchPriorOrder(data.Build.Arguments, allArgs)
		data.Build = &DockerBuildModel{
			Path:           types.StringValue(b.Config.BuildPath),
			ExtraArguments: extraArgs,
			Arguments:      allArgs,
			UseBuildx:      types.BoolValue(b.Config.UseBuildx),
		}
	}
	data.SkipSecretInterp = types.BoolValue(b.Config.SkipSecretInterp)

	// pre_build: only show when non-empty.
	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
		data.PreBuild = &SystemCommandModel{
			Path:    strOrNull(b.Config.PreBuild.Path),
			Command: NewTrimmedStringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else if data.PreBuild != nil {
		// Preserve user-set block even if both fields were empty.
		data.PreBuild = &SystemCommandModel{
			Path:    strOrNull(b.Config.PreBuild.Path),
			Command: NewTrimmedStringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else {
		data.PreBuild = nil
	}

	data.Labels = NewTrimmedStringValue(strings.TrimRight(b.Config.Labels, "\n\r"))
}

// buildArgsToString serialises a slice of BuildArgumentModel to the KEY=VALUE\n
// format expected by the Komodo API's build_args / secret_args fields.
func buildArgsToString(args []BuildArgumentModel) string {
	sorted := make([]BuildArgumentModel, len(args))
	copy(sorted, args)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name.ValueString() < sorted[j].Name.ValueString()
	})
	var sb strings.Builder
	for _, a := range sorted {
		sb.WriteString(a.Name.ValueString())
		sb.WriteByte('=')
		sb.WriteString(a.Value.ValueString())
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// parseBuildArguments parses a KEY=VALUE\n string returned by the API back
// into a sorted slice of BuildArgumentModel. secretEnabled controls the
// SecretEnabled field set on every returned entry.
func parseBuildArguments(raw string, secretEnabled bool) []BuildArgumentModel {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var result []BuildArgumentModel
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		result = append(result, BuildArgumentModel{
			Name:          types.StringValue(strings.TrimSpace(line[:idx])),
			Value:         types.StringValue(strings.TrimSpace(line[idx+1:])),
			SecretEnabled: types.BoolValue(secretEnabled),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name.ValueString() < result[j].Name.ValueString()
	})
	return result
}

// matchPriorOrder returns apiArgs reordered so that entries matching prior by
// name appear in the same relative order as in prior. Entries not found in
// prior (newly added) are appended at the end in their original order.
func matchPriorOrder(prior, apiArgs []BuildArgumentModel) []BuildArgumentModel {
	if len(prior) == 0 {
		return apiArgs
	}
	// Build name→index map for API results.
	idx := make(map[string]int, len(apiArgs))
	for i, a := range apiArgs {
		idx[a.Name.ValueString()] = i
	}
	used := make([]bool, len(apiArgs))
	result := make([]BuildArgumentModel, 0, len(apiArgs))
	for _, p := range prior {
		i, ok := idx[p.Name.ValueString()]
		if ok {
			result = append(result, apiArgs[i])
			used[i] = true
		}
	}
	for i, a := range apiArgs {
		if !used[i] {
			result = append(result, a)
		}
	}
	return result
}
