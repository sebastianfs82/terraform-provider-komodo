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

var _ resource.Resource = &RegistryAccountResource{}
var _ resource.ResourceWithImportState = &RegistryAccountResource{}

func NewRegistryAccountResource() resource.Resource {
	return &RegistryAccountResource{}
}

type RegistryAccountResource struct {
	client *client.Client
}

type RegistryAccountResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Domain   types.String `tfsdk:"domain"`
	Username types.String `tfsdk:"username"`
	Token    types.String `tfsdk:"token"`
}

func (r *RegistryAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_account"
}

func (r *RegistryAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo docker registry account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The docker registry account identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The registry domain (e.g. `registry.example.com`). Leave empty or omit for Docker Hub (`docker.io`).",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The account username.",
			},
			"token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The plaintext access token (password) for the account.",
			},
		},
	}
}

func (r *RegistryAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RegistryAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RegistryAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating docker registry account", map[string]interface{}{
		"domain":   data.Domain.ValueString(),
		"username": data.Username.ValueString(),
	})
	createReq := client.CreateDockerRegistryAccountRequest{
		Domain:   data.Domain.ValueString(),
		Username: data.Username.ValueString(),
		Token:    data.Token.ValueString(),
	}
	account, err := r.client.CreateDockerRegistryAccount(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create docker registry account, got error: %s", err))
		return
	}
	if account.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Docker registry account creation failed: missing ID",
			"The Komodo API did not return an account ID. Resource cannot be tracked in state.",
		)
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.Username = types.StringValue(account.Username)
	data.Token = types.StringValue(account.Token)
	tflog.Trace(ctx, "Created docker registry account resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegistryAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RegistryAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.client.GetDockerRegistryAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read docker registry account, got error: %s", err))
		return
	}
	if account == nil {
		tflog.Debug(ctx, "Docker registry account not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.Username = types.StringValue(account.Username)
	if account.Token != "" {
		data.Token = types.StringValue(account.Token)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegistryAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RegistryAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := client.CreateDockerRegistryAccountRequest{
		Domain:   data.Domain.ValueString(),
		Username: data.Username.ValueString(),
		Token:    data.Token.ValueString(),
	}
	account, err := r.client.UpdateDockerRegistryAccount(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update docker registry account, got error: %s", err))
		return
	}
	data.ID = types.StringValue(account.ID.OID)
	data.Domain = types.StringValue(account.Domain)
	data.Username = types.StringValue(account.Username)
	if account.Token != "" {
		data.Token = types.StringValue(account.Token)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegistryAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RegistryAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting docker registry account", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	err := r.client.DeleteDockerRegistryAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete docker registry account, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted docker registry account resource")
}

func (r *RegistryAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
