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

// ─── DeploymentAction (unified) ───────────────────────────────────────────────

var _ action.Action = (*DeploymentAction)(nil)
var _ action.ActionWithConfigure = (*DeploymentAction)(nil)

func NewDeploymentAction() action.Action { return &DeploymentAction{} }

type DeploymentAction struct{ client *client.Client }

// DeploymentActionModel is a superset of all individual deployment action models.
type DeploymentActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`

	// deploy only: stop options when taking the container down before redeploying
	StopSignal types.String `tfsdk:"stop_signal"`
	StopTime   types.Int64  `tfsdk:"stop_time"`

	// stop / destroy: termination signal and timeout
	Signal types.String `tfsdk:"signal"`
	Time   types.Int64  `tfsdk:"time"`
}

func (a *DeploymentAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (a *DeploymentAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo deployment resource. " +
			"Set `action` to select the operation: `deploy`, `destroy`, `pause`, `unpause`, " +
			"`pull`, `restart`, `start`, or `stop`.",
		Attributes: map[string]schema.Attribute{
			// ── required ──────────────────────────────────────────────────────
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target deployment.",
			},
			"action": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The deployment operation to perform. Valid values: `deploy`, `destroy`, " +
					"`pause`, `unpause`, `pull`, `restart`, `start`, `stop`.",
			},

			// ── deploy only ───────────────────────────────────────────────────
			"stop_signal": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination signal (e.g. `SIGTERM`). Applies to: `deploy`.",
			},
			"stop_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds). Applies to: `deploy`.",
			},

			// ── stop / destroy ────────────────────────────────────────────────
			"signal": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination signal (e.g. `SIGTERM`). Applies to: `stop`, `destroy`.",
			},
			"time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Override the default termination max time (seconds). Applies to: `stop`, `destroy`.",
			},
		},
	}
}

func (a *DeploymentAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *DeploymentAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DeploymentActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	depID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified deployment action", map[string]interface{}{"deployment": depID, "action": act})

	switch act {
	case "deploy":
		execReq := client.DeployRequest{Deployment: depID}
		if !data.StopSignal.IsNull() && !data.StopSignal.IsUnknown() {
			execReq.StopSignal = data.StopSignal.ValueString()
		}
		if !data.StopTime.IsNull() && !data.StopTime.IsUnknown() {
			v := data.StopTime.ValueInt64()
			execReq.StopTime = &v
		}
		if err := a.client.Deploy(ctx, execReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy deployment, got error: %s", err))
		}
	case "start":
		if err := a.client.StartDeployment(ctx, client.StartDeploymentRequest{Deployment: depID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start deployment, got error: %s", err))
		}
	case "stop":
		execReq := client.StopDeploymentRequest{Deployment: depID}
		if !data.Signal.IsNull() && !data.Signal.IsUnknown() {
			execReq.Signal = data.Signal.ValueString()
		}
		if !data.Time.IsNull() && !data.Time.IsUnknown() {
			v := data.Time.ValueInt64()
			execReq.Time = &v
		}
		if err := a.client.StopDeployment(ctx, execReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop deployment, got error: %s", err))
		}
	case "restart":
		if err := a.client.RestartDeployment(ctx, client.RestartDeploymentRequest{Deployment: depID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to restart deployment, got error: %s", err))
		}
	case "pull":
		if err := a.client.PullDeployment(ctx, client.PullDeploymentRequest{Deployment: depID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull deployment, got error: %s", err))
		}
	case "pause":
		if err := a.client.PauseDeployment(ctx, client.PauseDeploymentRequest{Deployment: depID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pause deployment, got error: %s", err))
		}
	case "unpause":
		if err := a.client.UnpauseDeployment(ctx, client.UnpauseDeploymentRequest{Deployment: depID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unpause deployment, got error: %s", err))
		}
	case "destroy":
		execReq := client.DestroyDeploymentRequest{Deployment: depID}
		if !data.Signal.IsNull() && !data.Signal.IsUnknown() {
			execReq.Signal = data.Signal.ValueString()
		}
		if !data.Time.IsNull() && !data.Time.IsUnknown() {
			v := data.Time.ValueInt64()
			execReq.Time = &v
		}
		if err := a.client.DestroyDeployment(ctx, execReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to destroy deployment, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown deployment action %q. Valid values are: deploy, destroy, pause, unpause, pull, restart, start, stop.", act),
		)
	}
}
