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

var _ datasource.DataSource = &ProviderAccountDataSource{}

func NewProviderAccountDataSource() datasource.DataSource {
	return &ProviderAccountDataSource{}
}

type ProviderAccountDataSource struct {
	client *client.Client
}

type ProviderAccountDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Domain       types.String `tfsdk:"domain"`
	HttpsEnabled types.Bool   `tfsdk:"https_enabled"`
	Username     types.String `tfsdk:"username"`
	Token        types.String `tfsdk:"token"`
}

func (d *ProviderAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_account"
}

func (d *ProviderAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo git provider account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The git provider account identifier (ObjectId).",
			},
			"domain": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git provider domain without protocol prefix (e.g. `github.com`).",
			},
			"https_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether HTTPS is used for cloning.",
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The account username.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The plaintext access token for the account.",
			},
		},
	}
}

func (d *ProviderAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProviderAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading git provider account data source", map[string]interface{}{"id": data.ID.ValueString()})

	account, err := d.client.GetGitProviderAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read git provider account, got error: %s", err))
		return
	}
	if account == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Git provider account with ID %q not found.", data.ID.ValueString()))
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.HttpsEnabled = types.BoolValue(account.HttpsEnabled)
	data.Username = types.StringValue(account.Username)
	data.Token = types.StringValue(account.Token)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
