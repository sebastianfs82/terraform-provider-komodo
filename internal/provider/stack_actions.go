// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// ─── DeployStack ─────────────────────────────────────────────────────────────

var _ action.Action = (*StackDeployAction)(nil)
var _ action.ActionWithConfigure = (*StackDeployAction)(nil)

func NewStackDeployAction() action.Action { return &StackDeployAction{} }

type StackDeployAction struct{ client *client.Client }

type StackDeployModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
	StopTime types.Int64  `tfsdk:"stop_time"`
}

func (a *StackDeployAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_deploy"
}

func (a *StackDeployAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose up` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to deploy.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only deploy specific services. If empty, deploys all services.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds). Only used if the stack needs to be taken down first.",
			},
		},
	}
}

func (a *StackDeployAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackDeployAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackDeployModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.DeployStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}

	tflog.Debug(ctx, "Executing DeployStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.DeployStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "DeployStack action completed")
}

// ─── StartStack ──────────────────────────────────────────────────────────────

var _ action.Action = (*StackStartAction)(nil)
var _ action.ActionWithConfigure = (*StackStartAction)(nil)

func NewStackStartAction() action.Action { return &StackStartAction{} }

type StackStartAction struct{ client *client.Client }

type StackStartModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
}

func (a *StackStartAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_start"
}

func (a *StackStartAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose start` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to start.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only start specific services. If empty, starts all services.",
			},
		},
	}
}

func (a *StackStartAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackStartAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackStartModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.StartStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Executing StartStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.StartStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "StartStack action completed")
}

// ─── StopStack ───────────────────────────────────────────────────────────────

var _ action.Action = (*StackStopAction)(nil)
var _ action.ActionWithConfigure = (*StackStopAction)(nil)

func NewStackStopAction() action.Action { return &StackStopAction{} }

type StackStopAction struct{ client *client.Client }

type StackStopModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
	StopTime types.Int64  `tfsdk:"stop_time"`
}

func (a *StackStopAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_stop"
}

func (a *StackStopAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose stop` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to stop.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only stop specific services. If empty, stops all services.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds).",
			},
		},
	}
}

func (a *StackStopAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackStopAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackStopModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.StopStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}

	tflog.Debug(ctx, "Executing StopStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.StopStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "StopStack action completed")
}

// ─── PauseStack ──────────────────────────────────────────────────────────────

var _ action.Action = (*StackPauseAction)(nil)
var _ action.ActionWithConfigure = (*StackPauseAction)(nil)

func NewStackPauseAction() action.Action { return &StackPauseAction{} }

type StackPauseAction struct{ client *client.Client }

type StackPauseModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
}

func (a *StackPauseAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_pause"
}

func (a *StackPauseAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose pause` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to pause.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only pause specific services. If empty, pauses all services.",
			},
		},
	}
}

func (a *StackPauseAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackPauseAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackPauseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.PauseStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Executing PauseStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.PauseStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pause stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PauseStack action completed")
}

// ─── DestroyStack ─────────────────────────────────────────────────────────────

var _ action.Action = (*StackDestroyAction)(nil)
var _ action.ActionWithConfigure = (*StackDestroyAction)(nil)

func NewStackDestroyAction() action.Action { return &StackDestroyAction{} }

type StackDestroyAction struct{ client *client.Client }

type StackDestroyModel struct {
	Stack         types.String `tfsdk:"stack"`
	Services      types.List   `tfsdk:"services"`
	RemoveOrphans types.Bool   `tfsdk:"remove_orphans"`
	StopTime      types.Int64  `tfsdk:"stop_time"`
}

func (a *StackDestroyAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_destroy"
}

func (a *StackDestroyAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose down` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to destroy.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only destroy specific services. If empty, destroys all services.",
			},
			"remove_orphans": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Pass `--remove-orphans` to docker compose down.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds).",
			},
		},
	}
}

func (a *StackDestroyAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackDestroyAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackDestroyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.DestroyStackActionRequest{
		Stack:         data.Stack.ValueString(),
		Services:      []string{},
		RemoveOrphans: data.RemoveOrphans.ValueBool(),
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}

	tflog.Debug(ctx, "Executing DestroyStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.DestroyStackAction(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to destroy stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "DestroyStack action completed")
}

// ─── RestartStack ─────────────────────────────────────────────────────────────

var _ action.Action = (*StackRestartAction)(nil)
var _ action.ActionWithConfigure = (*StackRestartAction)(nil)

func NewStackRestartAction() action.Action { return &StackRestartAction{} }

type StackRestartAction struct{ client *client.Client }

type StackRestartModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
}

func (a *StackRestartAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_restart"
}

func (a *StackRestartAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose restart` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to restart.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only restart specific services. If empty, restarts all services.",
			},
		},
	}
}

func (a *StackRestartAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackRestartAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackRestartModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.RestartStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Executing RestartStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.RestartStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to restart stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RestartStack action completed")
}

// ─── UnpauseStack ─────────────────────────────────────────────────────────────

var _ action.Action = (*StackUnpauseAction)(nil)
var _ action.ActionWithConfigure = (*StackUnpauseAction)(nil)

func NewStackUnpauseAction() action.Action { return &StackUnpauseAction{} }

type StackUnpauseAction struct{ client *client.Client }

type StackUnpauseModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
}

func (a *StackUnpauseAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_unpause"
}

func (a *StackUnpauseAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose unpause` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to unpause.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only unpause specific services. If empty, unpauses all services.",
			},
		},
	}
}

func (a *StackUnpauseAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackUnpauseAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackUnpauseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.UnpauseStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Executing UnpauseStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.UnpauseStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unpause stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "UnpauseStack action completed")
}

// ─── PullStack ─────────────────────────────────────────────────────────────

var _ action.Action = (*StackPullAction)(nil)
var _ action.ActionWithConfigure = (*StackPullAction)(nil)

func NewStackPullAction() action.Action { return &StackPullAction{} }

type StackPullAction struct{ client *client.Client }

type StackPullModel struct {
	Stack    types.String `tfsdk:"stack"`
	Services types.List   `tfsdk:"services"`
}

func (a *StackPullAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_pull"
}

func (a *StackPullAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a `docker compose pull` on the target Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"stack": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the stack to pull images for.",
			},
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to only pull images for specific services. If empty, pulls for all services.",
			},
		},
	}
}

func (a *StackPullAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *StackPullAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackPullModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.PullStackRequest{
		Stack:    data.Stack.ValueString(),
		Services: []string{},
	}
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &execReq.Services, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Executing PullStack", map[string]interface{}{"stack": execReq.Stack})
	if err := a.client.PullStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull stack images, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PullStack action completed")
}
