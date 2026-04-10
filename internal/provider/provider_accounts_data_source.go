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

var _ datasource.DataSource = &ProviderAccountsDataSource{}

func NewProviderAccountsDataSource() datasource.DataSource {
	return &ProviderAccountsDataSource{}
}

type ProviderAccountsDataSource struct {
	client *client.Client
}

type ProviderAccountsDataSourceModel struct {
	ProviderAccounts []ProviderAccountListItemModel `tfsdk:"provider_accounts"`
}

type ProviderAccountListItemModel struct {
	ID       types.String `tfsdk:"id"`
	Domain   types.String `tfsdk:"domain"`
	Username types.String `tfsdk:"username"`
	Https    types.Bool   `tfsdk:"https"`
}

func (d *ProviderAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_accounts"
}

func (d *ProviderAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo git provider accounts visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"provider_accounts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of git provider accounts.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git provider account identifier (ObjectId).",
						},
						"domain": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git provider domain (e.g. `github.com`).",
						},
						"username": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git account username.",
						},
						"https": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether HTTPS is used for this git provider.",
						},
					},
				},
			},
		},
	}
}

func (d *ProviderAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProviderAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing git provider accounts")

	accounts, err := d.client.ListGitProviderAccounts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list git provider accounts, got error: %s", err))
		return
	}

	items := make([]ProviderAccountListItemModel, 0, len(accounts))
	for _, a := range accounts {
		items = append(items, ProviderAccountListItemModel{
			ID:       types.StringValue(a.ID.OID),
			Domain:   types.StringValue(a.Domain),
			Username: types.StringValue(a.Username),
			Https:    types.BoolValue(a.HttpsEnabled),
		})
	}
	data.ProviderAccounts = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
