// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &VariableResource{}
var _ resource.ResourceWithImportState = &VariableResource{}

func NewVariableResource() resource.Resource {
	return &VariableResource{}
}

type VariableResource struct {
	client *client.Client
}

type VariableResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
	IsSecret    types.Bool   `tfsdk:"is_secret"`
}

func (r *VariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

func (r *VariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The variable identifier (same as name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The variable name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The variable value.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The variable description.",
			},
			"is_secret": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the variable is secret.",
			},
		},
	}
}

func (r *VariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating variable", map[string]interface{}{
		"name":        data.Name.ValueString(),
		"is_secret":   data.IsSecret.ValueBool(),
		"description": data.Description.ValueString(),
		"value":       data.Value.ValueString(),
	})
	createReq := client.CreateVariableRequest{
		Name:        data.Name.ValueString(),
		Value:       data.Value.ValueString(),
		Description: data.Description.ValueString(),
		IsSecret:    data.IsSecret.ValueBool(),
	}
	variable, err := r.client.CreateVariable(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create variable, got error: %s", err))
		return
	}
	data.ID = types.StringValue(variable.Name)
	data.Name = types.StringValue(variable.Name)
	data.Value = types.StringValue(variable.Value)
	data.Description = types.StringValue(variable.Description)
	data.IsSecret = types.BoolValue(variable.IsSecret)
	tflog.Trace(ctx, "Created variable resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	variable, err := r.client.GetVariable(ctx, data.Name.ValueString())
	if err != nil {
		// If not found, remove from state so Terraform will recreate it
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		// For other errors, report as a real error
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read variable, got error: %s", err))
		return
	}
	data.ID = types.StringValue(variable.Name)
	data.Name = types.StringValue(variable.Name)
	data.Value = types.StringValue(variable.Value)
	data.Description = types.StringValue(variable.Description)
	data.IsSecret = types.BoolValue(variable.IsSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := client.CreateVariableRequest{
		Name:        data.Name.ValueString(),
		Value:       data.Value.ValueString(),
		Description: data.Description.ValueString(),
		IsSecret:    data.IsSecret.ValueBool(),
	}
	variable, err := r.client.UpdateVariable(ctx, data.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update variable, got error: %s", err))
		return
	}
	data.ID = types.StringValue(variable.Name)
	data.Name = types.StringValue(variable.Name)
	data.Value = types.StringValue(variable.Value)
	data.Description = types.StringValue(variable.Description)
	data.IsSecret = types.BoolValue(variable.IsSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting variable", map[string]interface{}{
		"name": data.Name.ValueString(),
	})
	err := r.client.DeleteVariable(ctx, client.DeleteVariableRequest{
		ID: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete variable, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted variable resource")
}

func (r *VariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
