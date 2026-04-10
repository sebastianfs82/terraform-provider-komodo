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

var _ datasource.DataSource = &ProcedureDataSource{}

func NewProcedureDataSource() datasource.DataSource {
	return &ProcedureDataSource{}
}

type ProcedureDataSource struct {
	client *client.Client
}

type ProcedureDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Stages           types.String `tfsdk:"stages"`
	ScheduleFormat   types.String `tfsdk:"schedule_format"`
	Schedule         types.String `tfsdk:"schedule"`
	ScheduleEnabled  types.Bool   `tfsdk:"schedule_enabled"`
	ScheduleTimezone types.String `tfsdk:"schedule_timezone"`
	ScheduleAlert    types.Bool   `tfsdk:"schedule_alert"`
	FailureAlert     types.Bool   `tfsdk:"failure_alert"`
	WebhookEnabled   types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret    types.String `tfsdk:"webhook_secret"`
}

func (d *ProcedureDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (d *ProcedureDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo procedure resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The procedure identifier (ObjectId or name).",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique name of the procedure.",
			},
			"stages": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON array of procedure stages.",
			},
			"schedule_format": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The schedule format: `Cron` or `English`.",
			},
			"schedule": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The schedule expression.",
			},
			"schedule_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the schedule is enabled.",
			},
			"schedule_timezone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timezone for the schedule (IANA TZ identifier).",
			},
			"schedule_alert": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether an alert is sent on scheduled runs.",
			},
			"failure_alert": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether an alert is sent on procedure failure.",
			},
			"webhook_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether webhook triggering is enabled.",
			},
			"webhook_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The webhook secret override for this procedure.",
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

func (d *ProcedureDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProcedureDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading procedure", map[string]interface{}{"id": data.ID.ValueString()})
	proc, err := d.client.GetProcedure(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read procedure, got error: %s", err))
		return
	}
	if proc == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Procedure %q not found.", data.ID.ValueString()))
		return
	}
	data.Name = types.StringValue(proc.Name)

	stagesStr := string(proc.Config.Stages)
	if len(proc.Config.Stages) > 0 && stagesStr != "null" {
		data.Stages = types.StringValue(stagesStr)
	} else {
		data.Stages = types.StringNull()
	}

	data.ScheduleFormat = types.StringValue(proc.Config.ScheduleFormat)
	data.Schedule = types.StringValue(proc.Config.Schedule)
	data.ScheduleEnabled = types.BoolValue(proc.Config.ScheduleEnabled)
	data.ScheduleTimezone = types.StringValue(proc.Config.ScheduleTimezone)
	data.ScheduleAlert = types.BoolValue(proc.Config.ScheduleAlert)
	data.FailureAlert = types.BoolValue(proc.Config.FailureAlert)
	data.WebhookEnabled = types.BoolValue(proc.Config.WebhookEnabled)
	data.WebhookSecret = types.StringValue(proc.Config.WebhookSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
