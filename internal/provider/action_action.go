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

// ─── ActionAction (unified) ───────────────────────────────────────────────────

var _ action.Action = (*KomodoActionAction)(nil)
var _ action.ActionWithConfigure = (*KomodoActionAction)(nil)

func NewKomodoActionAction() action.Action { return &KomodoActionAction{} }

type KomodoActionAction struct{ client *client.Client }

type KomodoActionActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *KomodoActionAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (a *KomodoActionAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo action resource. " +
			"Set `action` to select the operation: `run`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target action.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The action operation to perform. Valid values: `run`.",
			},
		},
	}
}

func (a *KomodoActionAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *KomodoActionAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data KomodoActionActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	actionID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified action action", map[string]interface{}{"action": actionID, "op": act})

	switch act {
	case "run":
		if err := a.client.RunAction(ctx, client.RunActionRequest{Action: actionID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run action, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown action operation %q. Valid values are: run.", act),
		)
	}
}
