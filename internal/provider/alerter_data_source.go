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

func NewAlerterDataSource() datasource.DataSource {
	return &AlerterDataSource{}
}

type AlerterDataSource struct {
	client *client.Client
}

type AlerterDataSourceModel struct {
	ID               types.String          `tfsdk:"id"`
	Name             types.String          `tfsdk:"name"`
	Enabled          types.Bool            `tfsdk:"enabled"`
	EndpointType     types.String          `tfsdk:"endpoint_type"`
	AlertTypes       types.List            `tfsdk:"alert_types"`
	CustomEndpoint   *AlerterCustomModel   `tfsdk:"custom_endpoint"`
	SlackEndpoint    *AlerterSlackModel    `tfsdk:"slack_endpoint"`
	DiscordEndpoint  *AlerterDiscordModel  `tfsdk:"discord_endpoint"`
	NtfyEndpoint     *AlerterNtfyModel     `tfsdk:"ntfy_endpoint"`
	PushoverEndpoint *AlerterPushoverModel `tfsdk:"pushover_endpoint"`
}

func (d *AlerterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerter"
}

func (d *AlerterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo alerter resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The alerter identifier (ObjectId or name).",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique name of the alerter.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the alerter is enabled.",
			},
			"endpoint_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The alerter endpoint type: `Custom`, `Slack`, `Discord`, `Ntfy`, or `Pushover`.",
			},
			"alert_types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Alert types the alerter is configured to send. Empty means all types.",
			},
			"custom_endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a Custom HTTP endpoint alerter. Populated when `endpoint_type` is `Custom`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The HTTP/S endpoint URL.",
					},
				},
			},
			"slack_endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a Slack alerter. Populated when `endpoint_type` is `Slack`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Slack app webhook URL.",
					},
				},
			},
			"discord_endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a Discord alerter. Populated when `endpoint_type` is `Discord`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Discord webhook URL.",
					},
				},
			},
			"ntfy_endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a Ntfy alerter. Populated when `endpoint_type` is `Ntfy`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Ntfy topic URL.",
					},
					"email": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Optional email address for Ntfy email notifications.",
					},
				},
			},
			"pushover_endpoint": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a Pushover alerter. Populated when `endpoint_type` is `Pushover`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Pushover URL including application and user tokens.",
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

func (d *AlerterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AlerterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading alerter", map[string]interface{}{"id": data.ID.ValueString()})
	a, err := d.client.GetAlerter(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alerter, got error: %s", err))
		return
	}
	if a == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Alerter %q not found.", data.ID.ValueString()))
		return
	}

	rm := AlerterResourceModel{
		ID:   data.ID,
		Name: types.StringValue(a.Name),
	}
	resp.Diagnostics.Append(alerterToModel(ctx, a, &rm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Name = rm.Name
	data.Enabled = rm.Enabled
	data.EndpointType = rm.EndpointType
	data.AlertTypes = rm.AlertTypes
	data.CustomEndpoint = rm.CustomEndpoint
	data.SlackEndpoint = rm.SlackEndpoint
	data.DiscordEndpoint = rm.DiscordEndpoint
	data.NtfyEndpoint = rm.NtfyEndpoint
	data.PushoverEndpoint = rm.PushoverEndpoint

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
