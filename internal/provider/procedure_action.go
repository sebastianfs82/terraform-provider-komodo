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

// ─── ProcedureAction (unified) ────────────────────────────────────────────────

var _ action.Action = (*ProcedureAction)(nil)
var _ action.ActionWithConfigure = (*ProcedureAction)(nil)

func NewProcedureAction() action.Action { return &ProcedureAction{} }

type ProcedureAction struct{ client *client.Client }

type ProcedureActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *ProcedureAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (a *ProcedureAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo procedure resource. " +
			"Set `action` to select the operation: `run`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target procedure.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The procedure operation to perform. Valid values: `run`.",
			},
		},
	}
}

func (a *ProcedureAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ProcedureAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ProcedureActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	procID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified procedure action", map[string]interface{}{"procedure": procID, "action": act})

	switch act {
	case "run":
		if err := a.client.RunProcedure(ctx, client.RunProcedureRequest{Procedure: procID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run procedure, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown procedure action %q. Valid values are: run.", act),
		)
	}
}
