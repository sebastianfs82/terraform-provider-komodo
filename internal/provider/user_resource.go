// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}
var _ resource.ResourceWithValidateConfig = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *client.Client
}

type UserResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Admin         types.Bool   `tfsdk:"admin"`
	CreateServers types.Bool   `tfsdk:"create_servers"`
	CreateBuilds  types.Bool   `tfsdk:"create_builds"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo local user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Terraform resource ID (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The globally unique username for the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The password for the local user. Changes to this field will recreate the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the user is enabled and able to access the API.",
			},
			"admin": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the user has global admin permissions.",
			},
			"create_servers": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the user can create servers. Cannot be set when admin is true.",
			},
			"create_builds": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the user can create builds. Cannot be set when admin is true.",
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Admin.ValueBool() {
		if !data.CreateServers.IsNull() && !data.CreateServers.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("create_servers"),
				"Invalid Configuration",
				"create_servers cannot be set alongside admin = true. Admins implicitly have all permissions.",
			)
		}
		if !data.CreateBuilds.IsNull() && !data.CreateBuilds.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("create_builds"),
				"Invalid Configuration",
				"create_builds cannot be set alongside admin = true. Admins implicitly have all permissions.",
			)
		}
	}
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateLocalUser(ctx, client.CreateLocalUserRequest{
		Username: data.Username.ValueString(),
		Password: data.Password.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	// FindUser by username to get the full record including ID
	user, err := r.client.FindUser(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find user after creation, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("User %q not found after creation", data.Username.ValueString()))
		return
	}

	if !data.Admin.IsNull() && !data.Admin.IsUnknown() && data.Admin.ValueBool() {
		if err := r.client.UpdateUserAdmin(ctx, client.UpdateUserAdminRequest{
			UserID: user.ID.OID,
			Admin:  true,
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set admin status, got error: %s", err))
			return
		}
	}

	permReq := client.UpdateUserBasePermissionsRequest{UserID: user.ID.OID}
	needsPerms := false
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		v := data.Enabled.ValueBool()
		permReq.Enabled = &v
		needsPerms = true
	}
	if !data.CreateServers.IsNull() && !data.CreateServers.IsUnknown() {
		v := data.CreateServers.ValueBool()
		permReq.CreateServers = &v
		needsPerms = true
	}
	if !data.CreateBuilds.IsNull() && !data.CreateBuilds.IsUnknown() {
		v := data.CreateBuilds.ValueBool()
		permReq.CreateBuilds = &v
		needsPerms = true
	}
	if needsPerms {
		if err := r.client.UpdateUserBasePermissions(ctx, permReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set user permissions, got error: %s", err))
			return
		}
	}

	user, err = r.client.FindUser(ctx, user.ID.OID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user after creation, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("User %q not found after creation", data.Username.ValueString()))
		return
	}

	data.ID = types.StringValue(user.ID.OID)
	data.Username = types.StringValue(user.Username)
	data.Enabled = types.BoolValue(user.Enabled)
	data.Admin = types.BoolValue(user.Admin)
	data.CreateServers = types.BoolValue(user.CreateServers)
	data.CreateBuilds = types.BoolValue(user.CreateBuilds)
	// password is not returned by the API; preserve what was configured
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.FindUser(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}
	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(user.ID.OID)
	data.Username = types.StringValue(user.Username)
	data.Enabled = types.BoolValue(user.Enabled)
	data.Admin = types.BoolValue(user.Admin)
	data.CreateServers = types.BoolValue(user.CreateServers)
	data.CreateBuilds = types.BoolValue(user.CreateBuilds)
	// password is not returned by the API; preserve whatever is in state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Admin.Equal(state.Admin) {
		err := r.client.UpdateUserAdmin(ctx, client.UpdateUserAdminRequest{
			UserID: state.ID.ValueString(),
			Admin:  plan.Admin.ValueBool(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user admin status, got error: %s", err))
			return
		}
	}

	if !plan.Enabled.Equal(state.Enabled) || !plan.CreateServers.Equal(state.CreateServers) || !plan.CreateBuilds.Equal(state.CreateBuilds) {
		enabledVal := plan.Enabled.ValueBool()
		createServersVal := plan.CreateServers.ValueBool()
		createBuildsVal := plan.CreateBuilds.ValueBool()
		err := r.client.UpdateUserBasePermissions(ctx, client.UpdateUserBasePermissionsRequest{
			UserID:        state.ID.ValueString(),
			Enabled:       &enabledVal,
			CreateServers: &createServersVal,
			CreateBuilds:  &createBuildsVal,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user permissions, got error: %s", err))
			return
		}
	}

	user, err := r.client.FindUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user after update, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("User %q not found after update", state.ID.ValueString()))
		return
	}

	plan.ID = types.StringValue(user.ID.OID)
	plan.Username = types.StringValue(user.Username)
	plan.Enabled = types.BoolValue(user.Enabled)
	plan.Admin = types.BoolValue(user.Admin)
	plan.CreateServers = types.BoolValue(user.CreateServers)
	plan.CreateBuilds = types.BoolValue(user.CreateBuilds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
