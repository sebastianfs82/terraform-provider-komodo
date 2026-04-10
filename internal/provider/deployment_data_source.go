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

var _ datasource.DataSource = &DeploymentDataSource{}

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

type DeploymentDataSource struct {
	client *client.Client
}

// DeploymentDataSourceModel is the Terraform data source model for komodo_deployment.
type DeploymentDataSourceModel struct {
	ID                   types.String          `tfsdk:"id"`
	Name                 types.String          `tfsdk:"name"`
	SwarmID              types.String          `tfsdk:"swarm_id"`
	ServerID             types.String          `tfsdk:"server_id"`
	Image                *DeploymentImageModel `tfsdk:"image"`
	ImageRegistryAccount types.String          `tfsdk:"image_registry_account"`
	SkipSecretInterp     types.Bool            `tfsdk:"skip_secret_interp"`
	RedeployOnBuild      types.Bool            `tfsdk:"redeploy_on_build"`
	PollForUpdates       types.Bool            `tfsdk:"poll_for_updates"`
	AutoUpdate           types.Bool            `tfsdk:"auto_update"`
	SendAlerts           types.Bool            `tfsdk:"send_alerts"`
	Links                types.List            `tfsdk:"links"`
	Network              types.String          `tfsdk:"network"`
	Restart              types.String          `tfsdk:"restart"`
	Command              types.String          `tfsdk:"command"`
	Replicas             types.Int64           `tfsdk:"replicas"`
	TerminationSignal    types.String          `tfsdk:"termination_signal"`
	TerminationTimeout   types.Int64           `tfsdk:"termination_timeout"`
	ExtraArgs            types.List            `tfsdk:"extra_args"`
	TermSignalLabels     types.String          `tfsdk:"term_signal_labels"`
	Ports                types.String          `tfsdk:"ports"`
	Volumes              types.String          `tfsdk:"volumes"`
	Environment          types.String          `tfsdk:"environment"`
	Labels               types.String          `tfsdk:"labels"`
}

func (d *DeploymentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *DeploymentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Komodo deployment resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The deployment identifier (ObjectId).",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the deployment.",
			},
			"swarm_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Swarm ID the deployment runs on (Swarm mode).",
			},
			"server_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Server ID the deployment runs on (Container mode).",
			},
			"image": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The image source for this deployment.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Image type: `Image` or `Build`.",
					},
					"image": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Docker image. Set when `type` is `Image`.",
					},
					"build_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Komodo Build ID. Set when `type` is `Build`.",
					},
					"version": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Build version. Set when `type` is `Build`.",
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
				},
			},
			"image_registry_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account used to pull the image.",
			},
			"skip_secret_interp": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secret interpolation is skipped.",
			},
			"redeploy_on_build": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the deployment redeploys when its Build finishes.",
			},
			"poll_for_updates": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether image updates are polled for.",
			},
			"auto_update": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the deployment auto-updates when a newer image is found.",
			},
			"send_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether ContainerStateChange alerts are sent.",
			},
			"links": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links for this deployment.",
			},
			"network": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network attached to the container.",
			},
			"restart": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Restart mode for the container.",
			},
			"command": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Command appended to `docker run`.",
			},
			"replicas": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of replicas (Swarm mode only).",
			},
			"termination_signal": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Default termination signal.",
			},
			"termination_timeout": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Termination timeout in seconds.",
			},
			"extra_args": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra arguments for `docker run`.",
			},
			"term_signal_labels": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Labels for termination signal options.",
			},
			"ports": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Container port mapping.",
			},
			"volumes": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Container volume mapping.",
			},
			"environment": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Environment variables for the container.",
			},
			"labels": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Docker labels for the container.",
			},
		},
	}
}

func (d *DeploymentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading deployment", map[string]interface{}{"id": data.ID.ValueString()})

	dep, err := d.client.GetDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read deployment, got error: %s", err))
		return
	}
	if dep == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Deployment %q not found.", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(dep.ID.OID)
	data.Name = types.StringValue(dep.Name)
	data.SwarmID = types.StringValue(dep.Config.SwarmID)
	data.ServerID = types.StringValue(dep.Config.ServerID)

	// image block: always populate from API.
	if dep.Config.Image.Build != nil {
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Build"),
			Image:   types.StringValue(""),
			BuildID: types.StringValue(dep.Config.Image.Build.BuildID),
			Version: &BuildVersionModel{
				Major: types.Int64Value(int64(dep.Config.Image.Build.Version.Major)),
				Minor: types.Int64Value(int64(dep.Config.Image.Build.Version.Minor)),
				Patch: types.Int64Value(int64(dep.Config.Image.Build.Version.Patch)),
			},
		}
	} else if dep.Config.Image.Image != nil {
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Image"),
			Image:   types.StringValue(dep.Config.Image.Image.Image),
			BuildID: types.StringValue(""),
			Version: nil,
		}
	} else {
		data.Image = nil
	}

	data.ImageRegistryAccount = types.StringValue(dep.Config.ImageRegistryAccount)
	data.SkipSecretInterp = types.BoolValue(dep.Config.SkipSecretInterp)
	data.RedeployOnBuild = types.BoolValue(dep.Config.RedeployOnBuild)
	data.PollForUpdates = types.BoolValue(dep.Config.PollForUpdates)
	data.AutoUpdate = types.BoolValue(dep.Config.AutoUpdate)
	data.SendAlerts = types.BoolValue(dep.Config.SendAlerts)

	links, _ := types.ListValueFrom(ctx, types.StringType, dep.Config.Links)
	data.Links = links

	data.Network = types.StringValue(dep.Config.Network)
	data.Restart = types.StringValue(dep.Config.Restart)
	data.Command = types.StringValue(dep.Config.Command)
	data.Replicas = types.Int64Value(int64(dep.Config.Replicas))
	data.TerminationSignal = types.StringValue(dep.Config.TerminationSignal)
	data.TerminationTimeout = types.Int64Value(int64(dep.Config.TerminationTimeout))

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, dep.Config.ExtraArgs)
	data.ExtraArgs = extraArgs

	data.TermSignalLabels = types.StringValue(dep.Config.TermSignalLabels)
	data.Ports = types.StringValue(dep.Config.Ports)
	data.Volumes = types.StringValue(dep.Config.Volumes)
	data.Environment = types.StringValue(dep.Config.Environment)
	data.Labels = types.StringValue(dep.Config.Labels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
