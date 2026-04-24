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

// ─── ResourceSyncAction (unified) ─────────────────────────────────────────────

var _ action.Action = (*ResourceSyncAction)(nil)
var _ action.ActionWithConfigure = (*ResourceSyncAction)(nil)

func NewResourceSyncAction() action.Action { return &ResourceSyncAction{} }

type ResourceSyncAction struct{ client *client.Client }

type ResourceSyncActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *ResourceSyncAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_sync"
}

func (a *ResourceSyncAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo resource sync. " +
			"Set `action` to select the operation: `run`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target resource sync.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The resource sync operation to perform. Valid values: `run`.",
			},
		},
	}
}

func (a *ResourceSyncAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ResourceSyncAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ResourceSyncActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	syncID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified resource sync action", map[string]interface{}{"sync": syncID, "action": act})

	switch act {
	case "run":
		if err := a.client.RunSync(ctx, client.RunSyncRequest{Sync: syncID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run resource sync, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown resource sync action %q. Valid values are: run.", act),
		)
	}
}
