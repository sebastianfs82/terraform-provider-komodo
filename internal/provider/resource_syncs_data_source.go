// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &ResourceSyncsDataSource{}

func NewResourceSyncsDataSource() datasource.DataSource {
	return &ResourceSyncsDataSource{}
}

type ResourceSyncsDataSource struct {
	client *client.Client
}

type ResourceSyncsDataSourceModel struct {
	RepoID        types.String                `tfsdk:"repo_id"`
	ResourceSyncs []ResourceSyncListItemModel `tfsdk:"resource_syncs"`
}

type ResourceSyncListItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Repo           types.String `tfsdk:"repo"`
	Branch         types.String `tfsdk:"branch"`
	WebhookEnabled types.Bool   `tfsdk:"webhook_enabled"`
	Managed        types.Bool   `tfsdk:"managed_mode_enabled"`
}

func (d *ResourceSyncsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_syncs"
}

func (d *ResourceSyncsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo resource syncs visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"repo_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter resource syncs by linked Komodo Repo name or ID.",
			},
			"resource_syncs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of resource syncs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The resource sync identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the resource sync.",
						},
						"repo": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git repository path (owner/repo).",
						},
						"branch": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git branch.",
						},
						"webhook_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether webhook triggers are enabled.",
						},
						"managed_mode_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the sync manages resources (creates/deletes).",
						},
					},
				},
			},
		},
	}
}

func (d *ResourceSyncsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *ResourceSyncsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourceSyncsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing resource syncs")

	syncs, err := d.client.ListFullResourceSyncs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list resource syncs, got error: %s", err))
		return
	}

	repoIDFilter := data.RepoID.ValueString()
	// linked_repo stores the repo name, not the ID. Resolve the filter value to
	// the repo name so we can match against what the API returns.
	repoNameFilter := ""
	if repoIDFilter != "" {
		if repos, err := d.client.ListGitRepositories(ctx); err == nil {
			for _, r := range repos {
				if r.ID.OID == repoIDFilter {
					repoNameFilter = r.Name
					break
				}
			}
		}
		if repoNameFilter == "" {
			repoNameFilter = repoIDFilter
		}
	}

	items := make([]ResourceSyncListItemModel, 0, len(syncs))
	for _, s := range syncs {
		if repoIDFilter != "" && s.Config.LinkedRepo != repoIDFilter && s.Config.LinkedRepo != repoNameFilter {
			continue
		}
		items = append(items, ResourceSyncListItemModel{
			ID:             types.StringValue(s.ID.OID),
			Name:           types.StringValue(s.Name),
			Repo:           types.StringValue(s.Config.Repo),
			Branch:         types.StringValue(s.Config.Branch),
			WebhookEnabled: types.BoolValue(s.Config.WebhookEnabled),
			Managed:        types.BoolValue(s.Config.Managed),
		})
	}
	data.ResourceSyncs = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
