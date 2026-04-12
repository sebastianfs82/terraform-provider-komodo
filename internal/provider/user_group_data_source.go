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
)

var _ datasource.DataSource = &UserGroupDataSource{}
var _ datasource.DataSourceWithValidateConfig = &UserGroupDataSource{}

type UserGroupDataSource struct {
	client *client.Client
}

func NewUserGroupDataSource() datasource.DataSource {
	return &UserGroupDataSource{}
}

func (d *UserGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserGroupDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data UserGroupDataSourceModel
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

func (d *UserGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	group, err := d.client.GetUserGroup(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("User group %q not found: %s", lookup, err))
		return
	}
	data.ID = types.StringValue(group.ID.OID)
	data.Name = types.StringValue(group.Name)
	data.EveryoneEnabled = types.BoolValue(group.Everyone)
	userIDs := group.Users
	if userIDs == nil {
		userIDs = []string{}
	}
	data.Users, _ = types.ListValueFrom(ctx, types.StringType, userIDs)
	allMap := map[string]attr.Value{}
	for k, v := range group.All {
		allMap[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	data.All = types.MapValueMust(types.StringType, allMap)
	data.UpdatedAt = types.StringValue(msToRFC3339(group.UpdatedAt))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type UserGroupDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	EveryoneEnabled types.Bool   `tfsdk:"everyone_enabled"`
	Users           types.List   `tfsdk:"users"`
	All             types.Map    `tfsdk:"all"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (d *UserGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (d *UserGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo user group by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user group ID (Mongo OID). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user group name. One of `name` or `id` must be set.",
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
				MarkdownDescription: "All permissions or metadata.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp in RFC3339 format.",
			},
		},
	}
}

// ...Read method to be implemented...
