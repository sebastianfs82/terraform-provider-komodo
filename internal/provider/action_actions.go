// Copyright (c) HashiCorp, Inc.
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

// ─── RunAction ───────────────────────────────────────────────────────────────

var _ action.Action = (*RunActionAction)(nil)
var _ action.ActionWithConfigure = (*RunActionAction)(nil)

func NewRunActionAction() action.Action { return &RunActionAction{} }

type RunActionAction struct{ client *client.Client }

type RunActionModel struct {
	Action types.String `tfsdk:"action"`
}

func (a *RunActionAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_run_action"
}

func (a *RunActionAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs the target action.",
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the action to run.",
			},
		},
	}
}

func (a *RunActionAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *RunActionAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RunActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing RunAction", map[string]interface{}{"action": data.Action.ValueString()})
	if err := a.client.RunAction(ctx, client.RunActionRequest{Action: data.Action.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run action, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RunAction action completed")
}
