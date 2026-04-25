// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ResourceSyncResource{}
var _ resource.ResourceWithImportState = &ResourceSyncResource{}

func NewResourceSyncResource() resource.Resource {
	return &ResourceSyncResource{}
}

type ResourceSyncResource struct {
	client *client.Client
}

// ResourceSyncSourceModel groups source configuration for a resource sync.
type ResourceSyncSourceModel struct {
	RepoID        types.String       `tfsdk:"repo_id"`
	URL           types.String       `tfsdk:"url"`
	AccountID     types.String       `tfsdk:"account_id"`
	Path          types.String       `tfsdk:"path"`
	Branch        types.String       `tfsdk:"branch"`
	Commit        types.String       `tfsdk:"commit"`
	FilesOnHost   types.Bool         `tfsdk:"on_host_enabled"`
	FileContents  TrimmedStringValue `tfsdk:"contents"`
	ResourcePaths types.List         `tfsdk:"resource_paths"`
}

// ResourceSyncManagedModeModel represents the managed_mode block.
type ResourceSyncManagedModeModel struct {
	Enabled   types.Bool `tfsdk:"enabled"`
	TagFilter types.List `tfsdk:"tag_filter"`
}

type ResourceSyncResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Tags types.List   `tfsdk:"tags"`

	// Git source
	Source *ResourceSyncSourceModel `tfsdk:"source"`

	// Webhook
	Webhook *WebhookModel `tfsdk:"webhook"`

	// Sync behaviour
	Delete        types.Bool                    `tfsdk:"delete_undeclared_resources_enabled"`
	Scope         types.List                    `tfsdk:"scope"`
	ManagedMode   *ResourceSyncManagedModeModel `tfsdk:"managed_mode"`
	AlertsEnabled types.Bool                    `tfsdk:"alerts_enabled"`
}

func (r *ResourceSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_sync"
}

func (r *ResourceSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo resource sync.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource sync identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the resource sync.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of tag IDs to attach to this resource. Use `komodo_tag.<name>.id` to reference tags.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},

			// Sync behaviour
			"delete_undeclared_resources_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should delete resources not declared in the resource files.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"scope": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Which entity types to include in the sync. Valid values: `\"resources\"`, `\"variables\"`, `\"user_groups\"`.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("resources", "variables", "user_groups"),
					),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"alerts_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should send an alert when it enters Pending state. Default: `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"source": schema.SingleNestedBlock{
				MarkdownDescription: "Git source configuration for repo-based resource syncs.",
				Attributes: map[string]schema.Attribute{
					"repo_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Id or name of a linked `komodo_repo` resource.",
					}, "url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The URL of the git provider, e.g. `https://github.com`.",
					}, "account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Id of a `komodo_provider_account` (git) used to access private repos. The git provider domain and HTTPS setting are derived automatically.",
					},
					"path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The git repository slug, e.g. `owner/repo`.",
					},
					"branch": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The branch to check out.",
					},
					"commit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "A specific commit hash to check out.",
					},
					"on_host_enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether resource files are available on the Komodo Core host. Specify paths with `resource_paths`.",
					},
					"contents": schema.StringAttribute{
						Optional:            true,
						CustomType:          TrimmedStringType{},
						MarkdownDescription: "Manage resource file contents directly in Terraform state (UI-managed mode).",
					},
					"resource_paths": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Path(s) to the resource file(s) to sync. Relative to `sync_directory` (files on host) or repo root (git-based).",
					},
				},
			}, "managed_mode": schema.SingleNestedBlock{
				MarkdownDescription: "Managed mode configuration: exports resources matching tags back into the sync file.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether managed mode is enabled.",
					},
					"tag_filter": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Only export resources matching all of these tags. Empty means all resources.",
					},
				},
			}, "webhook": schema.SingleNestedBlock{
				MarkdownDescription: "Webhook configuration for the sync.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether incoming webhooks trigger the sync.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Override the default webhook secret for this sync.",
					},
				},
			},
		},
	}
}

func (r *ResourceSyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *ResourceSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceSyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating resource sync", map[string]interface{}{"name": data.Name.ValueString()})
	createReq := client.CreateResourceSyncRequest{
		Name:   data.Name.ValueString(),
		Config: partialResourceSyncConfigFromModel(ctx, r.client, &data),
	}
	rs, err := r.client.CreateResourceSync(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create resource sync, got error: %s", err))
		return
	}
	if rs.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Resource sync creation failed: missing ID",
			"The Komodo API did not return a resource sync ID. Resource cannot be tracked in state.",
		)
		return
	}
	plannedTags := data.Tags
	resourceSyncToModel(ctx, r.client, rs, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "ResourceSync", ID: rs.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on resource sync, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	tflog.Trace(ctx, "Created resource sync")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceSyncResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rs, err := r.client.GetResourceSync(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read resource sync, got error: %s", err))
		return
	}
	if rs == nil {
		tflog.Debug(ctx, "Resource sync not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	resourceSyncToModel(ctx, r.client, rs, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceSyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state ResourceSyncResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameResourceSync(ctx, client.RenameResourceSyncRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename resource sync, got error: %s", err))
			return
		}
	}
	updateReq := client.UpdateResourceSyncRequest{
		ID:     data.ID.ValueString(),
		Config: partialResourceSyncConfigFromModel(ctx, r.client, &data),
	}
	rs, err := r.client.UpdateResourceSync(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update resource sync, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	resourceSyncToModel(ctx, r.client, rs, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "ResourceSync", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on resource sync, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceSyncResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting resource sync", map[string]interface{}{"id": data.ID.ValueString()})
	if err := r.client.DeleteResourceSync(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete resource sync, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted resource sync")
}

func (r *ResourceSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ── helpers ──────────────────────────────────────────────────────────────────

// resourceSyncToModel writes API response fields into the Terraform model.
func resourceSyncToModel(ctx context.Context, c *client.Client, rs *client.ResourceSync, m *ResourceSyncResourceModel) {
	m.ID = types.StringValue(rs.ID.OID)
	m.Name = types.StringValue(rs.Name)
	tagsSlice := rs.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	m.Tags = tags

	cfg := rs.Config

	strOrNull := func(s string) types.String {
		if s != "" {
			return types.StringValue(s)
		}
		return types.StringNull()
	}

	// Source block: set when any source field is non-empty or the prior state had one.
	if m.Source != nil || cfg.LinkedRepo != "" || cfg.Repo != "" || cfg.Branch != "" ||
		cfg.GitProvider != "" || cfg.GitAccount != "" || cfg.Commit != "" ||
		cfg.FilesOnHost || cfg.FileContents != "" || len(cfg.ResourcePath) > 0 {
		// Only populate branch if user set it (prior state non-null), to avoid
		// leaking API defaults (e.g. "main") into state.
		branchVal := func() types.String {
			if m.Source != nil && !m.Source.Branch.IsNull() {
				return strOrNull(cfg.Branch)
			}
			return types.StringNull()
		}()
		// Only populate on_host_enabled when user set it or it is true.
		filesOnHostVal := func() types.Bool {
			if cfg.FilesOnHost {
				return types.BoolValue(true)
			}
			if m.Source != nil && !m.Source.FilesOnHost.IsNull() {
				return types.BoolValue(false)
			}
			return types.BoolNull()
		}()
		// Only populate resource_paths when API returns paths or user had it set.
		var rpList types.List
		if len(cfg.ResourcePath) > 0 {
			rpList, _ = types.ListValueFrom(ctx, types.StringType, cfg.ResourcePath)
		} else if m.Source != nil && !m.Source.ResourcePaths.IsNull() {
			rpList, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		} else {
			rpList = types.ListNull(types.StringType)
		}
		if cfg.LinkedRepo != "" {
			// When repo_id is set, only store the repo ID; other fields are derived
			// by the API from the linked repo and must not be persisted to state.
			m.Source = &ResourceSyncSourceModel{
				RepoID:      types.StringValue(cfg.LinkedRepo),
				URL:         types.StringNull(),
				AccountID:   types.StringNull(),
				Path:        types.StringNull(),
				Branch:      types.StringNull(),
				Commit:      types.StringNull(),
				FilesOnHost: filesOnHostVal,
				FileContents: func() TrimmedStringValue {
					if cfg.FileContents != "" {
						return NewTrimmedStringValue(strings.TrimRight(cfg.FileContents, "\n"))
					}
					return NewTrimmedStringNull()
				}(),
				ResourcePaths: rpList,
			}
		} else {
			var urlVal types.String
			priorURLSet := m.Source != nil && !m.Source.URL.IsNull()
			if cfg.GitProvider != "" && priorURLSet {
				if cfg.GitHttps {
					urlVal = types.StringValue("https://" + cfg.GitProvider)
				} else {
					urlVal = types.StringValue("http://" + cfg.GitProvider)
				}
			} else {
				urlVal = types.StringNull()
			}
			accountID := c.ResolveGitAccountID(ctx, cfg.GitProvider, cfg.GitAccount)
			m.Source = &ResourceSyncSourceModel{
				RepoID:      types.StringNull(),
				URL:         urlVal,
				AccountID:   strOrNull(accountID),
				Path:        strOrNull(cfg.Repo),
				Branch:      branchVal,
				Commit:      strOrNull(cfg.Commit),
				FilesOnHost: filesOnHostVal,
				FileContents: func() TrimmedStringValue {
					if cfg.FileContents != "" {
						return NewTrimmedStringValue(strings.TrimRight(cfg.FileContents, "\n"))
					}
					if m.Source != nil && !m.Source.FileContents.IsNull() {
						return NewTrimmedStringValue("")
					}
					return NewTrimmedStringNull()
				}(),
				ResourcePaths: rpList,
			}
		}
	} else {
		m.Source = nil
	}
	webhookSecret := types.StringNull()
	if cfg.WebhookSecret != "" {
		webhookSecret = types.StringValue(cfg.WebhookSecret)
	}
	if cfg.WebhookEnabled || cfg.WebhookSecret != "" {
		m.Webhook = &WebhookModel{
			Enabled: types.BoolValue(cfg.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		m.Webhook = nil
	}
	m.Delete = types.BoolValue(cfg.Delete)
	m.AlertsEnabled = types.BoolValue(cfg.PendingAlert)

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
	m.Scope, _ = types.ListValueFrom(ctx, types.StringType, scopeItems)

	// managed_mode block: populate when managed is set or there are match tags.
	if cfg.Managed || len(cfg.MatchTags) > 0 || m.ManagedMode != nil {
		var tagFilter types.List
		if tags := cfg.MatchTags; len(tags) > 0 {
			tagFilter, _ = types.ListValueFrom(ctx, types.StringType, tags)
		} else {
			tagFilter, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		}
		m.ManagedMode = &ResourceSyncManagedModeModel{
			Enabled:   types.BoolValue(cfg.Managed),
			TagFilter: tagFilter,
		}
	} else {
		m.ManagedMode = nil
	}
}

// partialResourceSyncConfigFromModel converts Terraform model into PartialResourceSyncConfig.
func partialResourceSyncConfigFromModel(ctx context.Context, c *client.Client, m *ResourceSyncResourceModel) client.PartialResourceSyncConfig {
	cfg := client.PartialResourceSyncConfig{}

	if m.Source != nil {
		if !m.Source.RepoID.IsNull() && !m.Source.RepoID.IsUnknown() {
			v := m.Source.RepoID.ValueString()
			cfg.LinkedRepo = &v
		}
		// When not using a repo_id, derive git_provider and git_https from url or account.
		if m.Source.RepoID.IsNull() || m.Source.RepoID.ValueString() == "" {
			url := m.Source.URL.ValueString()
			if strings.HasPrefix(url, "https://") {
				domain := strings.TrimPrefix(url, "https://")
				https := true
				cfg.GitProvider = &domain
				cfg.GitHttps = &https
			} else if strings.HasPrefix(url, "http://") {
				domain := strings.TrimPrefix(url, "http://")
				https := false
				cfg.GitProvider = &domain
				cfg.GitHttps = &https
			} else if url != "" {
				https := true
				cfg.GitProvider = &url
				cfg.GitHttps = &https
			} else if !m.Source.AccountID.IsNull() && !m.Source.AccountID.IsUnknown() {
				if acc := c.ResolveGitAccountFull(ctx, m.Source.AccountID.ValueString()); acc != nil {
					domain := acc.Domain
					https := true
					cfg.GitProvider = &domain
					cfg.GitHttps = &https
				}
			}
			if !m.Source.AccountID.IsNull() && !m.Source.AccountID.IsUnknown() {
				accID := m.Source.AccountID.ValueString()
				username, err := c.ResolveGitAccountUsername(ctx, accID)
				if err != nil {
					username = accID
				}
				cfg.GitAccount = &username
			}
		}
		if !m.Source.Path.IsNull() && !m.Source.Path.IsUnknown() {
			v := m.Source.Path.ValueString()
			cfg.Repo = &v
		}
		if !m.Source.Branch.IsNull() && !m.Source.Branch.IsUnknown() {
			v := m.Source.Branch.ValueString()
			cfg.Branch = &v
		}
		if !m.Source.Commit.IsNull() && !m.Source.Commit.IsUnknown() {
			v := m.Source.Commit.ValueString()
			cfg.Commit = &v
		}
		if !m.Source.FilesOnHost.IsNull() && !m.Source.FilesOnHost.IsUnknown() {
			v := m.Source.FilesOnHost.ValueBool()
			cfg.FilesOnHost = &v
		}
		if !m.Source.FileContents.IsNull() && !m.Source.FileContents.IsUnknown() {
			v := m.Source.FileContents.ValueString()
			cfg.FileContents = &v
		}
		if !m.Source.ResourcePaths.IsNull() && !m.Source.ResourcePaths.IsUnknown() {
			var paths []string
			_ = m.Source.ResourcePaths.ElementsAs(ctx, &paths, false)
			cfg.ResourcePath = paths
		}
	}
	if m.Webhook != nil {
		if !m.Webhook.Enabled.IsNull() && !m.Webhook.Enabled.IsUnknown() {
			v := m.Webhook.Enabled.ValueBool()
			cfg.WebhookEnabled = &v
		}
		if !m.Webhook.Secret.IsNull() && !m.Webhook.Secret.IsUnknown() {
			v := m.Webhook.Secret.ValueString()
			cfg.WebhookSecret = &v
		}
	} else {
		f, s := false, ""
		cfg.WebhookEnabled = &f
		cfg.WebhookSecret = &s
	}
	if !m.Delete.IsNull() && !m.Delete.IsUnknown() {
		v := m.Delete.ValueBool()
		cfg.Delete = &v
	}
	// Expand scope list into the three include booleans.
	if !m.Scope.IsNull() && !m.Scope.IsUnknown() {
		var scopeItems []string
		_ = m.Scope.ElementsAs(ctx, &scopeItems, false)
		scopeSet := make(map[string]bool, len(scopeItems))
		for _, s := range scopeItems {
			scopeSet[s] = true
		}
		incRes := scopeSet["resources"]
		incVars := scopeSet["variables"]
		incGroups := scopeSet["user_groups"]
		cfg.IncludeResources = &incRes
		cfg.IncludeVariables = &incVars
		cfg.IncludeUserGroups = &incGroups
	}
	if m.ManagedMode != nil {
		if !m.ManagedMode.Enabled.IsNull() && !m.ManagedMode.Enabled.IsUnknown() {
			v := m.ManagedMode.Enabled.ValueBool()
			cfg.Managed = &v
		}
		if !m.ManagedMode.TagFilter.IsNull() && !m.ManagedMode.TagFilter.IsUnknown() {
			var tags []string
			_ = m.ManagedMode.TagFilter.ElementsAs(ctx, &tags, false)
			cfg.MatchTags = tags
		}
	}
	if !m.AlertsEnabled.IsNull() && !m.AlertsEnabled.IsUnknown() {
		v := m.AlertsEnabled.ValueBool()
		cfg.PendingAlert = &v
	}

	return cfg
}
