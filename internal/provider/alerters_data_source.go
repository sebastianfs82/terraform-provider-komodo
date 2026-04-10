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

var _ datasource.DataSource = &AlertersDataSource{}

func NewAlertersDataSource() datasource.DataSource {
	return &AlertersDataSource{}
}

type AlertersDataSource struct {
	client *client.Client
}

type AlertersDataSourceModel struct {
	Alerters []AlerterListItemModel `tfsdk:"alerters"`
}

type AlerterListItemModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	EndpointType types.String `tfsdk:"endpoint_type"`
}

func (d *AlertersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerters"
}

func (d *AlertersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo alerters visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"alerters": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of alerters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The alerter identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the alerter.",
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the alerter is enabled.",
						},
						"endpoint_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The alerter endpoint type (Slack, Discord, Custom, etc.).",
						},
					},
				},
			},
		},
	}
}

func (d *AlertersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AlertersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AlertersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing alerters")

	alerters, err := d.client.ListAlerters(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list alerters, got error: %s", err))
		return
	}

	items := make([]AlerterListItemModel, 0, len(alerters))
	for _, a := range alerters {
		items = append(items, AlerterListItemModel{
			ID:           types.StringValue(a.ID.OID),
			Name:         types.StringValue(a.Name),
			Enabled:      types.BoolValue(a.Config.Enabled),
			EndpointType: types.StringValue(a.Config.Endpoint.Type),
		})
	}
	data.Alerters = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
