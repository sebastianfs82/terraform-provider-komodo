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

var _ datasource.DataSource = &ActionDataSource{}
var _ datasource.DataSourceWithValidateConfig = &ActionDataSource{}

func NewActionDataSource() datasource.DataSource {
	return &ActionDataSource{}
}

type ActionDataSource struct {
	client *client.Client
}

type ActionDataSourceModel struct {
	ID                        types.String    `tfsdk:"id"`
	Name                      types.String    `tfsdk:"name"`
	Tags                      types.List      `tfsdk:"tags"`
	RunOnStartupEnabled       types.Bool      `tfsdk:"run_on_startup_enabled"`
	Schedule                  *ScheduleModel  `tfsdk:"schedule"`
	FailureAlert              types.Bool      `tfsdk:"failure_alert_enabled"`
	Webhook                   *WebhookModel   `tfsdk:"webhook"`
	ReloadDependenciesEnabled types.Bool      `tfsdk:"reload_dependencies_enabled"`
	FileContents              types.String    `tfsdk:"file_contents"`
	Arguments                 []ArgumentModel `tfsdk:"argument"`
}

func (d *ActionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (d *ActionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo action resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The action identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique name of the action. One of `name` or `id` must be set.",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The list of tag IDs attached to this action.",
			},
			"run_on_startup_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the action runs at Komodo startup.",
			},
			"schedule": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Schedule configuration for the action.",
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
				MarkdownDescription: "Whether an alert is sent on action failure.",
			},
			"webhook": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook configuration for the action.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether webhook triggering is enabled.",
					},
					"secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "The webhook secret override for this action.",
					},
				},
			},
			"reload_dependencies_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether Deno dependencies are reloaded on each run.",
			},
			"file_contents": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "TypeScript file contents using the Komodo client.",
			},
		},
		Blocks: map[string]schema.Block{
			"argument": schema.ListNestedBlock{
				MarkdownDescription: "Key-value arguments passed to the action as the `ARGS` variable.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The argument name (environment variable name).",
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The argument value.",
						},
					},
				},
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

func (d *ActionDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data ActionDataSourceModel
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

func (d *ActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading action", map[string]interface{}{"lookup": lookup})
	a, err := d.client.GetAction(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read action, got error: %s", err))
		return
	}
	if a == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Action %q not found.", lookup))
		return
	}
	data.ID = types.StringValue(a.ID.OID)
	data.Name = types.StringValue(a.Name)
	tagVals := make([]attr.Value, len(a.Tags))
	for i, t := range a.Tags {
		tagVals[i] = types.StringValue(t)
	}
	data.Tags = types.ListValueMust(types.StringType, tagVals)
	data.RunOnStartupEnabled = types.BoolValue(a.Config.RunAtStartup)
	if a.Config.ScheduleEnabled || a.Config.Schedule != "" || a.Config.ScheduleTimezone != "" || a.Config.ScheduleAlert {
		data.Schedule = &ScheduleModel{
			Format:       types.StringValue(a.Config.ScheduleFormat),
			Expression:   types.StringValue(a.Config.Schedule),
			Enabled:      types.BoolValue(a.Config.ScheduleEnabled),
			Timezone:     types.StringValue(a.Config.ScheduleTimezone),
			AlertEnabled: types.BoolValue(a.Config.ScheduleAlert),
		}
	} else {
		data.Schedule = nil
	}
	data.FailureAlert = types.BoolValue(a.Config.FailureAlert)
	webhookSecret := types.StringNull()
	if a.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(a.Config.WebhookSecret)
	}
	if a.Config.WebhookEnabled || a.Config.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(a.Config.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
	data.ReloadDependenciesEnabled = types.BoolValue(a.Config.ReloadDenoDeps)
	data.FileContents = types.StringValue(strings.TrimRight(a.Config.FileContents, "\n"))
	data.Arguments = parseActionArguments(a.Config.ArgumentsFormat, a.Config.Arguments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
