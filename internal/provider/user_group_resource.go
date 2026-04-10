// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.ResourceWithValidateConfig = &UserGroupResource{}

func (r *UserGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data UserGroupResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	everyoneSet := !data.Everyone.IsNull() && !data.Everyone.IsUnknown() && data.Everyone.ValueBool()
	usersSet := !data.Users.IsNull() && !data.Users.IsUnknown() && len(data.Users.Elements()) > 0
	if everyoneSet && usersSet {
		resp.Diagnostics.AddAttributeError(
			path.Root("users"),
			"Conflicting configuration",
			"\"everyone\" and \"users\" are mutually exclusive. Set only one.",
		)
	}
}

func (r *UserGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

type UserGroupResource struct {
	client *client.Client
}

type UserGroupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Everyone  types.Bool   `tfsdk:"everyone"`
	Users     types.List   `tfsdk:"users"`
	All       types.Map    `tfsdk:"all"`
	UpdatedAt types.Int64  `tfsdk:"updated_at"`
}

func NewUserGroupResource() resource.Resource {
	return &UserGroupResource{}
}

func (r *UserGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *UserGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var users []string
	if !data.Users.IsNull() && !data.Users.IsUnknown() {
		data.Users.ElementsAs(ctx, &users, false)
	}
	var all map[string]interface{}
	if !data.All.IsNull() && !data.All.IsUnknown() {
		data.All.ElementsAs(ctx, &all, false)
	}
	createReq := client.CreateUserGroupRequest{
		Name:     data.Name.ValueString(),
		Everyone: data.Everyone.ValueBool(),
		Users:    users,
		All:      all,
	}
	group, err := r.client.CreateUserGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user group, got error: %s", err))
		return
	}

	groupID := group.ID.OID

	// Add users individually — CreateUserGroup does not populate the user list
	for _, u := range users {
		_, err := r.client.AddUserToUserGroup(ctx, client.AddUserToUserGroupRequest{UserGroup: groupID, User: u})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add user %s to group, got error: %s", u, err))
			return
		}
	}

	// Refresh from backend to get authoritative state
	fetched, err := r.client.GetUserGroup(ctx, groupID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch user group after create, got error: %s", err))
		return
	}
	data.ID = types.StringValue(fetched.ID.OID)
	data.Name = types.StringValue(fetched.Name)
	data.Everyone = types.BoolValue(fetched.Everyone)
	if len(fetched.Users) == 0 {
		data.Users = types.ListNull(types.StringType)
	} else {
		data.Users, _ = types.ListValueFrom(ctx, types.StringType, fetched.Users)
	}
	if len(fetched.All) == 0 {
		data.All = types.MapNull(types.StringType)
	} else {
		allMap := map[string]attr.Value{}
		for k, v := range fetched.All {
			allMap[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.All = types.MapValueMust(types.StringType, allMap)
	}
	data.UpdatedAt = types.Int64Value(fetched.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	userGroupKey := data.ID.ValueString()
	if userGroupKey == "" {
		userGroupKey = data.Name.ValueString()
	}
	group, err := r.client.GetUserGroup(ctx, userGroupKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no UserGroup found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user group, got error: %s", err))
		return
	}
	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.ID = types.StringValue(group.ID.OID)
	data.Name = types.StringValue(group.Name)
	data.Everyone = types.BoolValue(group.Everyone)
	if len(group.Users) == 0 {
		data.Users = types.ListNull(types.StringType)
	} else {
		data.Users, _ = types.ListValueFrom(ctx, types.StringType, group.Users)
	}
	if len(group.All) == 0 {
		data.All = types.MapNull(types.StringType)
	} else {
		allMap := map[string]attr.Value{}
		for k, v := range group.All {
			allMap[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.All = types.MapValueMust(types.StringType, allMap)
	}
	data.UpdatedAt = types.Int64Value(group.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var oldData UserGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := oldData.ID.ValueString()

	// Handle rename
	if oldData.Name.ValueString() != "" && oldData.Name.ValueString() != data.Name.ValueString() {
		_, err := r.client.RenameUserGroup(ctx, groupID, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename user group, got error: %s", err))
			return
		}
	}

	// Handle everyone attribute change
	if oldData.Everyone.ValueBool() != data.Everyone.ValueBool() {
		setEveryoneReq := client.SetEveryoneUserGroupRequest{
			UserGroup: groupID,
			Everyone:  data.Everyone.ValueBool(),
		}
		_, err := r.client.SetEveryoneUserGroup(ctx, setEveryoneReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set everyone on user group, got error: %s", err))
			return
		}
	}

	// Handle user list changes
	var oldUsers, newUsers []string
	if !oldData.Users.IsNull() && !oldData.Users.IsUnknown() {
		resp.Diagnostics.Append(oldData.Users.ElementsAs(ctx, &oldUsers, false)...)
	}
	if !data.Users.IsNull() && !data.Users.IsUnknown() {
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &newUsers, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	oldSet := make(map[string]bool, len(oldUsers))
	for _, u := range oldUsers {
		oldSet[u] = true
	}
	newSet := make(map[string]bool, len(newUsers))
	for _, u := range newUsers {
		newSet[u] = true
	}

	for _, u := range newUsers {
		if !oldSet[u] {
			_, err := r.client.AddUserToUserGroup(ctx, client.AddUserToUserGroupRequest{UserGroup: groupID, User: u})
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add user %s to group, got error: %s", u, err))
				return
			}
		}
	}
	for _, u := range oldUsers {
		if !newSet[u] {
			_, err := r.client.RemoveUserFromUserGroup(ctx, client.RemoveUserFromUserGroupRequest{UserGroup: groupID, User: u})
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove user %s from group, got error: %s", u, err))
				return
			}
		}
	}

	// Read final state from API
	group, err := r.client.GetUserGroup(ctx, groupID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user group after update, got error: %s", err))
		return
	}
	data.ID = types.StringValue(group.ID.OID)
	data.Name = types.StringValue(group.Name)
	data.Everyone = types.BoolValue(group.Everyone)
	if len(group.Users) == 0 {
		data.Users = types.ListNull(types.StringType)
	} else {
		data.Users, _ = types.ListValueFrom(ctx, types.StringType, group.Users)
	}
	if len(group.All) == 0 {
		data.All = types.MapNull(types.StringType)
	} else {
		allMap := map[string]attr.Value{}
		for k, v := range group.All {
			allMap[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		data.All = types.MapValueMust(types.StringType, allMap)
	}
	data.UpdatedAt = types.Int64Value(group.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteUserGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user group, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted user group resource")
}

func (r *UserGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo user group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user group ID (Mongo OID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user group name.",
			},
			"everyone": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this is the 'everyone' group.",
			},
			"users": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of user IDs in the group.",
			},
			"all": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "All permissions or metadata.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp.",
			},
		},
	}
}

// ...CRUD methods to be implemented...
