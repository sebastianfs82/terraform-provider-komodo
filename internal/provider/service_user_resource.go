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

var _ resource.Resource = &ServiceUserResource{}
var _ resource.ResourceWithImportState = &ServiceUserResource{}
var _ resource.ResourceWithValidateConfig = &ServiceUserResource{}

func NewServiceUserResource() resource.Resource {
	return &ServiceUserResource{}
}

type ServiceUserResource struct {
	client *client.Client
}

type ServiceUserResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Username            types.String `tfsdk:"username"`
	Description         types.String `tfsdk:"description"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	AdminEnabled        types.Bool   `tfsdk:"admin_enabled"`
	CreateServerEnabled types.Bool   `tfsdk:"create_server_enabled"`
	CreateBuildEnabled  types.Bool   `tfsdk:"create_build_enabled"`
}

func (r *ServiceUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (r *ServiceUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo service user. Service users authenticate via API keys only and have no password.",
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
				MarkdownDescription: "The globally unique username for the service user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A description for the service user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the service user is enabled and able to access the API.",
			},
			"admin_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the service user has global admin permissions.",
			},
			"create_server_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the service user can create servers. Cannot be set when admin_enabled is true.",
			},
			"create_build_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the service user can create builds. Cannot be set when admin_enabled is true.",
			},
		},
	}
}

func (r *ServiceUserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ServiceUserResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AdminEnabled.ValueBool() {
		if data.CreateServerEnabled.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("create_server_enabled"),
				"Invalid Configuration",
				"create_server_enabled cannot be set to true alongside admin_enabled = true. Admins implicitly have all permissions.",
			)
		}
		if data.CreateBuildEnabled.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("create_build_enabled"),
				"Invalid Configuration",
				"create_build_enabled cannot be set to true alongside admin_enabled = true. Admins implicitly have all permissions.",
			)
		}
	}
}

func (r *ServiceUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := data.Description.ValueString()
	_, err := r.client.CreateServiceUser(ctx, client.CreateServiceUserRequest{
		Username:    data.Username.ValueString(),
		Description: description,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service user, got error: %s", err))
		return
	}

	user, err := r.client.FindUser(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find service user after creation, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Service user %q not found after creation", data.Username.ValueString()))
		return
	}

	if !data.AdminEnabled.IsNull() && !data.AdminEnabled.IsUnknown() && data.AdminEnabled.ValueBool() {
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
	if !data.CreateServerEnabled.IsNull() && !data.CreateServerEnabled.IsUnknown() {
		v := data.CreateServerEnabled.ValueBool()
		permReq.CreateServers = &v
		needsPerms = true
	}
	if !data.CreateBuildEnabled.IsNull() && !data.CreateBuildEnabled.IsUnknown() {
		v := data.CreateBuildEnabled.ValueBool()
		permReq.CreateBuilds = &v
		needsPerms = true
	}
	if needsPerms {
		if err := r.client.UpdateUserBasePermissions(ctx, permReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set service user permissions, got error: %s", err))
			return
		}
	}

	user, err = r.client.FindUser(ctx, user.ID.OID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service user after creation, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Service user %q not found after creation", data.Username.ValueString()))
		return
	}

	data.ID = types.StringValue(user.ID.OID)
	data.Username = types.StringValue(user.Username)
	data.Enabled = types.BoolValue(user.Enabled)
	data.AdminEnabled = types.BoolValue(user.Admin)
	data.CreateServerEnabled = types.BoolValue(user.CreateServers)
	data.CreateBuildEnabled = types.BoolValue(user.CreateBuilds)
	// description is not returned by the API; ensure it is always a known value
	if data.Description.IsNull() || data.Description.IsUnknown() {
		data.Description = types.StringValue("")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceUserResourceModel
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service user, got error: %s", err))
		return
	}
	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(user.ID.OID)
	data.Username = types.StringValue(user.Username)
	data.Enabled = types.BoolValue(user.Enabled)
	data.AdminEnabled = types.BoolValue(user.Admin)
	data.CreateServerEnabled = types.BoolValue(user.CreateServers)
	data.CreateBuildEnabled = types.BoolValue(user.CreateBuilds)
	// description is not returned by the API; preserve whatever is in state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ServiceUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		if err := r.client.UpdateServiceUserDescription(ctx, client.UpdateServiceUserDescriptionRequest{
			Username:    state.Username.ValueString(),
			Description: plan.Description.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service user description, got error: %s", err))
			return
		}
	}

	if !plan.AdminEnabled.Equal(state.AdminEnabled) {
		err := r.client.UpdateUserAdmin(ctx, client.UpdateUserAdminRequest{
			UserID: state.ID.ValueString(),
			Admin:  plan.AdminEnabled.ValueBool(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service user admin status, got error: %s", err))
			return
		}
	}

	if !plan.Enabled.Equal(state.Enabled) || !plan.CreateServerEnabled.Equal(state.CreateServerEnabled) || !plan.CreateBuildEnabled.Equal(state.CreateBuildEnabled) {
		enabledVal := plan.Enabled.ValueBool()
		createServersVal := plan.CreateServerEnabled.ValueBool()
		createBuildsVal := plan.CreateBuildEnabled.ValueBool()
		err := r.client.UpdateUserBasePermissions(ctx, client.UpdateUserBasePermissionsRequest{
			UserID:        state.ID.ValueString(),
			Enabled:       &enabledVal,
			CreateServers: &createServersVal,
			CreateBuilds:  &createBuildsVal,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service user permissions, got error: %s", err))
			return
		}
	}

	user, err := r.client.FindUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service user after update, got error: %s", err))
		return
	}
	if user == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Service user %q not found after update", state.ID.ValueString()))
		return
	}

	plan.ID = types.StringValue(user.ID.OID)
	plan.Username = types.StringValue(user.Username)
	plan.Enabled = types.BoolValue(user.Enabled)
	plan.AdminEnabled = types.BoolValue(user.Admin)
	plan.CreateServerEnabled = types.BoolValue(user.CreateServers)
	plan.CreateBuildEnabled = types.BoolValue(user.CreateBuilds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service user, got error: %s", err))
		return
	}
}

func (r *ServiceUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
