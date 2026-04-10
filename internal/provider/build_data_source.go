// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &BuildDataSource{}

func NewBuildDataSource() datasource.DataSource {
	return &BuildDataSource{}
}

type BuildDataSource struct {
	client *client.Client
}

type BuildDataSourceModel struct {
	ID                   types.String               `tfsdk:"id"`
	Name                 types.String               `tfsdk:"name"`
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
	WebhookEnabled       types.Bool                 `tfsdk:"webhook_enabled"`
	WebhookSecret        types.String               `tfsdk:"webhook_secret"`
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
				Required:            true,
				MarkdownDescription: "The build identifier (ObjectId or name).",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique name of the build.",
			},
			"builder_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the builder used.",
			},
			"version": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The current semantic version of the built image.",
				Attributes: map[string]schema.Attribute{
					"major": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Major version component.",
					},
					"minor": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Minor version component.",
					},
					"patch": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Patch version component.",
					},
				},
			},
			"auto_increment_version": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the patch version is automatically incremented on each build.",
			},
			"image_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Override for the image name.",
			},
			"image_tag": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extra tag suffix applied to the built image.",
			},
			"include_latest_tag": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether a `:latest` tag is pushed.",
			},
			"include_version_tags": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether individual semver component tags are pushed.",
			},
			"include_commit_tag": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether a git commit hash tag is pushed.",
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
			"webhook_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether webhook triggering is enabled.",
			},
			"webhook_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The webhook secret override for this build.",
			},
			"files_on_host": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether host filesystem files are used instead of a git repository.",
			},
			"build_path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Path to the Docker build context directory.",
			},
			"dockerfile_path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Path to the Dockerfile.",
			},
			"image_registry": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Image registry configurations the built image is pushed to.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Registry provider domain.",
						},
						"account": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Account used with this registry.",
						},
						"organization": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Organization name within the registry account.",
						},
					},
				},
			},
			"skip_secret_interp": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secret interpolation in build args is skipped.",
			},
			"use_buildx": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether `docker buildx` is used.",
			},
			"extra_args": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional arguments passed to `docker build`.",
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
			"build_args": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Docker build arguments.",
			},
			"secret_args": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Docker secret build arguments.",
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

func (d *BuildDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BuildDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading build", map[string]interface{}{"id": data.ID.ValueString()})
	b, err := d.client.GetBuild(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read build, got error: %s", err))
		return
	}
	if b == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Build %q not found.", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(b.Name)
	data.BuilderID = types.StringValue(b.Config.BuilderID)
	data.Version = &BuildVersionModel{
		Major: types.Int64Value(int64(b.Config.Version.Major)),
		Minor: types.Int64Value(int64(b.Config.Version.Minor)),
		Patch: types.Int64Value(int64(b.Config.Version.Patch)),
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
	data.WebhookEnabled = types.BoolValue(b.Config.WebhookEnabled)
	data.WebhookSecret = types.StringValue(b.Config.WebhookSecret)
	data.FilesOnHost = types.BoolValue(b.Config.FilesOnHost)
	data.BuildPath = types.StringValue(b.Config.BuildPath)
	data.DockerfilePath = types.StringValue(b.Config.DockerfilePath)

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
	} else {
		data.ImageRegistry = nil
	}

	data.SkipSecretInterp = types.BoolValue(b.Config.SkipSecretInterp)
	data.UseBuildx = types.BoolValue(b.Config.UseBuildx)
	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, b.Config.ExtraArgs)
	data.ExtraArgs = extraArgs

	if b.Config.PreBuild.Path != "" || b.Config.PreBuild.Command != "" {
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
