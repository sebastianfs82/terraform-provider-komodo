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

// ─── RunProcedure ────────────────────────────────────────────────────────────

var _ action.Action = (*RunProcedureAction)(nil)
var _ action.ActionWithConfigure = (*RunProcedureAction)(nil)

func NewRunProcedureAction() action.Action { return &RunProcedureAction{} }

type RunProcedureAction struct{ client *client.Client }

type RunProcedureModel struct {
	ID types.String `tfsdk:"id"`
}

func (a *RunProcedureAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_run_procedure"
}

func (a *RunProcedureAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs the target procedure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the procedure to run.",
			},
		},
	}
}

func (a *RunProcedureAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *RunProcedureAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RunProcedureModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing RunProcedure", map[string]interface{}{"procedure": data.ID.ValueString()})
	if err := a.client.RunProcedure(ctx, client.RunProcedureRequest{Procedure: data.ID.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run procedure, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RunProcedure action completed")
}
