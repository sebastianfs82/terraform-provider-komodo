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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &SwarmSecretResource{}
var _ resource.ResourceWithImportState = &SwarmSecretResource{}

func NewSwarmSecretResource() resource.Resource {
	return &SwarmSecretResource{}
}

type SwarmSecretResource struct {
	client *client.Client
}

type SwarmSecretResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Swarm          types.String `tfsdk:"swarm"`
	Name           types.String `tfsdk:"name"`
	Data           types.String `tfsdk:"data"`
	Driver         types.String `tfsdk:"driver"`
	Labels         types.List   `tfsdk:"labels"`
	TemplateDriver types.String `tfsdk:"template_driver"`
}

func (r *SwarmSecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_swarm_secret"
}

func (r *SwarmSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Docker Swarm secret on a Komodo swarm. Updating `data` rotates the secret in-place.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier in the form `swarm:name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"swarm": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The swarm name or ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The secret value. Changing this rotates the secret.",
			},
			"driver": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional custom secret driver.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Docker labels to set on the secret.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"template_driver": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional custom template driver.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SwarmSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SwarmSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SwarmSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := make([]string, 0)
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateSwarmSecretRequest{
		Swarm:  data.Swarm.ValueString(),
		Name:   data.Name.ValueString(),
		Data:   data.Data.ValueString(),
		Labels: labels,
	}
	if v := data.Driver.ValueString(); v != "" {
		createReq.Driver = &v
	}
	if v := data.TemplateDriver.ValueString(); v != "" {
		createReq.TemplateDriver = &v
	}

	tflog.Debug(ctx, "Creating swarm secret", map[string]interface{}{"swarm": createReq.Swarm, "name": createReq.Name})
	if err := r.client.CreateSwarmSecret(ctx, createReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create swarm secret, got error: %s", err))
		return
	}

	data.ID = types.StringValue(data.Swarm.ValueString() + ":" + data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwarmSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SwarmSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := r.client.ListSwarmSecrets(ctx, client.ListSwarmSecretsRequest{Swarm: data.Swarm.ValueString()})
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "did not find") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list swarm secrets, got error: %s", err))
		return
	}

	found := false
	for _, it := range items {
		if it.Name != nil && *it.Name == data.Name.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(data.Swarm.ValueString() + ":" + data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwarmSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SwarmSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SwarmSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Data.ValueString() != state.Data.ValueString() {
		tflog.Debug(ctx, "Rotating swarm secret", map[string]interface{}{"swarm": state.Swarm.ValueString(), "name": state.Name.ValueString()})
		if err := r.client.RotateSwarmSecret(ctx, client.RotateSwarmSecretRequest{
			Swarm:  state.Swarm.ValueString(),
			Secret: state.Name.ValueString(),
			Data:   plan.Data.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotate swarm secret, got error: %s", err))
			return
		}
	}

	plan.ID = types.StringValue(plan.Swarm.ValueString() + ":" + plan.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SwarmSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SwarmSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting swarm secret", map[string]interface{}{"swarm": data.Swarm.ValueString(), "name": data.Name.ValueString()})
	if err := r.client.RemoveSwarmSecrets(ctx, client.RemoveSwarmSecretsRequest{
		Swarm:   data.Swarm.ValueString(),
		Secrets: []string{data.Name.ValueString()},
	}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete swarm secret, got error: %s", err))
		return
	}
}

func (r *SwarmSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import id in the format 'swarm:name', got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("swarm"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
