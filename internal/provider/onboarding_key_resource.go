// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OnboardingKeyResource{}
var _ resource.ResourceWithImportState = &OnboardingKeyResource{}

func NewOnboardingKeyResource() resource.Resource {
	return &OnboardingKeyResource{}
}

// OnboardingKeyResource defines the resource implementation.
type OnboardingKeyResource struct {
	client *client.Client
}

// OnboardingKeyResourceModel describes the resource data model.
type OnboardingKeyResourceModel struct {
	PublicKey     types.String `tfsdk:"public_key"`
	PrivateKey    types.String `tfsdk:"private_key"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Expires       types.Int64  `tfsdk:"expires"`
	Tags          types.List   `tfsdk:"tags"`
	Privileged    types.Bool   `tfsdk:"privileged"`
	CopyServer    types.String `tfsdk:"copy_server"`
	CreateBuilder types.Bool   `tfsdk:"create_builder"`
	Onboarded     types.List   `tfsdk:"onboarded"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}

func (r *OnboardingKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_onboarding_key"
}

func (r *OnboardingKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo onboarding key used to onboard new servers.",

		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The onboarding key's unique public key identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The pkcs8-encoded private key (only available on creation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A human-friendly name for the onboarding key.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the onboarding key is enabled.",
			},
			"expires": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Expiration timestamp in milliseconds since epoch. Use 0 for no expiration.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default tags applied to servers onboarded using this key.",
			},
			"privileged": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When enabled, allows the key to enable disabled servers, remove address configuration, and update existing server public keys.",
			},
			"copy_server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "New servers onboarded by this key will copy configuration from the specified server.",
			},
			"create_builder": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to also create a Builder for servers onboarded using this key.",
			},
			"onboarded": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "IDs of servers that have been onboarded using this key.",
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp in milliseconds since epoch.",
			},
		},
	}
}

func (r *OnboardingKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OnboardingKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OnboardingKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := make([]string, 0)
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Creating onboarding key", map[string]interface{}{
		"name":    data.Name.ValueString(),
		"expires": data.Expires.ValueInt64(),
	})

	createResp, err := r.client.CreateOnboardingKey(ctx, client.CreateOnboardingKeyRequest{
		Name:          data.Name.ValueString(),
		Expires:       data.Expires.ValueInt64(),
		Tags:          tags,
		Privileged:    data.Privileged.ValueBool(),
		CopyServer:    data.CopyServer.ValueString(),
		CreateBuilder: data.CreateBuilder.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create onboarding key, got error: %s", err))
		return
	}

	k := createResp.Created

	data.PublicKey = types.StringValue(k.PublicKey)
	data.PrivateKey = types.StringValue(createResp.PrivateKey)
	data.Name = types.StringValue(k.Name)
	data.Enabled = types.BoolValue(k.Enabled)
	data.Expires = types.Int64Value(k.Expires)
	data.Privileged = types.BoolValue(k.Privileged)
	data.CopyServer = types.StringValue(k.CopyServer)
	data.CreateBuilder = types.BoolValue(k.CreateBuilder)
	data.CreatedAt = types.Int64Value(k.CreatedAt)

	tagList, diags := types.ListValueFrom(ctx, types.StringType, k.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagList

	onboardedList, diags := types.ListValueFrom(ctx, types.StringType, k.Onboarded)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Onboarded = onboardedList

	// Handle enabled: update if the plan requests disabled (API may default to enabled)
	if !data.Enabled.ValueBool() {
		enabled := false
		updated, err := r.client.UpdateOnboardingKey(ctx, client.UpdateOnboardingKeyRequest{
			PublicKey: k.PublicKey,
			Enabled:   &enabled,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update onboarding key enabled state after creation, got error: %s", err))
			return
		}
		data.Enabled = types.BoolValue(updated.Enabled)
	}

	tflog.Trace(ctx, "Created onboarding key resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnboardingKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OnboardingKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetOnboardingKey(ctx, data.PublicKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read onboarding key, got error: %s", err))
		return
	}

	if key == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.PublicKey = types.StringValue(key.PublicKey)
	data.Name = types.StringValue(key.Name)
	data.Enabled = types.BoolValue(key.Enabled)
	data.Expires = types.Int64Value(key.Expires)
	data.Privileged = types.BoolValue(key.Privileged)
	data.CopyServer = types.StringValue(key.CopyServer)
	data.CreateBuilder = types.BoolValue(key.CreateBuilder)
	data.CreatedAt = types.Int64Value(key.CreatedAt)
	// Note: private_key is not returned by List/Get — keep existing state value via UseStateForUnknown

	tagList, diags := types.ListValueFrom(ctx, types.StringType, key.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagList

	onboardedList, diags := types.ListValueFrom(ctx, types.StringType, key.Onboarded)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Onboarded = onboardedList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnboardingKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OnboardingKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := make([]string, 0)
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	name := data.Name.ValueString()
	expires := data.Expires.ValueInt64()
	enabled := data.Enabled.ValueBool()
	privileged := data.Privileged.ValueBool()
	copyServer := data.CopyServer.ValueString()
	createBuilder := data.CreateBuilder.ValueBool()

	tflog.Debug(ctx, "Updating onboarding key", map[string]interface{}{
		"public_key": data.PublicKey.ValueString(),
		"name":       name,
	})

	key, err := r.client.UpdateOnboardingKey(ctx, client.UpdateOnboardingKeyRequest{
		PublicKey:     data.PublicKey.ValueString(),
		Name:          &name,
		Expires:       &expires,
		Enabled:       &enabled,
		Tags:          &tags,
		Privileged:    &privileged,
		CopyServer:    &copyServer,
		CreateBuilder: &createBuilder,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update onboarding key, got error: %s", err))
		return
	}

	data.PublicKey = types.StringValue(key.PublicKey)
	data.Name = types.StringValue(key.Name)
	data.Enabled = types.BoolValue(key.Enabled)
	data.Expires = types.Int64Value(key.Expires)
	data.Privileged = types.BoolValue(key.Privileged)
	data.CopyServer = types.StringValue(key.CopyServer)
	data.CreateBuilder = types.BoolValue(key.CreateBuilder)
	data.CreatedAt = types.Int64Value(key.CreatedAt)

	tagList, diags := types.ListValueFrom(ctx, types.StringType, key.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagList

	onboardedList, diags := types.ListValueFrom(ctx, types.StringType, key.Onboarded)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Onboarded = onboardedList

	tflog.Trace(ctx, "Updated onboarding key resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnboardingKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OnboardingKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting onboarding key", map[string]interface{}{
		"public_key": data.PublicKey.ValueString(),
	})

	err := r.client.DeleteOnboardingKey(ctx, client.DeleteOnboardingKeyRequest{
		PublicKey: data.PublicKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete onboarding key, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted onboarding key resource")
}

func (r *OnboardingKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("public_key"), req, resp)
}
