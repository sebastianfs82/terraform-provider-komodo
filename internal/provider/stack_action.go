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

// ─── StackAction (unified) ───────────────────────────────────────────────────

var _ action.Action = (*StackAction)(nil)
var _ action.ActionWithConfigure = (*StackAction)(nil)

func NewStackAction() action.Action { return &StackAction{} }

type StackAction struct{ client *client.Client }

// StackActionModel is a superset of all individual stack action models.
type StackActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`

	// Shared by deploy / start / stop / pause / unpause / pull / restart / destroy
	Services types.List `tfsdk:"services"`

	// Shared by deploy / deploy_if_changed / stop / destroy
	StopTime types.Int64 `tfsdk:"stop_time"`

	// destroy only
	RemoveOrphans types.Bool `tfsdk:"remove_orphans"`

	// run_service — service identifier
	Service types.String `tfsdk:"service"`

	// run_service — command options
	Command      types.List   `tfsdk:"command"`
	NoTty        types.Bool   `tfsdk:"no_tty"`
	NoDeps       types.Bool   `tfsdk:"no_deps"`
	Detach       types.Bool   `tfsdk:"detach"`
	ServicePorts types.Bool   `tfsdk:"service_ports"`
	Env          types.Map    `tfsdk:"env"`
	Workdir      types.String `tfsdk:"workdir"`
	User         types.String `tfsdk:"user"`
	Entrypoint   types.String `tfsdk:"entrypoint"`
	Pull         types.Bool   `tfsdk:"pull"`
}

func (a *StackAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (a *StackAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo stack resource. " +
			"Set `action` to select the operation: `deploy`, `deploy_if_changed`, `destroy`, " +
			"`pause`, `unpause`, `pull`, `restart`, `run_service`, `start`, or `stop`.",
		Attributes: map[string]schema.Attribute{
			// ── required ──────────────────────────────────────────────────────
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target stack.",
			},
			"action": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The stack operation to perform. Valid values: `deploy`, `deploy_if_changed`, " +
					"`destroy`, `pause`, `unpause`, `pull`, `restart`, `run_service`, `start`, `stop`.",
			},

			// ── shared optional ───────────────────────────────────────────────
			"services": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter to specific services. Applies to: `deploy`, `destroy`, `pause`, `unpause`, `pull`, `restart`, `start`, `stop`. If empty, all services are targeted.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time in seconds. Applies to: `deploy`, `deploy_if_changed`, `destroy`, `stop`.",
			},

			// ── destroy ───────────────────────────────────────────────────────
			"remove_orphans": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Pass `--remove-orphans` to docker compose down. Applies to: `destroy`.",
			},

			// ── run_service ───────────────────────────────────────────────────
			"service": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service to run a one-off command against. Required when `action` is `run_service`.",
			},
			"command": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Command and arguments to pass to the container. Applies to: `run_service`.",
			},
			"no_tty": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Do not allocate a TTY. Applies to: `run_service`.",
			},
			"no_deps": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Do not start linked services. Applies to: `run_service`.",
			},
			"detach": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Detach container after run. Applies to: `run_service`.",
			},
			"service_ports": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Map service ports to the host. Applies to: `run_service`.",
			},
			"env": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra environment variables for the container. Applies to: `run_service`.",
			},
			"workdir": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Working directory inside the container. Applies to: `run_service`.",
			},
			"user": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User to run as inside the container. Applies to: `run_service`.",
			},
			"entrypoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the container entrypoint. Applies to: `run_service`.",
			},
			"pull": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Pull the image before running. Applies to: `run_service`.",
			},
		},
	}
}

func (a *StackAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *StackAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StackActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified stack action", map[string]interface{}{"stack": stackID, "action": act})

	switch act {
	case "deploy":
		a.invokeDeploy(ctx, stackID, data, resp)
	case "deploy_if_changed":
		a.invokeDeployIfChanged(ctx, stackID, data, resp)
	case "destroy":
		a.invokeDestroy(ctx, stackID, data, resp)
	case "pause":
		a.invokePause(ctx, stackID, data, resp)
	case "unpause":
		a.invokeUnpause(ctx, stackID, data, resp)
	case "pull":
		a.invokePull(ctx, stackID, data, resp)
	case "restart":
		a.invokeRestart(ctx, stackID, data, resp)
	case "run_service":
		a.invokeRunService(ctx, stackID, data, resp)
	case "start":
		a.invokeStart(ctx, stackID, data, resp)
	case "stop":
		a.invokeStop(ctx, stackID, data, resp)
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown stack action %q. Valid values are: deploy, deploy_if_changed, destroy, pause, unpause, pull, restart, run_service, start, stop.", act),
		)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func readServices(ctx context.Context, data StackActionModel, out *[]string, resp *action.InvokeResponse) bool {
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, out, false)...)
		return !resp.Diagnostics.HasError()
	}
	return true
}

func (a *StackAction) invokeDeploy(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.DeployStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}
	if err := a.client.DeployStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy stack, got error: %s", err))
	}
}

func (a *StackAction) invokeDeployIfChanged(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.DeployStackIfChangedRequest{Stack: stackID}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}
	if err := a.client.DeployStackIfChanged(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy stack if changed, got error: %s", err))
	}
}

func (a *StackAction) invokeDestroy(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.DestroyStackActionRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}
	if !data.RemoveOrphans.IsNull() && !data.RemoveOrphans.IsUnknown() {
		execReq.RemoveOrphans = data.RemoveOrphans.ValueBool()
	}
	if err := a.client.DestroyStackAction(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to destroy stack, got error: %s", err))
	}
}

func (a *StackAction) invokePause(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.PauseStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if err := a.client.PauseStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pause stack, got error: %s", err))
	}
}

func (a *StackAction) invokeUnpause(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.UnpauseStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if err := a.client.UnpauseStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unpause stack, got error: %s", err))
	}
}

func (a *StackAction) invokePull(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.PullStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if err := a.client.PullStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull stack images, got error: %s", err))
	}
}

func (a *StackAction) invokeRestart(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.RestartStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if err := a.client.RestartStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to restart stack, got error: %s", err))
	}
}

func (a *StackAction) invokeRunService(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	if data.Service.IsNull() || data.Service.IsUnknown() || data.Service.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid Configuration",
			`"service" is required when action is "run_service".`,
		)
		return
	}

	execReq := client.RunStackServiceRequest{
		Stack:   stackID,
		Service: data.Service.ValueString(),
	}
	if !data.Command.IsNull() && !data.Command.IsUnknown() {
		resp.Diagnostics.Append(data.Command.ElementsAs(ctx, &execReq.Command, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.NoTty.IsNull() && !data.NoTty.IsUnknown() {
		v := data.NoTty.ValueBool()
		execReq.NoTty = &v
	}
	if !data.NoDeps.IsNull() && !data.NoDeps.IsUnknown() {
		v := data.NoDeps.ValueBool()
		execReq.NoDeps = &v
	}
	if !data.Detach.IsNull() && !data.Detach.IsUnknown() {
		v := data.Detach.ValueBool()
		execReq.Detach = &v
	}
	if !data.ServicePorts.IsNull() && !data.ServicePorts.IsUnknown() {
		v := data.ServicePorts.ValueBool()
		execReq.ServicePorts = &v
	}
	if !data.Env.IsNull() && !data.Env.IsUnknown() {
		envMap := make(map[string]string)
		resp.Diagnostics.Append(data.Env.ElementsAs(ctx, &envMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		execReq.Env = envMap
	}
	if !data.Workdir.IsNull() && !data.Workdir.IsUnknown() {
		v := data.Workdir.ValueString()
		execReq.Workdir = &v
	}
	if !data.User.IsNull() && !data.User.IsUnknown() {
		v := data.User.ValueString()
		execReq.User = &v
	}
	if !data.Entrypoint.IsNull() && !data.Entrypoint.IsUnknown() {
		v := data.Entrypoint.ValueString()
		execReq.Entrypoint = &v
	}
	if !data.Pull.IsNull() && !data.Pull.IsUnknown() {
		v := data.Pull.ValueBool()
		execReq.Pull = &v
	}

	if err := a.client.RunStackService(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run stack service, got error: %s", err))
	}
}

func (a *StackAction) invokeStart(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.StartStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if err := a.client.StartStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start stack, got error: %s", err))
	}
}

func (a *StackAction) invokeStop(ctx context.Context, stackID string, data StackActionModel, resp *action.InvokeResponse) {
	execReq := client.StopStackRequest{Stack: stackID, Services: []string{}}
	if !readServices(ctx, data, &execReq.Services, resp) {
		return
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}
	if err := a.client.StopStack(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop stack, got error: %s", err))
	}
}
