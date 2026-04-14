// Copyright IBM Corp. 2026
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

var _ datasource.DataSource = &SwarmDataSource{}
var _ datasource.DataSourceWithValidateConfig = &SwarmDataSource{}

func NewSwarmDataSource() datasource.DataSource {
	return &SwarmDataSource{}
}

type SwarmDataSource struct {
	client *client.Client
}

type SwarmDataSourceModel struct {
	ID            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	Tags          types.List               `tfsdk:"tags"`
	ServerIDs     types.List               `tfsdk:"server_ids"`
	Links         types.List               `tfsdk:"links"`
	AlertsEnabled types.Bool               `tfsdk:"alerts_enabled"`
	Maintenance   []MaintenanceWindowModel `tfsdk:"maintenance"`
}

func (d *SwarmDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_swarm"
}

func (d *SwarmDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Komodo Docker Swarm by name or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The swarm identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The swarm name. One of `name` or `id` must be set.",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tag IDs attached to this swarm.",
			},
			"server_ids": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of the servers that are manager nodes of this swarm.",
			},
			"links": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links displayed in the Komodo UI for this swarm.",
			},
			"alerts_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether alerts are sent when the swarm is unhealthy.",
			},
		},
		Blocks: map[string]schema.Block{
			"maintenance": schema.ListNestedBlock{
				MarkdownDescription: "Scheduled maintenance windows for this swarm.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the maintenance window.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of the maintenance window.",
						},
						"schedule_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Schedule type: `Daily`, `Weekly`, or `OneTime`.",
						},
						"day_of_week": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "For `Weekly` schedules: day of the week.",
						},
						"date": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "For `OneTime` windows: ISO 8601 date.",
						},
						"hour": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Start hour in 24-hour format.",
						},
						"minute": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Start minute.",
						},
						"duration_minutes": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Duration of the maintenance window in minutes.",
						},
						"timezone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timezone for the maintenance window.",
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this maintenance window is active.",
						},
					},
				},
			},
		},
	}
}

func (d *SwarmDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SwarmDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data SwarmDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Name.IsUnknown() || data.ID.IsUnknown() {
		return
	}
	nameSet := !data.Name.IsNull() && !data.Name.IsUnknown()
	idSet := !data.ID.IsNull() && !data.ID.IsUnknown()
	if nameSet && idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of `name` or `id` may be set, not both.",
		)
		return
	}
	if !nameSet && !idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of `name` or `id` must be set.",
		)
	}
}

func (d *SwarmDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SwarmDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Reading swarm", map[string]interface{}{"lookup": lookup})

	swarm, err := d.client.GetSwarm(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read swarm, got error: %s", err))
		return
	}
	if swarm == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Swarm %q not found", lookup))
		return
	}

	swarmToDataSourceModel(ctx, swarm, &data)

	tflog.Trace(ctx, "Read swarm data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func swarmToDataSourceModel(ctx context.Context, s *client.Swarm, data *SwarmDataSourceModel) {
	data.ID = types.StringValue(s.ID.OID)
	data.Name = types.StringValue(s.Name)

	tagsSlice := s.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, tagsSlice)

	cfg := s.Config

	if cfg.ServerIDs != nil {
		data.ServerIDs, _ = types.ListValueFrom(ctx, types.StringType, cfg.ServerIDs)
	} else {
		data.ServerIDs, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	if cfg.Links != nil {
		data.Links, _ = types.ListValueFrom(ctx, types.StringType, cfg.Links)
	} else {
		data.Links, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	data.AlertsEnabled = types.BoolValue(cfg.AlertsEnabled)

	if len(cfg.MaintenanceWindows) > 0 {
		windows := make([]MaintenanceWindowModel, len(cfg.MaintenanceWindows))
		for i, w := range cfg.MaintenanceWindows {
			windows[i] = MaintenanceWindowModel{
				Name:            types.StringValue(w.Name),
				Description:     types.StringValue(w.Description),
				ScheduleType:    types.StringValue(w.ScheduleType),
				DayOfWeek:       types.StringValue(w.DayOfWeek),
				Date:            types.StringValue(w.Date),
				Hour:            types.Int64Value(w.Hour),
				Minute:          types.Int64Value(w.Minute),
				DurationMinutes: types.Int64Value(w.DurationMinutes),
				Timezone:        types.StringValue(w.Timezone),
				Enabled:         types.BoolValue(w.Enabled),
			}
		}
		data.Maintenance = windows
	} else {
		data.Maintenance = []MaintenanceWindowModel{}
	}
}
