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

// ─── shared configure helper ─────────────────────────────────────────────────

func miscActionConfigure(providerData any, addError func(string, string)) *client.Client {
	if providerData == nil {
		return nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		addError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}
	return c
}

// ─── StartDeployment ─────────────────────────────────────────────────────────

var _ action.Action = (*StartDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*StartDeploymentAction)(nil)

func NewStartDeploymentAction() action.Action { return &StartDeploymentAction{} }

type StartDeploymentAction struct{ client *client.Client }

type StartDeploymentModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *StartDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_start_deployment"
}

func (a *StartDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Starts the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to start.",
			},
		},
	}
}

func (a *StartDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *StartDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StartDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing StartDeployment", map[string]interface{}{"deployment": data.ID.ValueString()})
	if err := a.client.StartDeployment(ctx, client.StartDeploymentRequest{Deployment: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "StartDeployment action completed")
}

// ─── PullDeployment ──────────────────────────────────────────────────────────

var _ action.Action = (*PullDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*PullDeploymentAction)(nil)

func NewPullDeploymentAction() action.Action { return &PullDeploymentAction{} }

type PullDeploymentAction struct{ client *client.Client }

type PullDeploymentModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *PullDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pull_deployment"
}

func (a *PullDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pulls the latest image for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to pull.",
			},
		},
	}
}

func (a *PullDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *PullDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PullDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PullDeployment", map[string]interface{}{"deployment": data.ID.ValueString()})
	if err := a.client.PullDeployment(ctx, client.PullDeploymentRequest{Deployment: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PullDeployment action completed")
}

// ─── DeployDeployment ────────────────────────────────────────────────────────

var _ action.Action = (*DeployDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*DeployDeploymentAction)(nil)

func NewDeployDeploymentAction() action.Action { return &DeployDeploymentAction{} }

type DeployDeploymentAction struct{ client *client.Client }

type DeployDeploymentModel struct {
	ID         types.String `tfsdk:"id"`
	StopSignal types.String `tfsdk:"stop_signal"`
	StopTime   types.Int64  `tfsdk:"stop_time"`
}

func (a *DeployDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deploy_deployment"
}

func (a *DeployDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deploys (or redeploys) the container for the target deployment. Pulls the image, stops the existing container, and starts a new one.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to deploy.",
			},
			"stop_signal": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination signal (e.g. `SIGTERM`, `SIGINT`). Only used when the container must be stopped first.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds). Only used when the container must be stopped first.",
			},
		},
	}
}

func (a *DeployDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *DeployDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DeployDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	execReq := client.DeployRequest{
		Deployment: data.ID.ValueString(),
	}
	if !data.StopSignal.IsNull() && !data.StopSignal.IsUnknown() {
		execReq.StopSignal = data.StopSignal.ValueString()
	}
	if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
		v := data.StopTime.ValueInt64()
		execReq.StopTime = &v
	}
	tflog.Debug(ctx, "Executing Deploy", map[string]interface{}{"deployment": execReq.Deployment})
	if err := a.client.Deploy(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deploy action completed")
}

// ─── StopDeployment ──────────────────────────────────────────────────────────

var _ action.Action = (*StopDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*StopDeploymentAction)(nil)

func NewStopDeploymentAction() action.Action { return &StopDeploymentAction{} }

type StopDeploymentAction struct{ client *client.Client }

type StopDeploymentModel struct {
	ID     types.String `tfsdk:"id"`
	Signal types.String `tfsdk:"signal"`
	Time   types.Int64  `tfsdk:"time"`
}

func (a *StopDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stop_deployment"
}

func (a *StopDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stops the container for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to stop.",
			},
			"signal": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination signal (e.g. `SIGTERM`, `SIGINT`).",
			},
			"time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds).",
			},
		},
	}
}

func (a *StopDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *StopDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StopDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	execReq := client.StopDeploymentRequest{
		Deployment: data.ID.ValueString(),
	}
	if !data.Signal.IsNull() && !data.Signal.IsUnknown() {
		execReq.Signal = data.Signal.ValueString()
	}
	if !data.Time.IsNull() && !data.Time.IsUnknown() {
		v := data.Time.ValueInt64()
		execReq.Time = &v
	}
	tflog.Debug(ctx, "Executing StopDeployment", map[string]interface{}{"deployment": execReq.Deployment})
	if err := a.client.StopDeployment(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "StopDeployment action completed")
}

// ─── DestroyDeployment ───────────────────────────────────────────────────────

var _ action.Action = (*DestroyDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*DestroyDeploymentAction)(nil)

func NewDestroyDeploymentAction() action.Action { return &DestroyDeploymentAction{} }

type DestroyDeploymentAction struct{ client *client.Client }

type DestroyDeploymentModel struct {
	ID     types.String `tfsdk:"id"`
	Signal types.String `tfsdk:"signal"`
	Time   types.Int64  `tfsdk:"time"`
}

func (a *DestroyDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destroy_deployment"
}

func (a *DestroyDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stops and removes the container for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to destroy.",
			},
			"signal": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination signal (e.g. `SIGTERM`, `SIGINT`).",
			},
			"time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds).",
			},
		},
	}
}

func (a *DestroyDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *DestroyDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DestroyDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	execReq := client.DestroyDeploymentRequest{
		Deployment: data.ID.ValueString(),
	}
	if !data.Signal.IsNull() && !data.Signal.IsUnknown() {
		execReq.Signal = data.Signal.ValueString()
	}
	if !data.Time.IsNull() && !data.Time.IsUnknown() {
		v := data.Time.ValueInt64()
		execReq.Time = &v
	}
	tflog.Debug(ctx, "Executing DestroyDeployment", map[string]interface{}{"deployment": execReq.Deployment})
	if err := a.client.DestroyDeployment(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to destroy deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "DestroyDeployment action completed")
}

// ─── RestartDeployment ───────────────────────────────────────────────────────

var _ action.Action = (*RestartDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*RestartDeploymentAction)(nil)

func NewRestartDeploymentAction() action.Action { return &RestartDeploymentAction{} }

type RestartDeploymentAction struct{ client *client.Client }

type RestartDeploymentModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *RestartDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_restart_deployment"
}

func (a *RestartDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Restarts the container for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to restart.",
			},
		},
	}
}

func (a *RestartDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *RestartDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RestartDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing RestartDeployment", map[string]interface{}{"deployment": data.ID.ValueString()})
	if err := a.client.RestartDeployment(ctx, client.RestartDeploymentRequest{Deployment: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to restart deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RestartDeployment action completed")
}

// ─── PauseDeployment ─────────────────────────────────────────────────────────

var _ action.Action = (*PauseDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*PauseDeploymentAction)(nil)

func NewPauseDeploymentAction() action.Action { return &PauseDeploymentAction{} }

type PauseDeploymentAction struct{ client *client.Client }

type PauseDeploymentModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *PauseDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pause_deployment"
}

func (a *PauseDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pauses the container for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to pause.",
			},
		},
	}
}

func (a *PauseDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *PauseDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PauseDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PauseDeployment", map[string]interface{}{"deployment": data.ID.ValueString()})
	if err := a.client.PauseDeployment(ctx, client.PauseDeploymentRequest{Deployment: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pause deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PauseDeployment action completed")
}

// ─── UnpauseDeployment ───────────────────────────────────────────────────────

var _ action.Action = (*UnpauseDeploymentAction)(nil)
var _ action.ActionWithConfigure = (*UnpauseDeploymentAction)(nil)

func NewUnpauseDeploymentAction() action.Action { return &UnpauseDeploymentAction{} }

type UnpauseDeploymentAction struct{ client *client.Client }

type UnpauseDeploymentModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *UnpauseDeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unpause_deployment"
}

func (a *UnpauseDeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Unpauses the container for the target deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the deployment to unpause.",
			},
		},
	}
}

func (a *UnpauseDeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *UnpauseDeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data UnpauseDeploymentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing UnpauseDeployment", map[string]interface{}{"deployment": data.ID.ValueString()})
	if err := a.client.UnpauseDeployment(ctx, client.UnpauseDeploymentRequest{Deployment: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unpause deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "UnpauseDeployment action completed")
}
