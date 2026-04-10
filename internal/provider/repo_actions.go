// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

// ─── BuildRepo ───────────────────────────────────────────────────────────────

var _ action.Action = (*RepoBuildAction)(nil)
var _ action.ActionWithConfigure = (*RepoBuildAction)(nil)

func NewRepoBuildAction() action.Action { return &RepoBuildAction{} }

type RepoBuildAction struct{ client *client.Client }

type RepoBuildModel struct {
	Repo types.String `tfsdk:"repo"`
}

func (a *RepoBuildAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_build"
}

func (a *RepoBuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Builds the target repo using its attached builder.",
		Attributes: map[string]schema.Attribute{
			"repo": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the repo to build.",
			},
		},
	}
}

func (a *RepoBuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *RepoBuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RepoBuildModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.BuildRepoRequest{Repo: data.Repo.ValueString()}

	tflog.Debug(ctx, "Executing BuildRepo", map[string]interface{}{"repo": execReq.Repo})
	if err := a.client.BuildRepo(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to build repo, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "BuildRepo action completed")
}

// ─── CloneRepo ───────────────────────────────────────────────────────────────

var _ action.Action = (*RepoCloneAction)(nil)
var _ action.ActionWithConfigure = (*RepoCloneAction)(nil)

func NewRepoCloneAction() action.Action { return &RepoCloneAction{} }

type RepoCloneAction struct{ client *client.Client }

type RepoCloneModel struct {
	Repo types.String `tfsdk:"repo"`
}

func (a *RepoCloneAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_clone"
}

func (a *RepoCloneAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Clones the target repo onto its attached server.",
		Attributes: map[string]schema.Attribute{
			"repo": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the repo to clone.",
			},
		},
	}
}

func (a *RepoCloneAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *RepoCloneAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RepoCloneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.CloneRepoRequest{Repo: data.Repo.ValueString()}

	tflog.Debug(ctx, "Executing CloneRepo", map[string]interface{}{"repo": execReq.Repo})
	if err := a.client.CloneRepo(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clone repo, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "CloneRepo action completed")
}

// ─── PullRepo ────────────────────────────────────────────────────────────────

var _ action.Action = (*RepoPullAction)(nil)
var _ action.ActionWithConfigure = (*RepoPullAction)(nil)

func NewRepoPullAction() action.Action { return &RepoPullAction{} }

type RepoPullAction struct{ client *client.Client }

type RepoPullModel struct {
	Repo types.String `tfsdk:"repo"`
}

func (a *RepoPullAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_pull"
}

func (a *RepoPullAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pulls the latest changes for the target repo on its attached server.",
		Attributes: map[string]schema.Attribute{
			"repo": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the repo to pull.",
			},
		},
	}
}

func (a *RepoPullAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *RepoPullAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RepoPullModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	execReq := client.PullRepoRequest{Repo: data.Repo.ValueString()}

	tflog.Debug(ctx, "Executing PullRepo", map[string]interface{}{"repo": execReq.Repo})
	if err := a.client.PullRepo(ctx, execReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to pull repo, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PullRepo action completed")
}
