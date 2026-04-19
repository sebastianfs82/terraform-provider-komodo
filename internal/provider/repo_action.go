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

// ─── RepoAction (unified) ─────────────────────────────────────────────────────

var _ action.Action = (*RepoAction)(nil)
var _ action.ActionWithConfigure = (*RepoAction)(nil)

func NewRepoAction() action.Action { return &RepoAction{} }

type RepoAction struct{ client *client.Client }

type RepoActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *RepoAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo"
}

func (a *RepoAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo repo resource. " +
			"Set `action` to select the operation: `build`, `clone`, or `pull`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target repo.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The repo operation to perform. Valid values: `build`, `clone`, `pull`.",
			},
		},
	}
}

func (a *RepoAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = c
}

func (a *RepoAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RepoActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified repo action", map[string]interface{}{"repo": repoID, "action": act})

	switch act {
	case "build":
		if err := a.client.BuildRepo(ctx, client.BuildRepoRequest{Repo: repoID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to build repo, got error: %s", err))
		}
	case "clone":
		if err := a.client.CloneRepo(ctx, client.CloneRepoRequest{Repo: repoID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clone repo, got error: %s", err))
		}
	case "pull":
		if err := a.client.PullRepo(ctx, client.PullRepoRequest{Repo: repoID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull repo, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown repo action %q. Valid values are: build, clone, pull.", act),
		)
	}
}
