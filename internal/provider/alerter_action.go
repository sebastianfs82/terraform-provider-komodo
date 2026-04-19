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

// ─── AlerterAction (unified) ──────────────────────────────────────────────────

var _ action.Action = (*AlerterAction)(nil)
var _ action.ActionWithConfigure = (*AlerterAction)(nil)

func NewAlerterAction() action.Action { return &AlerterAction{} }

type AlerterAction struct{ client *client.Client }

type AlerterActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *AlerterAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerter"
}

func (a *AlerterAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo alerter resource. " +
			"Set `action` to select the operation: `test`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target alerter.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The alerter operation to perform. Valid values: `test`.",
			},
		},
	}
}

func (a *AlerterAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *AlerterAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data AlerterActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alerterID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified alerter action", map[string]interface{}{"alerter": alerterID, "action": act})

	switch act {
	case "test":
		if err := a.client.TestAlerter(ctx, client.TestAlerterRequest{Alerter: alerterID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to test alerter, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown alerter action %q. Valid values are: test.", act),
		)
	}
}
