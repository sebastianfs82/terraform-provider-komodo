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

// ─── TestAlerter ─────────────────────────────────────────────────────────────

var _ action.Action = (*TestAlerterAction)(nil)
var _ action.ActionWithConfigure = (*TestAlerterAction)(nil)

func NewTestAlerterAction() action.Action { return &TestAlerterAction{} }

type TestAlerterAction struct{ client *client.Client }

type TestAlerterModel struct {
	Alerter types.String `tfsdk:"alerter"`
}

func (a *TestAlerterAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_alerter"
}

func (a *TestAlerterAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Tests an alerter's ability to reach the configured endpoint.",
		Attributes: map[string]schema.Attribute{
			"alerter": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the alerter to test.",
			},
		},
	}
}

func (a *TestAlerterAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *TestAlerterAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data TestAlerterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing TestAlerter", map[string]interface{}{"alerter": data.Alerter.ValueString()})
	if err := a.client.TestAlerter(ctx, client.TestAlerterRequest{Alerter: data.Alerter.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to test alerter, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "TestAlerter action completed")
}
