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

// ─── RunSync ─────────────────────────────────────────────────────────────────

var _ action.Action = (*RunSyncAction)(nil)
var _ action.ActionWithConfigure = (*RunSyncAction)(nil)

func NewRunSyncAction() action.Action { return &RunSyncAction{} }

type RunSyncAction struct{ client *client.Client }

type RunSyncModel struct {
	Sync types.String `tfsdk:"sync"`
}

func (a *RunSyncAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_run_sync"
}

func (a *RunSyncAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs the target resource sync.",
		Attributes: map[string]schema.Attribute{
			"sync": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the resource sync to run.",
			},
		},
	}
}

func (a *RunSyncAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *RunSyncAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RunSyncModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing RunSync", map[string]interface{}{"sync": data.Sync.ValueString()})
	if err := a.client.RunSync(ctx, client.RunSyncRequest{Sync: data.Sync.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run sync, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RunSync action completed")
}
