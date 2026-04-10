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

// ─── RunBuild ────────────────────────────────────────────────────────────────

var _ action.Action = (*RunBuildAction)(nil)
var _ action.ActionWithConfigure = (*RunBuildAction)(nil)

func NewRunBuildAction() action.Action { return &RunBuildAction{} }

type RunBuildAction struct{ client *client.Client }

type RunBuildModel struct {
	Build types.String `tfsdk:"build"`
}

func (a *RunBuildAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_run_build"
}

func (a *RunBuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs the target build.",
		Attributes: map[string]schema.Attribute{
			"build": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the build to run.",
			},
		},
	}
}

func (a *RunBuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *RunBuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RunBuildModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing RunBuild", map[string]interface{}{"build": data.Build.ValueString()})
	if err := a.client.RunBuild(ctx, client.RunBuildRequest{Build: data.Build.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run build, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "RunBuild action completed")
}
