// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &DeploymentDataSource{}
var _ datasource.DataSourceWithValidateConfig = &DeploymentDataSource{}

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

type DeploymentDataSource struct {
	client *client.Client
}

// DeploymentDataSourceModel is the Terraform data source model for komodo_deployment.
type DeploymentDataSourceModel struct {
	ID                             types.String                `tfsdk:"id"`
	Name                           types.String                `tfsdk:"name"`
	SwarmID                        types.String                `tfsdk:"swarm_id"`
	ServerID                       types.String                `tfsdk:"server_id"`
	Image                          *DeploymentImageModel       `tfsdk:"image"`
	SkipSecretInterpolationEnabled types.Bool                  `tfsdk:"secret_interpolation_enabled"`
	PollForUpdatesEnabled          types.Bool                  `tfsdk:"poll_updates_enabled"`
	AutoUpdateEnabled              types.Bool                  `tfsdk:"auto_update_enabled"`
	SendAlertsEnabled              types.Bool                  `tfsdk:"alerts_enabled"`
	Container                      *DeploymentContainerModel   `tfsdk:"container"`
	Termination                    *DeploymentTerminationModel `tfsdk:"termination"`
}

func (d *DeploymentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *DeploymentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Komodo deployment resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The deployment identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the deployment. One of `name` or `id` must be set.",
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
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Docker image. Set when `type` is `Image`.",
					},
					"build_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Komodo Build ID. Set when `type` is `Build`.",
					},
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Build version string (e.g. `1.0.0`). Set when `type` is `Build`.",
					},
					"account_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Account used to pull the image.",
					},
					"redeploy_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the deployment redeploys when its Build finishes.",
					},
				},
			},
			"secret_interpolation_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secrets are interpolated into deployment environment variables.",
			},
			"poll_updates_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether image updates are polled for.",
			},
			"auto_update_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the deployment auto-updates when a newer image is found.",
			},
			"alerts_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether ContainerStateChange alerts are sent.",
			},
			"container": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Docker container runtime configuration.",
				Attributes: map[string]schema.Attribute{
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
					"extra_arguments": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Extra arguments for `docker run`.",
					},
					"ports": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Container port mappings.",
					},
					"volumes": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Container volume mappings.",
					},
					"environment": schema.MapAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Environment variables for the container.",
					},
					"labels": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Docker labels for the container as a list of `key=value` strings.",
					}, "links": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Quick links displayed in the resource header.",
					}},
			},
			"termination": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Container termination behaviour.",
				Attributes: map[string]schema.Attribute{
					"signal": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Default termination signal.",
					},
					"timeout": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Termination timeout in seconds.",
					},
					"signal_labels": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Labels for termination signal options.",
					},
				},
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

func (d *DeploymentDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data DeploymentDataSourceModel
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

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading deployment", map[string]interface{}{"lookup": lookup})

	dep, err := d.client.GetDeployment(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read deployment, got error: %s", err))
		return
	}
	if dep == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Deployment %q not found.", lookup))
		return
	}

	data.ID = types.StringValue(dep.ID.OID)
	data.Name = types.StringValue(dep.Name)
	data.SwarmID = types.StringValue(dep.Config.SwarmID)
	data.ServerID = types.StringValue(dep.Config.ServerID)

	// image block: always populate from API.
	if dep.Config.Image.Build != nil {
		bv := dep.Config.Image.Build.Version
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Build"),
			Image:   types.StringValue(""),
			BuildID: types.StringValue(dep.Config.Image.Build.BuildID),
			Version: types.StringValue(fmt.Sprintf("%d.%d.%d", bv.Major, bv.Minor, bv.Patch)),
		}
	} else if dep.Config.Image.Image != nil {
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Image"),
			Image:   types.StringValue(dep.Config.Image.Image.Image),
			BuildID: types.StringValue(""),
			Version: types.StringNull(),
		}
	} else {
		data.Image = nil
	}
	if data.Image != nil {
		data.Image.RegistryAccount = types.StringValue(dep.Config.ImageRegistryAccount)
		data.Image.RedeployEnabled = types.BoolValue(dep.Config.RedeployOnBuild)
	}

	data.SkipSecretInterpolationEnabled = types.BoolValue(!dep.Config.SkipSecretInterpolation)
	data.PollForUpdatesEnabled = types.BoolValue(dep.Config.PollForUpdates)
	data.AutoUpdateEnabled = types.BoolValue(dep.Config.AutoUpdate)
	data.SendAlertsEnabled = types.BoolValue(dep.Config.SendAlerts)

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, dep.Config.ExtraArguments)
	data.Container = &DeploymentContainerModel{
		Network:        types.StringValue(dep.Config.Network),
		Restart:        types.StringValue(dep.Config.Restart),
		Command:        types.StringValue(dep.Config.Command),
		Replicas:       types.Int64Value(int64(dep.Config.Replicas)),
		ExtraArguments: extraArgs,
		Ports: func() types.List {
			if raw := strings.TrimRight(dep.Config.Ports, "\n"); raw != "" {
				items := strings.Split(raw, "\n")
				v, _ := types.ListValueFrom(ctx, types.StringType, items)
				return v
			}
			return types.ListValueMust(types.StringType, []attr.Value{})
		}(),
		Volumes: func() types.List {
			if raw := strings.TrimRight(dep.Config.Volumes, "\n"); raw != "" {
				items := strings.Split(raw, "\n")
				v, _ := types.ListValueFrom(ctx, types.StringType, items)
				return v
			}
			return types.ListValueMust(types.StringType, []attr.Value{})
		}(),
		Environment: func() types.Map {
			m := envStringToMap(strings.TrimRight(dep.Config.Environment, "\n"))
			if m.IsNull() {
				return types.MapValueMust(types.StringType, map[string]attr.Value{})
			}
			return m
		}(),
		Labels: func() types.List {
			if rawLabels := strings.TrimRight(dep.Config.Labels, "\n"); rawLabels != "" {
				items := strings.Split(rawLabels, "\n")
				v, _ := types.ListValueFrom(ctx, types.StringType, items)
				return v
			}
			return types.ListValueMust(types.StringType, []attr.Value{})
		}(),
		Links: func() types.List {
			v, _ := types.ListValueFrom(ctx, types.StringType, dep.Config.Links)
			return v
		}(),
	}

	data.Termination = &DeploymentTerminationModel{
		Signal:       types.StringValue(dep.Config.TerminationSignal),
		Timeout:      types.Int64Value(int64(dep.Config.TerminationTimeout)),
		SignalLabels: types.StringValue(dep.Config.TerminationSignalLabels),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
