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

var _ datasource.DataSource = &AlerterDataSource{}
var _ datasource.DataSourceWithValidateConfig = &AlerterDataSource{}

func NewAlerterDataSource() datasource.DataSource {
	return &AlerterDataSource{}
}

type AlerterDataSource struct {
	client *client.Client
}

type AlerterDataSourceModel struct {
	ID          types.String             `tfsdk:"id"`
	Name        types.String             `tfsdk:"name"`
	Enabled     types.Bool               `tfsdk:"enabled"`
	AlertTypes  types.List               `tfsdk:"types"`
	Resources   []ResourceTargetModel    `tfsdk:"resource"`
	Endpoint    *AlerterEndpointModel    `tfsdk:"endpoint"`
	Maintenance []MaintenanceWindowModel `tfsdk:"maintenance"`
}

func (d *AlerterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerter"
}

func (d *AlerterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo alerter resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The alerter identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique name of the alerter. One of `name` or `id` must be set.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the alerter is enabled.",
			},
			"types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Alert types the alerter is configured to send. Empty means all types.",
			},
			"endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The alerter endpoint configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The endpoint type: `Custom`, `Slack`, `Discord`, `Ntfy`, or `Pushover`.",
					},
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The webhook or endpoint URL.",
					},
					"email": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Optional email address. Only populated when `type` is `Ntfy`.",
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource": schema.ListNestedBlock{
				MarkdownDescription: "Resources filtered by this alerter. `enabled = true` means included, `enabled = false` means excluded.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "If `true`, alerts are sent only for this resource (include). If `false`, alerts are never sent for this resource (exclude).",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The resource type.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name or ID of the resource.",
						},
					},
				},
			},
			"maintenance": schema.ListNestedBlock{
				MarkdownDescription: "Scheduled maintenance windows during which alerts from this alerter will be suppressed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name for the maintenance window.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of what maintenance is performed.",
						},
						"schedule_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Schedule type: `Daily`, `Weekly`, or `OneTime`.",
						},
						"day_of_week": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "For `Weekly` schedules: day of the week (e.g. `Monday`, `Tuesday`).",
						},
						"date": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "For `OneTime` windows: ISO 8601 date in `YYYY-MM-DD` format.",
						},
						"hour": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Start hour in 24-hour format (0–23).",
						},
						"minute": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Start minute (0–59).",
						},
						"duration_minutes": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Duration of the maintenance window in minutes.",
						},
						"timezone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timezone for the maintenance window. If empty, uses the Core timezone.",
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

func (d *AlerterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AlerterDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data AlerterDataSourceModel
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

func (d *AlerterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AlerterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading alerter", map[string]interface{}{"lookup": lookup})
	a, err := d.client.GetAlerter(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alerter, got error: %s", err))
		return
	}
	if a == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Alerter %q not found.", lookup))
		return
	}

	rm := AlerterResourceModel{}
	resp.Diagnostics.Append(alerterToModel(ctx, a, &rm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = rm.ID
	data.Name = rm.Name
	data.Enabled = rm.Enabled
	data.AlertTypes = rm.AlertTypes
	data.Resources = rm.Resources
	data.Endpoint = rm.Endpoint
	data.Maintenance = rm.Maintenance

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
