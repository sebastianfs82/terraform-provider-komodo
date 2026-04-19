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

// ─── BuildAction (unified) ────────────────────────────────────────────────────

var _ action.Action = (*BuildAction)(nil)
var _ action.ActionWithConfigure = (*BuildAction)(nil)

func NewBuildAction() action.Action { return &BuildAction{} }

type BuildAction struct{ client *client.Client }

type BuildActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *BuildAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (a *BuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo build resource. " +
			"Set `action` to select the operation: `run`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target build.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The build operation to perform. Valid values: `run`.",
			},
		},
	}
}

func (a *BuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *BuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data BuildActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified build action", map[string]interface{}{"build": buildID, "action": act})

	switch act {
	case "run":
		if err := a.client.RunBuild(ctx, client.RunBuildRequest{Build: buildID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run build, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown build action %q. Valid values are: run.", act),
		)
	}
}
