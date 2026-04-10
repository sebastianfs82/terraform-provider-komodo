package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}

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
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	Expires       types.Int64  `tfsdk:"expires"`
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
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp in milliseconds since epoch.",
			},
			"expires": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Expiration timestamp in milliseconds since epoch. Use 0 for no expiration.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
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

	if !data.ServiceUserID.IsNull() && !data.ServiceUserID.IsUnknown() && data.ServiceUserID.ValueString() != "" {
		key, err = r.client.CreateApiKeyForServiceUser(ctx, client.CreateApiKeyForServiceUserRequest{
			UserID:  data.ServiceUserID.ValueString(),
			Name:    data.Name.ValueString(),
			Expires: data.Expires.ValueInt64(),
		})
	} else {
		key, err = r.client.CreateApiKey(ctx, client.CreateApiKeyRequest{
			Name:    data.Name.ValueString(),
			Expires: data.Expires.ValueInt64(),
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
	data.CreatedAt = types.Int64Value(key.CreatedAt)
	data.Expires = types.Int64Value(key.Expires)

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
	data.CreatedAt = types.Int64Value(key.CreatedAt)
	data.Expires = types.Int64Value(key.Expires)
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
