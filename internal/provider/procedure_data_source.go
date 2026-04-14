// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &ProcedureDataSource{}
var _ datasource.DataSourceWithValidateConfig = &ProcedureDataSource{}

func NewProcedureDataSource() datasource.DataSource {
	return &ProcedureDataSource{}
}

type ProcedureDataSource struct {
	client *client.Client
}

type ProcedureDataSourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Stages       types.String   `tfsdk:"stages"`
	Schedule     *ScheduleModel `tfsdk:"schedule"`
	FailureAlert types.Bool     `tfsdk:"failure_alert_enabled"`
	Webhook      *WebhookModel  `tfsdk:"webhook"`
}

func (d *ProcedureDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (d *ProcedureDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo procedure resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The procedure identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique name of the procedure. One of `name` or `id` must be set.",
			},
			"stages": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON array of procedure stages.",
			},
			"schedule": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Schedule configuration for the procedure.",
				Attributes: map[string]schema.Attribute{
					"format": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The schedule format: `Cron` or `English`.",
					},
					"expression": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The schedule expression.",
					},
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the schedule is enabled.",
					},
					"timezone": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Timezone for the schedule (IANA TZ identifier).",
					},
					"alert_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether an alert is sent on scheduled runs.",
					},
				},
			},
			"failure_alert_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether an alert is sent on procedure failure.",
			},
			"webhook": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook configuration for the procedure.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether webhook triggering is enabled.",
					},
					"secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "The webhook secret override for this procedure.",
					},
				},
			},
		},
	}
}

func (d *ProcedureDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProcedureDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data ProcedureDataSourceModel
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

func (d *ProcedureDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProcedureDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading procedure", map[string]interface{}{"lookup": lookup})
	proc, err := d.client.GetProcedure(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read procedure, got error: %s", err))
		return
	}
	if proc == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Procedure %q not found.", lookup))
		return
	}
	data.ID = types.StringValue(proc.ID.OID)
	data.Name = types.StringValue(proc.Name)

	stagesBytes, _ := json.Marshal(proc.Config.Stages)
	stagesStr := string(stagesBytes)
	if len(proc.Config.Stages) > 0 && stagesStr != "null" && stagesStr != "[]" {
		data.Stages = types.StringValue(stagesStr)
	} else {
		data.Stages = types.StringNull()
	}

	if proc.Config.ScheduleEnabled || proc.Config.Schedule != "" || proc.Config.ScheduleTimezone != "" || proc.Config.ScheduleAlert {
		data.Schedule = &ScheduleModel{
			Format:       types.StringValue(proc.Config.ScheduleFormat),
			Expression:   types.StringValue(proc.Config.Schedule),
			Enabled:      types.BoolValue(proc.Config.ScheduleEnabled),
			Timezone:     types.StringValue(proc.Config.ScheduleTimezone),
			AlertEnabled: types.BoolValue(proc.Config.ScheduleAlert),
		}
	} else {
		data.Schedule = nil
	}
	data.FailureAlert = types.BoolValue(proc.Config.FailureAlert)
	webhookSecret := types.StringNull()
	if proc.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(proc.Config.WebhookSecret)
	}
	data.Webhook = &WebhookModel{
		Enabled: types.BoolValue(proc.Config.WebhookEnabled),
		Secret:  webhookSecret,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
