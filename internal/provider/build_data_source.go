// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &BuildDataSource{}
var _ datasource.DataSourceWithValidateConfig = &BuildDataSource{}

func NewBuildDataSource() datasource.DataSource {
	return &BuildDataSource{}
}

type BuildDataSource struct {
	client *client.Client
}

type BuildDataSourceModel struct {
	ID               types.String        `tfsdk:"id"`
	Name             types.String        `tfsdk:"name"`
	BuilderID        types.String        `tfsdk:"builder_id"`
	Version          *BuildVersionModel  `tfsdk:"version"`
	Image            *BuildImageModel    `tfsdk:"image"`
	Links            types.List          `tfsdk:"links"`
	LinkedRepo       types.String        `tfsdk:"linked_repo"`
	GitProvider      types.String        `tfsdk:"git_provider"`
	GitHttps         types.Bool          `tfsdk:"git_https"`
	GitAccount       types.String        `tfsdk:"git_account"`
	Repo             types.String        `tfsdk:"repo"`
	Branch           types.String        `tfsdk:"branch"`
	Commit           types.String        `tfsdk:"commit"`
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

func (d *BuildDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (d *BuildDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The working directory for the command.",
		},
		"command": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The shell command to run.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo build resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The build identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique name of the build. One of `name` or `id` must be set.",
			},
			"builder_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the builder used.",
			},
			"version": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Semantic version and auto-increment settings for the built image.",
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The current semantic version, e.g. `1.0.0`.",
					},
					"auto_increment_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the patch version is automatically incremented on each build.",
					},
				},
			},
			"image": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Image configuration for the build output.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Override for the image name.",
					},
					"tag": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Extra tag suffix applied to the built image.",
					},
					"include_latest_tag_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether a `:latest` tag is pushed.",
					},
					"include_version_tags_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether individual semver component tags are pushed.",
					},
					"include_commit_tag_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether a git commit hash tag is pushed.",
					},
					"registry": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Image registry configurations the built image is pushed to.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"account_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The ID of the `komodo_registry_account` used with this registry.",
								},
								"organization": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Organization name within the registry account.",
								},
							},
						},
					},
				},
			},
			"links": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links associated with this build.",
			},
			"linked_repo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the linked Komodo Repo resource.",
			},
			"git_provider": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Git provider domain.",
			},
			"git_https": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether HTTPS is used for git access.",
			},
			"git_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Git account used for private repositories.",
			},
			"repo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Repository path.",
			},
			"branch": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Branch built from.",
			},
			"commit": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Specific commit hash built.",
			},
			"webhook": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook configuration for the build.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether webhook triggering is enabled.",
					},
					"secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "The webhook secret override for this build.",
					},
				},
			},
			"files_on_host": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether host filesystem files are used instead of a git repository.",
			},
			"build": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Docker build configuration.",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Path to the Docker build context directory.",
					},
					"extra_args": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Additional arguments passed to `docker build`.",
					},
					"args": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Docker build arguments.",
					},
					"secret_args": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Docker secret build arguments.",
					},
				},
			},
			"dockerfile_path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Path to the Dockerfile.",
			},
			"skip_secret_interp": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secret interpolation in build args is skipped.",
			},
			"use_buildx": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether `docker buildx` is used.",
			},
			"pre_build": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Command run before the Docker build.",
				Attributes:          systemCommandAttrs,
			},
			"dockerfile": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Inline Dockerfile contents.",
			},
			"labels": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Docker image labels.",
			},
		},
	}
}

func (d *BuildDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *BuildDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data BuildDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Name.IsUnknown() || data.ID.IsUnknown() {
		return
	}
	nameSet := !data.Name.IsNull()
	idSet := !data.ID.IsNull()
	if nameSet && idSet {
		resp.Diagnostics.AddError("Invalid Configuration", "Only one of `name` or `id` may be set, not both.")
		return
	}
	if !nameSet && !idSet {
		resp.Diagnostics.AddError("Invalid Configuration", "One of `name` or `id` must be set.")
	}
}

func (d *BuildDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BuildDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading build", map[string]interface{}{"lookup": lookup})
	b, err := d.client.GetBuild(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read build, got error: %s", err))
		return
	}
	if b == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Build %q not found.", lookup))
		return
	}

	data.ID = types.StringValue(b.ID.OID)
	data.Name = types.StringValue(b.Name)
	data.BuilderID = types.StringValue(b.Config.BuilderID)
	v := b.Config.Version
	data.Version = &BuildVersionModel{
		Value:                types.StringValue(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)),
		AutoIncrementEnabled: types.BoolValue(b.Config.AutoIncrementVersion),
	}
	{
		var regs []ImageRegistryConfigModel
		if len(b.Config.ImageRegistry) > 0 {
			regs = make([]ImageRegistryConfigModel, len(b.Config.ImageRegistry))
			for i, r := range b.Config.ImageRegistry {
				accountID := d.client.ResolveDockerRegistryAccountID(ctx, r.Domain, r.Account)
				regs[i] = ImageRegistryConfigModel{
					Account:      types.StringValue(accountID),
					Organization: types.StringValue(r.Organization),
				}
			}
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
	{
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

	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: types.StringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else {
		data.PreBuild = nil
	}

	data.Dockerfile = types.StringValue(b.Config.Dockerfile)
	data.Labels = types.StringValue(b.Config.Labels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
