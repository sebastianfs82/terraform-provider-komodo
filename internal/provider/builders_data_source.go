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

	"terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &BuildersDataSource{}

func NewBuildersDataSource() datasource.DataSource {
	return &BuildersDataSource{}
}

type BuildersDataSource struct {
	client *client.Client
}

type BuildersDataSourceModel struct {
	Builders []BuilderListItemModel `tfsdk:"builders"`
}

type BuilderListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	BuilderType types.String `tfsdk:"builder_type"`
}

func (d *BuildersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_builders"
}

func (d *BuildersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo builders visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"builders": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of builders.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The builder identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the builder.",
						},
						"builder_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The builder type (Server, Aws, or Url).",
						},
					},
				},
			},
		},
	}
}

func (d *BuildersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BuildersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BuildersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing builders")

	builders, err := d.client.ListBuilders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list builders, got error: %s", err))
		return
	}

	items := make([]BuilderListItemModel, 0, len(builders))
	for _, b := range builders {
		items = append(items, BuilderListItemModel{
			ID:          types.StringValue(b.ID.OID),
			Name:        types.StringValue(b.Name),
			BuilderType: types.StringValue(b.Config.Type),
		})
	}
	data.Builders = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
