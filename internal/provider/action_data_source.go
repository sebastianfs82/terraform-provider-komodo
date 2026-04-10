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

var _ datasource.DataSource = &ActionDataSource{}

func NewActionDataSource() datasource.DataSource {
	return &ActionDataSource{}
}

type ActionDataSource struct {
	client *client.Client
}

type ActionDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	RunAtStartup     types.Bool   `tfsdk:"run_at_startup"`
	ScheduleFormat   types.String `tfsdk:"schedule_format"`
	Schedule         types.String `tfsdk:"schedule"`
	ScheduleEnabled  types.Bool   `tfsdk:"schedule_enabled"`
	ScheduleTimezone types.String `tfsdk:"schedule_timezone"`
	ScheduleAlert    types.Bool   `tfsdk:"schedule_alert"`
	FailureAlert     types.Bool   `tfsdk:"failure_alert"`
	WebhookEnabled   types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret    types.String `tfsdk:"webhook_secret"`
	ReloadDenoDeps   types.Bool   `tfsdk:"reload_deno_deps"`
	FileContents     types.String `tfsdk:"file_contents"`
	ArgumentsFormat  types.String `tfsdk:"arguments_format"`
	Arguments        types.String `tfsdk:"arguments"`
}

func (d *ActionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (d *ActionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo action resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The action identifier (ObjectId or name).",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique name of the action.",
			},
			"run_at_startup": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the action runs at Komodo startup.",
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
				MarkdownDescription: "Whether an alert is sent on action failure.",
			},
			"webhook_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether webhook triggering is enabled.",
			},
			"webhook_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The webhook secret override for this action.",
			},
			"reload_deno_deps": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether Deno dependencies are reloaded on each run.",
			},
			"file_contents": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "TypeScript file contents using the Komodo client.",
			},
			"arguments_format": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The format for action arguments.",
			},
			"arguments": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Default arguments passed to the action as the `ARGS` variable.",
			},
		},
	}
}

func (d *ActionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading action", map[string]interface{}{"id": data.ID.ValueString()})
	a, err := d.client.GetAction(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read action, got error: %s", err))
		return
	}
	if a == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Action %q not found.", data.ID.ValueString()))
		return
	}
	data.Name = types.StringValue(a.Name)
	data.RunAtStartup = types.BoolValue(a.Config.RunAtStartup)
	data.ScheduleFormat = types.StringValue(a.Config.ScheduleFormat)
	data.Schedule = types.StringValue(a.Config.Schedule)
	data.ScheduleEnabled = types.BoolValue(a.Config.ScheduleEnabled)
	data.ScheduleTimezone = types.StringValue(a.Config.ScheduleTimezone)
	data.ScheduleAlert = types.BoolValue(a.Config.ScheduleAlert)
	data.FailureAlert = types.BoolValue(a.Config.FailureAlert)
	data.WebhookEnabled = types.BoolValue(a.Config.WebhookEnabled)
	data.WebhookSecret = types.StringValue(a.Config.WebhookSecret)
	data.ReloadDenoDeps = types.BoolValue(a.Config.ReloadDenoDeps)
	data.FileContents = types.StringValue(strings.TrimRight(a.Config.FileContents, "\n"))
	data.ArgumentsFormat = types.StringValue(a.Config.ArgumentsFormat)
	data.Arguments = types.StringValue(a.Config.Arguments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
