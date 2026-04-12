// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}
var _ resource.ResourceWithValidateConfig = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

// ApiKeyResource defines the resource implementation.
type ApiKeyResource struct {
	client *client.Client
}

// ApiKeyResourceModel describes the resource data model.
type ApiKeyResourceModel struct {
	Key           types.String `tfsdk:"key"`
	Secret        types.String `tfsdk:"secret"`
	Name          types.String `tfsdk:"name"`
	UserID        types.String `tfsdk:"user_id"`
	ServiceUserID types.String `tfsdk:"service_user_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
}

func (r *ApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo API key.",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The API key identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key secret (only available on creation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A human-friendly name for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who owns this key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "When set, creates the API key for the specified service user instead of the authenticated user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp in RFC3339 format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Expiration time in RFC3339 format (e.g. `2030-01-01T00:00:00Z`). Use `\"\"` (empty string) for no expiration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ApiKeyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ApiKeyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.ExpiresAt.IsNull() || data.ExpiresAt.IsUnknown() || data.ExpiresAt.ValueString() == "" {
		return
	}
	t, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("expires_at"),
			"Invalid RFC3339 Timestamp",
			fmt.Sprintf(`"expires_at" must be a valid RFC3339 timestamp (e.g. "2030-01-01T00:00:00Z") or "" for no expiration. Got: %q.`, data.ExpiresAt.ValueString()),
		)
		return
	}
	if !t.After(time.Now()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("expires_at"),
			"Invalid Expiration Timestamp",
			fmt.Sprintf(`"expires_at" must be a future timestamp or "" for no expiration. Got %q, which is already in the past.`, data.ExpiresAt.ValueString()),
		)
	}
}

func (r *ApiKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApiKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating API key", map[string]interface{}{"name": data.Name.ValueString()})

	var key *client.ApiKey
	var err error

	expiresMs, err := rfc3339ToMs(data.ExpiresAt.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid expires_at value", err.Error())
		return
	}

	if !data.ServiceUserID.IsNull() && !data.ServiceUserID.IsUnknown() && data.ServiceUserID.ValueString() != "" {
		key, err = r.client.CreateApiKeyForServiceUser(ctx, client.CreateApiKeyForServiceUserRequest{
			UserID:  data.ServiceUserID.ValueString(),
			Name:    data.Name.ValueString(),
			Expires: expiresMs,
		})
	} else {
		key, err = r.client.CreateApiKey(ctx, client.CreateApiKeyRequest{
			Name:    data.Name.ValueString(),
			Expires: expiresMs,
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API key, got error: %s", err))
		return
	}

	data.Key = types.StringValue(key.Key)
	data.Secret = types.StringValue(key.Secret)
	data.UserID = types.StringValue(key.UserID)
	data.Name = types.StringValue(key.Name)
	data.CreatedAt = types.StringValue(msToRFC3339(key.CreatedAt))
	data.ExpiresAt = types.StringValue(msToRFC3339(key.Expires))

	tflog.Trace(ctx, "Created API key resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApiKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := data.Key.ValueString()
	var key *client.ApiKey
	var err error

	if !data.ServiceUserID.IsNull() && !data.ServiceUserID.IsUnknown() && data.ServiceUserID.ValueString() != "" {
		key, err = r.client.GetApiKeyForServiceUser(ctx, data.ServiceUserID.ValueString(), keyID)
	} else {
		key, err = r.client.GetApiKey(ctx, keyID)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API key, got error: %s", err))
		return
	}

	if key == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Key = types.StringValue(key.Key)
	data.UserID = types.StringValue(key.UserID)
	data.Name = types.StringValue(key.Name)
	data.CreatedAt = types.StringValue(msToRFC3339(key.CreatedAt))
	data.ExpiresAt = types.StringValue(msToRFC3339(key.Expires))
	// Secret is only returned on creation; preserve the existing state value.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All mutable attributes have RequiresReplace, so Update is never called.
	var data ApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApiKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting API key", map[string]interface{}{"key": data.Key.ValueString()})

	var err error
	if !data.ServiceUserID.IsNull() && !data.ServiceUserID.IsUnknown() && data.ServiceUserID.ValueString() != "" {
		err = r.client.DeleteApiKeyForServiceUser(ctx, client.DeleteApiKeyForServiceUserRequest{
			Key: data.Key.ValueString(),
		})
	} else {
		err = r.client.DeleteApiKey(ctx, client.DeleteApiKeyRequest{
			Key: data.Key.ValueString(),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API key, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted API key resource")
}

func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
