// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &UserGroupsDataSource{}

func NewUserGroupsDataSource() datasource.DataSource {
	return &UserGroupsDataSource{}
}

type UserGroupsDataSource struct {
	client *client.Client
}

type UserGroupsDataSourceModel struct {
	Groups []UserGroupItem `tfsdk:"groups"`
}

type UserGroupItem struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	EveryoneEnabled types.Bool   `tfsdk:"everyone_enabled"`
	Users           types.List   `tfsdk:"users"`
	All             types.Map    `tfsdk:"all"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (d *UserGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (d *UserGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo user groups visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of user groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user group ID (Mongo OID).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user group name.",
						},
						"everyone_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this is the 'everyone' group.",
						},
						"users": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of user IDs in the group.",
						},
						"all": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "All permissions or metadata associated with the group.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Last update timestamp in RFC3339 format.",
						},
					},
				},
			},
		},
	}
}

func (d *UserGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing user groups")

	groups, err := d.client.ListUserGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list user groups, got error: %s", err))
		return
	}

	items := make([]UserGroupItem, 0, len(groups))
	for _, g := range groups {
		userIDs := g.Users
		if userIDs == nil {
			userIDs = []string{}
		}
		users, diags := types.ListValueFrom(ctx, types.StringType, userIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		allMap := map[string]attr.Value{}
		for k, v := range g.All {
			allMap[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		all := types.MapValueMust(types.StringType, allMap)

		items = append(items, UserGroupItem{
			ID:              types.StringValue(g.ID.OID),
			Name:            types.StringValue(g.Name),
			EveryoneEnabled: types.BoolValue(g.Everyone),
			Users:           users,
			All:             all,
			UpdatedAt:       types.StringValue(msToRFC3339(g.UpdatedAt)),
		})
	}

	data.Groups = items
	tflog.Trace(ctx, "Listed user groups", map[string]interface{}{"count": len(items)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
