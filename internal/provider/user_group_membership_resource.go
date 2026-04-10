// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserGroupMembershipResource{}
var _ resource.ResourceWithImportState = &UserGroupMembershipResource{}

func NewUserGroupMembershipResource() resource.Resource {
	return &UserGroupMembershipResource{}
}

// UserGroupMembershipResource manages a single user's membership in a user group.
type UserGroupMembershipResource struct {
	client *client.Client
}

// UserGroupMembershipResourceModel describes the resource data model.
type UserGroupMembershipResourceModel struct {
	ID        types.String `tfsdk:"id"`
	UserGroup types.String `tfsdk:"user_group"`
	User      types.String `tfsdk:"user"`
}

func (r *UserGroupMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group_membership"
}

func (r *UserGroupMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a single user's membership in a Komodo user group.\n\n" +
			"~> **Note:** Do not use this resource alongside `komodo_user_group` with a `users` list " +
			"for the same group. Managing the same membership from both resources will cause conflicts.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource ID, formatted as `{user_group}/{user}`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_group": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name or ID of the user group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The username or ID of the user to add to the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *UserGroupMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserGroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserGroupMembershipResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userGroup := data.UserGroup.ValueString()
	user := data.User.ValueString()

	tflog.Debug(ctx, "Adding user to user group", map[string]interface{}{
		"user_group": userGroup,
		"user":       user,
	})

	_, err := r.client.AddUserToUserGroup(ctx, client.AddUserToUserGroupRequest{
		UserGroup: userGroup,
		User:      user,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add user to user group, got error: %s", err))
		return
	}

	data.ID = types.StringValue(userGroup + "/" + user)

	tflog.Trace(ctx, "Created user group membership resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserGroupMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userGroup := data.UserGroup.ValueString()
	user := data.User.ValueString()

	group, err := r.client.GetUserGroup(ctx, userGroup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user group, got error: %s", err))
		return
	}

	found := false
	for _, u := range group.Users {
		if strings.EqualFold(u, user) {
			found = true
			break
		}
	}

	if !found {
		tflog.Debug(ctx, "User no longer in group, removing from state", map[string]interface{}{
			"user_group": userGroup,
			"user":       user,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both attributes are ForceNew; Update is never called.
}

func (r *UserGroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserGroupMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userGroup := data.UserGroup.ValueString()
	user := data.User.ValueString()

	tflog.Debug(ctx, "Removing user from user group", map[string]interface{}{
		"user_group": userGroup,
		"user":       user,
	})

	_, err := r.client.RemoveUserFromUserGroup(ctx, client.RemoveUserFromUserGroupRequest{
		UserGroup: userGroup,
		User:      user,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove user from user group, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted user group membership resource")
}

func (r *UserGroupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format '{user_group}/{user}', got: %q", req.ID),
		)
		return
	}

	data := UserGroupMembershipResourceModel{
		ID:        types.StringValue(req.ID),
		UserGroup: types.StringValue(parts[0]),
		User:      types.StringValue(parts[1]),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
