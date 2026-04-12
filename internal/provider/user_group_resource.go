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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	everyoneSet := !data.EveryoneEnabled.IsNull() && !data.EveryoneEnabled.IsUnknown() && data.EveryoneEnabled.ValueBool()
	usersSet := !data.Users.IsNull() && !data.Users.IsUnknown()
	if everyoneSet && usersSet {
		resp.Diagnostics.AddAttributeError(
			path.Root("users"),
			"Conflicting configuration",
			"\"everyone_enabled = true\" grants access to all users automatically. Remove the \"users\" block or set \"everyone_enabled = false\".",
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
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	EveryoneEnabled types.Bool   `tfsdk:"everyone_enabled"`
	Users           types.List   `tfsdk:"users"`
	All             types.Map    `tfsdk:"all"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
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
	// Only populate the users list if it was explicitly configured (non-null in plan).
	// When null, leave it null so manually-added users are not tracked or removed.
	usersManaged := !data.Users.IsNull() && !data.Users.IsUnknown()
	var users []string
	if usersManaged {
		data.Users.ElementsAs(ctx, &users, false)
	}
	var all map[string]interface{}
	if !data.All.IsNull() && !data.All.IsUnknown() {
		data.All.ElementsAs(ctx, &all, false)
	}
	createReq := client.CreateUserGroupRequest{
		Name:     data.Name.ValueString(),
		Everyone: data.EveryoneEnabled.ValueBool(),
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
	data.EveryoneEnabled = types.BoolValue(fetched.Everyone)
	if !usersManaged {
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
	data.UpdatedAt = types.StringValue(msToRFC3339(fetched.UpdatedAt))
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
	data.EveryoneEnabled = types.BoolValue(group.Everyone)
	// Only refresh users from API if the attribute is managed (non-null in state).
	// Keeping it null prevents drift detection for users added outside of Terraform.
	if !data.Users.IsNull() {
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
	data.UpdatedAt = types.StringValue(msToRFC3339(group.UpdatedAt))
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
	if oldData.EveryoneEnabled.ValueBool() != data.EveryoneEnabled.ValueBool() {
		setEveryoneReq := client.SetEveryoneUserGroupRequest{
			UserGroup: groupID,
			Everyone:  data.EveryoneEnabled.ValueBool(),
		}
		_, err := r.client.SetEveryoneUserGroup(ctx, setEveryoneReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set everyone on user group, got error: %s", err))
			return
		}
	}

	// Handle user list changes.
	// Three cases:
	//   1. users is non-null in plan → manage the full list (add/remove to match exactly).
	//   2. users transitions from non-null (state) to null (plan) → one-time removal of all
	//      previously-managed users, then leave the list unmanaged going forward.
	//   3. users is null in both plan and state → skip entirely (unmanaged, no-op).
	if !data.Users.IsNull() {
		// Case 1: actively managed list.
		var oldUsers, newUsers []string
		if !oldData.Users.IsNull() && !oldData.Users.IsUnknown() {
			resp.Diagnostics.Append(oldData.Users.ElementsAs(ctx, &oldUsers, false)...)
		}
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &newUsers, false)...)
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
	} else if !oldData.Users.IsNull() {
		// Case 2: users was managed before (non-null in state) but is now unset (null in plan).
		// Remove all previously-managed users once, then stop managing the list.
		var oldUsers []string
		resp.Diagnostics.Append(oldData.Users.ElementsAs(ctx, &oldUsers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, u := range oldUsers {
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
	data.EveryoneEnabled = types.BoolValue(group.Everyone)
	// Only refresh users from API if the attribute is managed (non-null in plan).
	if !data.Users.IsNull() {
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
	data.UpdatedAt = types.StringValue(msToRFC3339(group.UpdatedAt))
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
			"everyone_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this is the 'everyone' group.",
				Default:             booldefault.StaticBool(false),
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
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp in RFC3339 format.",
			},
		},
	}
}
