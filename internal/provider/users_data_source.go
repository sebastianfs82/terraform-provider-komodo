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

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *client.Client
}

type UsersDataSourceModel struct {
	Users []UserDataSourceModel `tfsdk:"users"`
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo users visible to the authenticated admin.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user identifier (ObjectId).",
						},
						"username": schema.StringAttribute{
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
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing users")

	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list users, got error: %s", err))
		return
	}

	items := make([]UserDataSourceModel, 0, len(users))
	for _, u := range users {
		items = append(items, UserDataSourceModel{
			ID:            types.StringValue(u.ID.OID),
			Username:      types.StringValue(u.Username),
			Enabled:       types.BoolValue(u.Enabled),
			Admin:         types.BoolValue(u.Admin),
			CreateServers: types.BoolValue(u.CreateServers),
			CreateBuilds:  types.BoolValue(u.CreateBuilds),
		})
	}
	data.Users = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
