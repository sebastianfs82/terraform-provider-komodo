// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ProviderAccountResource{}
var _ resource.ResourceWithImportState = &ProviderAccountResource{}

func NewProviderAccountResource() resource.Resource {
	return &ProviderAccountResource{}
}

type ProviderAccountResource struct {
	client *client.Client
}

type ProviderAccountResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Domain       types.String `tfsdk:"domain"`
	HttpsEnabled types.Bool   `tfsdk:"https_enabled"`
	Username     types.String `tfsdk:"username"`
	Token        types.String `tfsdk:"token"`
}

func (r *ProviderAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_account"
}

func (r *ProviderAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo git provider account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git provider account identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The git provider domain without protocol prefix (e.g. `github.com`).",
			},
			"https_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use HTTPS (true) or HTTP (false) for cloning.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The account username.",
			},
			"token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The plaintext access token for the account.",
			},
		},
	}
}

func (r *ProviderAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProviderAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProviderAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating git provider account", map[string]interface{}{
		"domain":        data.Domain.ValueString(),
		"https_enabled": data.HttpsEnabled.ValueBool(),
		"username":      data.Username.ValueString(),
	})
	createReq := client.CreateGitProviderAccountRequest{
		Domain:       data.Domain.ValueString(),
		HttpsEnabled: data.HttpsEnabled.ValueBool(),
		Username:     data.Username.ValueString(),
		Token:        data.Token.ValueString(),
	}
	account, err := r.client.CreateGitProviderAccount(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create git provider account, got error: %s", err))
		return
	}
	if account.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Git provider account creation failed: missing ID",
			"The Komodo API did not return an account ID. Resource cannot be tracked in state.",
		)
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.HttpsEnabled = types.BoolValue(account.HttpsEnabled)
	data.Username = types.StringValue(account.Username)
	data.Token = types.StringValue(account.Token)
	tflog.Trace(ctx, "Created git provider account resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProviderAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProviderAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.client.GetGitProviderAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read git provider account, got error: %s", err))
		return
	}
	if account == nil {
		tflog.Debug(ctx, "Git provider account not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.HttpsEnabled = types.BoolValue(account.HttpsEnabled)
	data.Username = types.StringValue(account.Username)
	if account.Token != "" {
		data.Token = types.StringValue(account.Token)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProviderAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProviderAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := client.CreateGitProviderAccountRequest{
		Domain:       data.Domain.ValueString(),
		HttpsEnabled: data.HttpsEnabled.ValueBool(),
		Username:     data.Username.ValueString(),
		Token:        data.Token.ValueString(),
	}
	account, err := r.client.UpdateGitProviderAccount(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update git provider account, got error: %s", err))
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.HttpsEnabled = types.BoolValue(account.HttpsEnabled)
	data.Username = types.StringValue(account.Username)
	if account.Token != "" {
		data.Token = types.StringValue(account.Token)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProviderAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProviderAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting git provider account", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	err := r.client.DeleteGitProviderAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete git provider account, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted git provider account resource")
}

func (r *ProviderAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
