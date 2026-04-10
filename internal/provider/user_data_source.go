// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *client.Client
}

type UserDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Admin         types.Bool   `tfsdk:"admin"`
	CreateServers types.Bool   `tfsdk:"create_servers"`
	CreateBuilds  types.Bool   `tfsdk:"create_builds"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo user by username or id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user id (ObjectId). If set, takes precedence over username.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The globally unique username.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user is enabled and able to access the API.",
			},
			"admin": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user has global admin permissions.",
			},
			"create_servers": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user can create servers.",
			},
			"create_builds": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user can create builds.",
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lookup := data.ID.ValueString()
	if lookup == "" {
		lookup = data.Username.ValueString()
	}
	if lookup == "" {
		resp.Diagnostics.AddError("Missing Query Attribute", "Either id or username must be set to query a user.")
		return
	}

	user, err := d.client.FindUser(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("User %q not found", lookup))
		return
	}

	data.ID = types.StringValue(user.ID.OID)
	data.Username = types.StringValue(user.Username)
	data.Enabled = types.BoolValue(user.Enabled)
	data.Admin = types.BoolValue(user.Admin)
	data.CreateServers = types.BoolValue(user.CreateServers)
	data.CreateBuilds = types.BoolValue(user.CreateBuilds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
