// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type ResourceSyncResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Tags types.List   `tfsdk:"tags"`

	// Git source
	LinkedRepo  types.String `tfsdk:"linked_repo"`
	GitProvider types.String `tfsdk:"git_provider"`
	GitHttps    types.Bool   `tfsdk:"git_https"`
	Repo        types.String `tfsdk:"repo"`
	Branch      types.String `tfsdk:"branch"`
	Commit      types.String `tfsdk:"commit"`
	GitAccount  types.String `tfsdk:"git_account"`

	// Files
	FilesOnHost  types.Bool   `tfsdk:"files_on_host"`
	ResourcePath types.List   `tfsdk:"resource_path"`
	FileContents types.String `tfsdk:"file_contents"`

	// Webhook
	Webhook *WebhookModel `tfsdk:"webhook"`

	// Sync behaviour
	Managed           types.Bool `tfsdk:"managed"`
	Delete            types.Bool `tfsdk:"delete"`
	IncludeResources  types.Bool `tfsdk:"include_resources"`
	MatchTags         types.List `tfsdk:"match_tags"`
	IncludeVariables  types.Bool `tfsdk:"include_variables"`
	IncludeUserGroups types.Bool `tfsdk:"include_user_groups"`
	PendingAlert      types.Bool `tfsdk:"pending_alert"`
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

			// Git source
			"linked_repo": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Id or name of a Komodo Repo resource to source the sync files from.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_provider": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The git provider domain. Default: `github.com`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_https": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use HTTPS to clone the repo (versus HTTP). Default: `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"repo": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The git repository slug, e.g. `owner/repo`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The branch to check out.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"commit": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A specific commit hash to check out.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_account": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The git account used to access private repos.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Files
			"files_on_host": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether resource files are available on the Komodo Core host. Specify paths with `resource_path`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_path": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Path(s) to the resource file(s) to sync. Relative to `sync_directory` (files on host) or repo root (git-based).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"file_contents": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Manage resource file contents directly in Terraform state (UI-managed mode).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Sync behaviour
			"managed": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable managed mode: resource exports matching tags are pushed to the sync file.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should delete resources not declared in the resource files.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"include_resources": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should include resources. Default: `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"match_tags": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "In managed mode, only export resources matching all of these tags. Empty means all resources.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"include_variables": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should include variables.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"include_user_groups": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should include user groups.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"pending_alert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the sync should send an alert when it enters Pending state. Default: `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"webhook": schema.SingleNestedBlock{
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
		Config: partialResourceSyncConfigFromModel(ctx, &data),
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
	resourceSyncToModel(ctx, rs, &data)
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
	resourceSyncToModel(ctx, rs, &data)
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
		Config: partialResourceSyncConfigFromModel(ctx, &data),
	}
	rs, err := r.client.UpdateResourceSync(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update resource sync, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	resourceSyncToModel(ctx, rs, &data)
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
func resourceSyncToModel(ctx context.Context, rs *client.ResourceSync, m *ResourceSyncResourceModel) {
	m.ID = types.StringValue(rs.ID.OID)
	m.Name = types.StringValue(rs.Name)
	tagsSlice := rs.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	m.Tags = tags

	cfg := rs.Config
	m.LinkedRepo = types.StringValue(cfg.LinkedRepo)
	m.GitProvider = types.StringValue(cfg.GitProvider)
	m.GitHttps = types.BoolValue(cfg.GitHttps)
	m.Repo = types.StringValue(cfg.Repo)
	m.Branch = types.StringValue(cfg.Branch)
	m.Commit = types.StringValue(cfg.Commit)
	m.GitAccount = types.StringValue(cfg.GitAccount)
	m.FilesOnHost = types.BoolValue(cfg.FilesOnHost)
	m.FileContents = types.StringValue(cfg.FileContents)
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
	m.Managed = types.BoolValue(cfg.Managed)
	m.Delete = types.BoolValue(cfg.Delete)
	m.IncludeResources = types.BoolValue(cfg.IncludeResources)
	m.IncludeVariables = types.BoolValue(cfg.IncludeVariables)
	m.IncludeUserGroups = types.BoolValue(cfg.IncludeUserGroups)
	m.PendingAlert = types.BoolValue(cfg.PendingAlert)

	if paths := cfg.ResourcePath; len(paths) > 0 {
		elems := make([]types.String, len(paths))
		for i, p := range paths {
			elems[i] = types.StringValue(p)
		}
		listVal, _ := types.ListValueFrom(ctx, types.StringType, paths)
		m.ResourcePath = listVal
	} else {
		m.ResourcePath, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	if tags := cfg.MatchTags; len(tags) > 0 {
		listVal, _ := types.ListValueFrom(ctx, types.StringType, tags)
		m.MatchTags = listVal
	} else {
		m.MatchTags, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}
}

// partialResourceSyncConfigFromModel converts Terraform model into PartialResourceSyncConfig.
func partialResourceSyncConfigFromModel(ctx context.Context, m *ResourceSyncResourceModel) client.PartialResourceSyncConfig {
	cfg := client.PartialResourceSyncConfig{}

	if !m.LinkedRepo.IsNull() && !m.LinkedRepo.IsUnknown() {
		v := m.LinkedRepo.ValueString()
		cfg.LinkedRepo = &v
	}
	if !m.GitProvider.IsNull() && !m.GitProvider.IsUnknown() {
		v := m.GitProvider.ValueString()
		cfg.GitProvider = &v
	}
	if !m.GitHttps.IsNull() && !m.GitHttps.IsUnknown() {
		v := m.GitHttps.ValueBool()
		cfg.GitHttps = &v
	}
	if !m.Repo.IsNull() && !m.Repo.IsUnknown() {
		v := m.Repo.ValueString()
		cfg.Repo = &v
	}
	if !m.Branch.IsNull() && !m.Branch.IsUnknown() {
		v := m.Branch.ValueString()
		cfg.Branch = &v
	}
	if !m.Commit.IsNull() && !m.Commit.IsUnknown() {
		v := m.Commit.ValueString()
		cfg.Commit = &v
	}
	if !m.GitAccount.IsNull() && !m.GitAccount.IsUnknown() {
		v := m.GitAccount.ValueString()
		cfg.GitAccount = &v
	}
	if !m.FilesOnHost.IsNull() && !m.FilesOnHost.IsUnknown() {
		v := m.FilesOnHost.ValueBool()
		cfg.FilesOnHost = &v
	}
	if !m.ResourcePath.IsNull() && !m.ResourcePath.IsUnknown() {
		var paths []string
		_ = m.ResourcePath.ElementsAs(ctx, &paths, false)
		cfg.ResourcePath = paths
	}
	if !m.FileContents.IsNull() && !m.FileContents.IsUnknown() {
		v := m.FileContents.ValueString()
		cfg.FileContents = &v
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
	if !m.Managed.IsNull() && !m.Managed.IsUnknown() {
		v := m.Managed.ValueBool()
		cfg.Managed = &v
	}
	if !m.Delete.IsNull() && !m.Delete.IsUnknown() {
		v := m.Delete.ValueBool()
		cfg.Delete = &v
	}
	if !m.IncludeResources.IsNull() && !m.IncludeResources.IsUnknown() {
		v := m.IncludeResources.ValueBool()
		cfg.IncludeResources = &v
	}
	if !m.MatchTags.IsNull() && !m.MatchTags.IsUnknown() {
		var tags []string
		_ = m.MatchTags.ElementsAs(ctx, &tags, false)
		cfg.MatchTags = tags
	}
	if !m.IncludeVariables.IsNull() && !m.IncludeVariables.IsUnknown() {
		v := m.IncludeVariables.ValueBool()
		cfg.IncludeVariables = &v
	}
	if !m.IncludeUserGroups.IsNull() && !m.IncludeUserGroups.IsUnknown() {
		v := m.IncludeUserGroups.ValueBool()
		cfg.IncludeUserGroups = &v
	}
	if !m.PendingAlert.IsNull() && !m.PendingAlert.IsUnknown() {
		v := m.PendingAlert.ValueBool()
		cfg.PendingAlert = &v
	}

	return cfg
}
