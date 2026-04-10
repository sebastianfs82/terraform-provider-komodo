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

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &RegistryAccountsDataSource{}

func NewRegistryAccountsDataSource() datasource.DataSource {
	return &RegistryAccountsDataSource{}
}

type RegistryAccountsDataSource struct {
	client *client.Client
}

type RegistryAccountsDataSourceModel struct {
	RegistryAccounts []RegistryAccountListItemModel `tfsdk:"registry_accounts"`
}

type RegistryAccountListItemModel struct {
	ID       types.String `tfsdk:"id"`
	Domain   types.String `tfsdk:"domain"`
	Username types.String `tfsdk:"username"`
}

func (d *RegistryAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_accounts"
}

func (d *RegistryAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo Docker registry accounts visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"registry_accounts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of Docker registry accounts.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The registry account identifier (ObjectId).",
						},
						"domain": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The registry domain (e.g. `docker.io`).",
						},
						"username": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The registry account username.",
						},
					},
				},
			},
		},
	}
}

func (d *RegistryAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RegistryAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegistryAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing Docker registry accounts")

	accounts, err := d.client.ListDockerRegistryAccounts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list Docker registry accounts, got error: %s", err))
		return
	}

	items := make([]RegistryAccountListItemModel, 0, len(accounts))
	for _, a := range accounts {
		items = append(items, RegistryAccountListItemModel{
			ID:       types.StringValue(a.ID.OID),
			Domain:   types.StringValue(a.Domain),
			Username: types.StringValue(a.Username),
		})
	}
	data.RegistryAccounts = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
