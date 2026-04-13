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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &TerminalResource{}
var _ resource.ResourceWithImportState = &TerminalResource{}
var _ resource.ResourceWithValidateConfig = &TerminalResource{}

func NewTerminalResource() resource.Resource {
	return &TerminalResource{}
}

type TerminalResource struct {
	client *client.Client
}

type TerminalResourceModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	TargetType   types.String  `tfsdk:"target_type"`
	TargetID     types.String  `tfsdk:"target_id"`
	Container    types.String  `tfsdk:"container"`
	Service      types.String  `tfsdk:"service"`
	Mode         types.String  `tfsdk:"mode"`
	Command      types.String  `tfsdk:"command"`
	CreatedAt    types.String  `tfsdk:"created_at"`
	StoredSizeKB types.Float64 `tfsdk:"stored_size_kb"`
}

func (r *TerminalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminal"
}

func (r *TerminalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo terminal session on a target resource. " +
			"The `target_type`, `target_id`, and `name` attributes force a new resource when changed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier in the form `target_id:name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name for the terminal session. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of target for the terminal. One of `Server`, `Container`, `Stack`, `Deployment`. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The primary target ID or name. For `Server`: the server. For `Container`: the server hosting the container. For `Stack`: the stack. For `Deployment`: the deployment. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"container": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The container name. Only used when `target_type` is `Container`. Changing this forces a new resource.",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The stack service name. Only used when `target_type` is `Stack`. Changing this forces a new resource.",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The terminal mode. One of `exec` or `attach`. Only supported for `Container` and `Stack` target types. Ignored for `Server` and `Deployment`. Defaults to `exec`. Changing this forces a new resource.",
				Default:             stringdefault.StaticString("exec"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"command": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The shell command to initialise the terminal (e.g. `bash`). Defaults to the server's configured shell.",
				Default:             stringdefault.StaticString(""),
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
			"stored_size_kb": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "The size of stored terminal output in kilobytes.",
			},
		},
	}
}

func (r *TerminalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TerminalResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data TerminalResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// attach mode is only supported for Container and Stack.
	if data.Mode.ValueString() == "attach" {
		tt := data.TargetType.ValueString()
		if tt != "Container" && tt != "Stack" {
			resp.Diagnostics.AddAttributeError(
				path.Root("mode"),
				"Invalid mode",
				fmt.Sprintf("mode=\"attach\" is only supported for target_type=\"Container\" or target_type=\"Stack\", got %q.", tt),
			)
		}
	}
}

func (r *TerminalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TerminalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := data.TargetType.ValueString()
	targetID := data.TargetID.ValueString()

	tflog.Debug(ctx, "Creating terminal", map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
		"name":        data.Name.ValueString(),
	})

	name := data.Name.ValueString()
	createTarget := client.NewTerminalTarget(targetType, targetID, data.Container.ValueString(), data.Service.ValueString())
	createReq := client.CreateTerminalRequest{
		Target:   createTarget,
		Name:     &name,
		Recreate: "Never",
	}
	// Set mode only for target types that support it.
	if tt := data.TargetType.ValueString(); tt == "Container" || tt == "Stack" || tt == "Deployment" {
		if m := data.Mode.ValueString(); m != "" {
			createReq.Mode = &m
		}
	}
	if cmd := data.Command.ValueString(); cmd != "" {
		createReq.Command = &cmd
	}

	t, err := r.client.CreateTerminal(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create terminal, got error: %s", err))
		return
	}

	data.ID = types.StringValue(targetID + ":" + t.Name)
	data.Name = types.StringValue(t.Name)
	data.CreatedAt = types.StringValue(msToRFC3339(t.CreatedAt))
	data.StoredSizeKB = types.Float64Value(t.StoredSizeKB)
	data.Container = types.StringValue(t.Target.ContainerName())
	data.Service = types.StringValue(t.Target.ServiceName())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TerminalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := data.TargetType.ValueString()
	targetID := data.TargetID.ValueString()
	if targetID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	readTarget := client.NewTerminalTarget(targetType, targetID, data.Container.ValueString(), data.Service.ValueString())
	terminals, err := r.client.ListTerminals(ctx, client.ListTerminalsRequest{
		Target: &readTarget,
	})
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "did not find") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list terminals, got error: %s", err))
		return
	}

	var found *client.Terminal
	for i := range terminals {
		if terminals[i].Name == data.Name.ValueString() {
			found = &terminals[i]
			break
		}
	}

	if found == nil {
		tflog.Debug(ctx, "Terminal not found, removing from state", map[string]interface{}{
			"target_id": targetID,
			"name":      data.Name.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(targetID + ":" + found.Name)
	data.Name = types.StringValue(found.Name)
	data.CreatedAt = types.StringValue(msToRFC3339(found.CreatedAt))
	data.StoredSizeKB = types.Float64Value(found.StoredSizeKB)
	data.Container = types.StringValue(found.Target.ContainerName())
	data.Service = types.StringValue(found.Target.ServiceName())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replacement; this method should never be called.
}

func (r *TerminalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TerminalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := data.TargetType.ValueString()
	targetID := data.TargetID.ValueString()

	tflog.Debug(ctx, "Deleting terminal", map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
		"name":        data.Name.ValueString(),
	})

	delTarget := client.NewTerminalTarget(targetType, targetID, data.Container.ValueString(), data.Service.ValueString())
	err := r.client.DeleteTerminal(ctx, client.DeleteTerminalRequest{
		Target:   delTarget,
		Terminal: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete terminal, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted terminal resource")
}

func (r *TerminalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: "target_id:name" (assumes Server type)
	//         or "target_type:target_id:name" for other target types.
	parts := strings.SplitN(req.ID, ":", 3)
	var targetType, targetID, terminalName string
	switch len(parts) {
	case 2:
		if parts[0] != "" && parts[1] != "" {
			targetType = "Server"
			targetID = parts[0]
			terminalName = parts[1]
		}
	case 3:
		if parts[0] != "" && parts[1] != "" && parts[2] != "" {
			targetType = parts[0]
			targetID = parts[1]
			terminalName = parts[2]
		}
	}
	if targetID == "" || terminalName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import id in the format 'target_id:name' or 'target_type:target_id:name', got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), targetID+":"+terminalName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), targetType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_id"), targetID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), terminalName)...)
}
