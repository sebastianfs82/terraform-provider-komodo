// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
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
	FilesOnHost      types.Bool          `tfsdk:"on_host_enabled"`
	Build            *DockerBuildModel   `tfsdk:"build"`
	SkipSecretInterp types.Bool          `tfsdk:"skip_secret_interpolation_enabled"`
	PreBuild         *SystemCommandModel `tfsdk:"pre_build"`
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
			CustomType:          TrimmedStringType{},
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
					"dockerfile": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Dockerfile configuration.",
						Attributes: map[string]schema.Attribute{
							"contents": schema.StringAttribute{
								Computed: true, CustomType: TrimmedStringType{}, MarkdownDescription: "Inline Dockerfile contents.",
							},
							"path": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Path to the Dockerfile.",
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
			"on_host_enabled": schema.BoolAttribute{
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
					"extra_arguments": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Additional arguments passed to `docker build`.",
					},
					"argument": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Docker build arguments.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The build argument name.",
								},
								"value": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The build argument value.",
								},
								"secret_enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether this argument is passed as a Docker secret.",
								},
							},
						},
					},
					"buildx_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether `docker buildx` is used.",
					},
				},
			},
			"skip_secret_interpolation_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secret interpolation in build args is skipped.",
			},
			"pre_build": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Command run before the Docker build.",
				Attributes:          systemCommandAttrs,
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
		var docfile *DockerfileModel
		if b.Config.DockerfilePath != "" || b.Config.Dockerfile != "" {
			docfile = &DockerfileModel{
				Path:     types.StringValue(b.Config.DockerfilePath),
				Contents: NewTrimmedStringValue(b.Config.Dockerfile),
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
		data.Build = &DockerBuildModel{
			Path:           types.StringValue(b.Config.BuildPath),
			ExtraArguments: extraArgs,
			Arguments:      allArgs,
			UseBuildx:      types.BoolValue(b.Config.UseBuildx),
		}
	}
	data.SkipSecretInterp = types.BoolValue(b.Config.SkipSecretInterp)
	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
		data.PreBuild = &SystemCommandModel{
			Path:    types.StringValue(b.Config.PreBuild.Path),
			Command: NewTrimmedStringValue(strings.TrimRight(b.Config.PreBuild.Command, "\n\r")),
		}
	} else {
		data.PreBuild = nil
	}
	data.Labels = types.StringValue(b.Config.Labels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
