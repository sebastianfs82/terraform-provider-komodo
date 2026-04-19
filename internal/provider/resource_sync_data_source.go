// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &ResourceSyncDataSource{}
var _ datasource.DataSourceWithValidateConfig = &ResourceSyncDataSource{}

func NewResourceSyncDataSource() datasource.DataSource {
	return &ResourceSyncDataSource{}
}

type ResourceSyncDataSource struct {
	client *client.Client
}

type ResourceSyncDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	// Git source
	LinkedRepo  types.String `tfsdk:"linked_repo"`
	GitProvider types.String `tfsdk:"git_provider"`
	GitHttps    types.Bool   `tfsdk:"git_https"`
	Repo        types.String `tfsdk:"repo"`
	Branch      types.String `tfsdk:"branch"`
	Commit      types.String `tfsdk:"commit"`
	GitAccount  types.String `tfsdk:"git_account"`

	// Files
	FilesOnHost   types.Bool   `tfsdk:"on_host_enabled"`
	ResourcePaths types.List   `tfsdk:"resource_paths"`
	FileContents  types.String `tfsdk:"contents"`

	// Webhook
	Webhook *WebhookModel `tfsdk:"webhook"`

	// Sync behaviour
	Delete        types.Bool                    `tfsdk:"delete_undeclared_resources_enabled"`
	Scope         types.List                    `tfsdk:"scope"`
	ManagedMode   *ResourceSyncManagedModeModel `tfsdk:"managed_mode"`
	AlertsEnabled types.Bool                    `tfsdk:"alerts_enabled"`
}

func (d *ResourceSyncDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_sync"
}

func (d *ResourceSyncDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo resource sync.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The resource sync identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique name of the resource sync. One of `name` or `id` must be set.",
			},

			// Git source
			"linked_repo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Id or name of the linked Komodo Repo.",
			},
			"git_provider": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git provider domain.",
			},
			"git_https": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether HTTPS is used to clone the repo.",
			},
			"repo": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git repository slug.",
			},
			"branch": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The branch checked out.",
			},
			"commit": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The specific commit hash checked out.",
			},
			"git_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git account used to access private repos.",
			},

			// Files
			"on_host_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether resource files are on the Komodo Core host.",
			},
			"resource_paths": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Path(s) to the resource file(s) to sync.",
			},
			"contents": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UI-managed resource file contents.",
			},

			// Webhook
			"webhook": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook configuration for the sync.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether incoming webhooks trigger the sync.",
					},
					"secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "The webhook secret override for this sync.",
					},
				},
			},

			// Sync behaviour
			"delete_undeclared_resources_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the sync deletes undeclared resources.",
			},
			"scope": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Which entity types are included in the sync. Values: `\"resources\"`, `\"variables\"`, `\"user_groups\"`.",
			},
			"managed_mode": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Managed mode configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether managed mode is enabled.",
					},
					"tag_filter": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Tags used to filter resource exports in managed mode.",
					},
				},
			},
			"alerts_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether an alert is sent when the sync enters Pending state.",
			},
		},
	}
}

func (d *ResourceSyncDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ResourceSyncDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data ResourceSyncDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Name.IsUnknown() || data.ID.IsUnknown() {
		return
	}
	nameSet := !data.Name.IsNull()
	idSet := !data.ID.IsNull()
	if nameSet && idSet {
		resp.Diagnostics.AddError("Invalid Configuration", "Only one of `name` or `id` may be set, not both.")
		return
	}
	if !nameSet && !idSet {
		resp.Diagnostics.AddError("Invalid Configuration", "One of `name` or `id` must be set.")
	}
}

func (d *ResourceSyncDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourceSyncDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading resource sync", map[string]interface{}{"lookup": lookup})
	rs, err := d.client.GetResourceSync(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read resource sync, got error: %s", err))
		return
	}
	if rs == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Resource sync %q not found.", lookup))
		return
	}

	data.ID = types.StringValue(rs.ID.OID)
	data.Name = types.StringValue(rs.Name)
	cfg := rs.Config
	data.LinkedRepo = types.StringValue(cfg.LinkedRepo)
	data.GitProvider = types.StringValue(cfg.GitProvider)
	data.GitHttps = types.BoolValue(cfg.GitHttps)
	data.Repo = types.StringValue(cfg.Repo)
	data.Branch = types.StringValue(cfg.Branch)
	data.Commit = types.StringValue(cfg.Commit)
	data.GitAccount = types.StringValue(cfg.GitAccount)
	data.FilesOnHost = types.BoolValue(cfg.FilesOnHost)
	data.FileContents = types.StringValue(strings.TrimRight(cfg.FileContents, "\n"))
	webhookSecret := types.StringNull()
	if cfg.WebhookSecret != "" {
		webhookSecret = types.StringValue(cfg.WebhookSecret)
	}
	if cfg.WebhookEnabled || cfg.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(cfg.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
	data.Delete = types.BoolValue(cfg.Delete)
	data.AlertsEnabled = types.BoolValue(cfg.PendingAlert)

	// Build scope list from the three include booleans.
	var scopeItems []string
	if cfg.IncludeResources {
		scopeItems = append(scopeItems, "resources")
	}
	if cfg.IncludeVariables {
		scopeItems = append(scopeItems, "variables")
	}
	if cfg.IncludeUserGroups {
		scopeItems = append(scopeItems, "user_groups")
	}
	data.Scope, _ = types.ListValueFrom(ctx, types.StringType, scopeItems)

	// managed_mode nested attribute.
	var tagFilter types.List
	if tags := cfg.MatchTags; len(tags) > 0 {
		tagFilter, _ = types.ListValueFrom(ctx, types.StringType, tags)
	} else {
		tagFilter, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	data.ManagedMode = &ResourceSyncManagedModeModel{
		Enabled:   types.BoolValue(cfg.Managed),
		TagFilter: tagFilter,
	}

	if paths := cfg.ResourcePath; len(paths) > 0 {
		listVal, _ := types.ListValueFrom(ctx, types.StringType, paths)
		data.ResourcePaths = listVal
	} else {
		data.ResourcePaths, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
