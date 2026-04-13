// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &StackResource{}
var _ resource.ResourceWithImportState = &StackResource{}
var _ resource.ResourceWithConfigValidators = &StackResource{}

func NewStackResource() resource.Resource {
	return &StackResource{}
}

type StackResource struct {
	client *client.Client
}

// StackWebhookModel extends the base webhook model with force_deploy.
type StackWebhookModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Secret      types.String `tfsdk:"secret"`
	ForceDeploy types.Bool   `tfsdk:"force_deploy"`
}

type RegistryConfigModel struct {
	AccountID types.String `tfsdk:"account_id"`
}

type StackCmdWrapperModel struct {
	Command types.String `tfsdk:"command"`
	Include types.List   `tfsdk:"include"`
}

type StackSourceModel struct {
	RepoID        types.String `tfsdk:"repo_id"`
	URL           types.String `tfsdk:"url"`
	AccountID     types.String `tfsdk:"account_id"`
	Path          types.String `tfsdk:"path"`
	Branch        types.String `tfsdk:"branch"`
	Commit        types.String `tfsdk:"commit"`
	CloneEnforced types.Bool   `tfsdk:"reclone_enforced"`
}

type FilesConfigModel struct {
	Contents     types.String `tfsdk:"contents"`
	FilePaths    types.List   `tfsdk:"file_paths"`
	LocalEnabled types.Bool   `tfsdk:"local_enabled"`
	Directory    types.String `tfsdk:"directory"`
}

type StackResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Tags        types.List   `tfsdk:"tags"`
	ServerID    types.String `tfsdk:"server_id"`
	SwarmID     types.String `tfsdk:"swarm_id"`
	ProjectName types.String `tfsdk:"project_name"`

	// Source
	Source *StackSourceModel `tfsdk:"source"`

	Compose *FilesConfigModel `tfsdk:"compose"`

	// Environment
	Environment *EnvironmentModel `tfsdk:"environment"`

	// Behavior flags
	AutoPullEnabled    types.Bool        `tfsdk:"auto_pull_enabled"`
	Build              *BuildConfigModel `tfsdk:"build"`
	DestroyEnforced    types.Bool        `tfsdk:"destroy_enforced"`
	AutoUpdateEnabled  types.Bool        `tfsdk:"auto_update_enabled"`
	AutoUpdateScope    types.String      `tfsdk:"auto_update_scope"`
	PollUpdatesEnabled types.Bool        `tfsdk:"poll_updates_enabled"`
	AlertsEnabled      types.Bool        `tfsdk:"alerts_enabled"`

	// Webhook
	Webhook *StackWebhookModel `tfsdk:"webhook"`

	// Deploy commands
	PreDeploy  *SystemCommandModel `tfsdk:"pre_deploy"`
	PostDeploy *SystemCommandModel `tfsdk:"post_deploy"`

	// Registry
	Registry *RegistryConfigModel `tfsdk:"registry"`

	// Extra args
	ExtraArguments types.List `tfsdk:"extra_arguments"`
	IgnoreServices types.List `tfsdk:"ignore_services"`
	Links          types.List `tfsdk:"links"`

	// Wrapper block
	Wrapper *StackCmdWrapperModel `tfsdk:"wrapper"`
}

func (r *StackResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (r *StackResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The working directory for the command.",
		},
		"command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The shell command to run.",
			PlanModifiers: []planmodifier.String{
				trimTrailingNewlinePlanModifier{},
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo stack resource (Docker Compose / Swarm stack).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The stack identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the stack.",
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
			"server_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the server to deploy the stack on (Compose mode). If both `server_id` and `swarm_id` are set, `swarm_id` takes precedence.",
			},
			"swarm_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the swarm to deploy the stack on (Swarm mode). Overrides `server_id`.",
			},
			"project_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom project name for `docker compose -p`. Defaults to the stack name when empty.",
			},
			"auto_pull_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to run `compose pull` before redeploying to ensure the latest images are used.",
			},
			"destroy_enforced": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to run `docker compose down` before `compose up`.",
			},
			"auto_update_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to automatically redeploy when newer images are found.",
			},
			"auto_update_scope": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("service"),
				MarkdownDescription: "How services are redeployed when `auto_update_enabled` is active. Allowed values: `\"stack\"`, `\"service\"`.",
			},
			"poll_updates_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to poll for image updates.",
			},
			"alerts_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to send stack-state-change alerts for this stack.",
			},
			"extra_arguments": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra arguments appended to `docker compose up -d` (Compose) or `docker stack deploy` (Swarm).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_services": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Services to ignore when checking stack health status (e.g. init containers).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Quick links displayed in the Komodo UI for this stack.",
			},
		},
		Blocks: map[string]schema.Block{
			"source": schema.SingleNestedBlock{
				MarkdownDescription: "Git source configuration for repo-based stacks.",
				Attributes: map[string]schema.Attribute{
					"repo_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Id or name of a linked `komodo_repo` resource.",
					},
					"url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The URL of the git provider, e.g. `https://github.com`.",
					},
					"account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Git account for private repositories.",
					},
					"path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The repository path, e.g. `owner/repo`.",
					},
					"branch": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The branch to check out.",
					},
					"commit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "A specific commit hash to check out.",
					},
					"reclone_enforced": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Whether to delete and reclone the repo folder instead of `git pull`.",
					},
				},
			},
			"compose": schema.SingleNestedBlock{
				MarkdownDescription: "Compose file configuration.",
				Attributes: map[string]schema.Attribute{
					"contents": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Inline compose file contents. When set, this takes precedence over git repo sourcing.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"local_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Whether to source compose files from the host filesystem instead of a git repo or inline contents.",
					},
					"directory": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Directory to `cd` into before running `docker compose up`.",
					},
					"file_paths": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Paths to compose files relative to `directory`. Defaults to `[\"compose.yaml\"]` when empty.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"environment": schema.SingleNestedBlock{
				MarkdownDescription: "Environment variable configuration written to an env file before deploying.",
				Attributes: map[string]schema.Attribute{
					"file_path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Relative path (from `run_directory`) for the written env file. Defaults to `.env`.",
					},
					"variables": schema.MapAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Environment variables to inject. Keys are automatically uppercased.",
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"build": schema.SingleNestedBlock{
				MarkdownDescription: "Build configuration. When set, `docker compose build` is run before deploying.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Whether to run `docker compose build` before deploying.",
					},
					"extra_arguments": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Extra arguments appended to `docker compose build`.",
					},
				},
			},
			"webhook": schema.SingleNestedBlock{
				MarkdownDescription: "Webhook configuration for the stack.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether incoming webhooks trigger a deployment.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Alternate webhook secret. Defaults to the global secret when empty.",
					},
					"force_deploy": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "When true, always runs `DeployStack` instead of `DeployStackIfChanged`.",
					},
				},
			},
			"pre_deploy": schema.SingleNestedBlock{
				MarkdownDescription: "Command to run before the stack is deployed.",
				Attributes:          systemCommandAttrs,
			},
			"post_deploy": schema.SingleNestedBlock{
				MarkdownDescription: "Command to run after the stack is deployed.",
				Attributes:          systemCommandAttrs,
			},
			"registry": schema.SingleNestedBlock{
				MarkdownDescription: "Registry login configuration. When set, `docker login` is run before deploying.",
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The ID of a `komodo_registry_account` resource to authenticate with.",
					},
				},
			},
			"wrapper": schema.SingleNestedBlock{
				MarkdownDescription: "Compose command wrapper configuration for secrets management or custom tooling.",
				Attributes: map[string]schema.Attribute{
					"command": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "A command prefix to wrap the compose command, e.g. for secrets management. Use `[[COMPOSE_COMMAND]]` as placeholder.",
					},
					"include": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Which compose subcommands get wrapped by `command`. Allowed values: `\"config\"`, `\"build\"`, `\"pull\"`, `\"up\"`, `\"run\"`.",
						Validators: []validator.List{
							listvalidator.ValueStringsAre(
								stringvalidator.OneOf("config", "build", "pull", "up", "run"),
							),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *StackResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ConfigValidators returns validators that run against the whole resource config.
func (r *StackResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		systemCommandPathRequiresCommandValidator{blockName: "pre_deploy"},
		systemCommandPathRequiresCommandValidator{blockName: "post_deploy"},
		autoUpdateRequiresPollUpdatesValidator{},
		autoUpdateScopeValidator{},
		gitRepoConflictsValidator{},
	}
}

// systemCommandPathRequiresCommandValidator rejects configs where `path` is set
// inside a SystemCommand block but `command` is absent — a path without a
// command to run is meaningless.
type systemCommandPathRequiresCommandValidator struct {
	blockName string
}

func (v systemCommandPathRequiresCommandValidator) Description(_ context.Context) string {
	return fmt.Sprintf("`%s.path` cannot be set without `%s.command`", v.blockName, v.blockName)
}

func (v systemCommandPathRequiresCommandValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v systemCommandPathRequiresCommandValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var block *SystemCommandModel
	switch v.blockName {
	case "pre_deploy":
		block = data.PreDeploy
	case "post_deploy":
		block = data.PostDeploy
	}

	if block == nil {
		return
	}
	if !block.Path.IsNull() && !block.Path.IsUnknown() && block.Path.ValueString() != "" &&
		(block.Command.IsNull() || block.Command.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root(v.blockName).AtName("path"),
			"path requires command",
			fmt.Sprintf("`%s.path` is set but `%s.command` is not. A working directory without a command to run has no effect.", v.blockName, v.blockName),
		)
	}
}

// autoUpdateRequiresPollUpdatesValidator rejects configs where auto_update_enabled
// is true but poll_updates_enabled is false, since auto_update_enabled has no effect
// without polling.
type autoUpdateRequiresPollUpdatesValidator struct{}

func (v autoUpdateRequiresPollUpdatesValidator) Description(_ context.Context) string {
	return "`auto_update_enabled = true` requires `poll_updates_enabled = true`"
}

func (v autoUpdateRequiresPollUpdatesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v autoUpdateRequiresPollUpdatesValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AutoUpdateEnabled.ValueBool() && !data.PollUpdatesEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("auto_update_enabled"),
			"auto_update_enabled requires poll_updates_enabled",
			"`auto_update_enabled = true` has no effect unless `poll_updates_enabled = true` is also set.",
		)
	}
}

// autoUpdateScopeValidator rejects configs where auto_update_scope is not one of
// the two allowed values, or where "stack" is set but auto_update_enabled
// is false.
type autoUpdateScopeValidator struct{}

func (v autoUpdateScopeValidator) Description(_ context.Context) string {
	return "`auto_update_scope` must be \"stack\" or \"service\"; \"stack\" requires `auto_update_enabled = true`"
}

func (v autoUpdateScopeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v autoUpdateScopeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AutoUpdateScope.IsNull() || data.AutoUpdateScope.IsUnknown() {
		return
	}
	t := data.AutoUpdateScope.ValueString()
	if t != "stack" && t != "service" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auto_update_scope"),
			"Invalid auto_update_scope",
			`auto_update_scope must be "stack" or "service".`,
		)
		return
	}
	if t == "stack" && !data.AutoUpdateEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("auto_update_scope"),
			"auto_update_scope requires auto_update_enabled",
			`auto_update_scope = "stack" has no effect unless auto_update_enabled = true and poll_updates_enabled = true are also set.`,
		)
	}
}

// gitRepoConflictsValidator rejects configs where source.repo_id is set alongside
// any of the direct-clone fields (url, account, path, branch, commit).
// reclone, compose.file_paths, and compose.directory are valid alongside repo_id.
type gitRepoConflictsValidator struct{}

func (v gitRepoConflictsValidator) Description(_ context.Context) string {
	return "`source.repo_id` cannot be set together with `source.url`, `source.account_id`, `source.path`, `source.branch`, or `source.commit`"
}

func (v gitRepoConflictsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v gitRepoConflictsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Source == nil || data.Source.RepoID.IsNull() || data.Source.RepoID.IsUnknown() {
		return
	}
	conflicts := map[string]bool{
		"url":        !data.Source.URL.IsNull() && !data.Source.URL.IsUnknown(),
		"account_id": !data.Source.AccountID.IsNull() && !data.Source.AccountID.IsUnknown(),
		"path":       !data.Source.Path.IsNull() && !data.Source.Path.IsUnknown(),
		"branch":     !data.Source.Branch.IsNull() && !data.Source.Branch.IsUnknown(),
		"commit":     !data.Source.Commit.IsNull() && !data.Source.Commit.IsUnknown(),
	}
	for field, set := range conflicts {
		if set {
			resp.Diagnostics.AddAttributeError(
				path.Root("source").AtName("repo_id"),
				"source.repo_id conflicts with other source fields",
				fmt.Sprintf("`source.repo_id` cannot be set together with `source.%s`. Use either a linked `komodo_repo` (`repo_id`) or direct source fields, not both.", field),
			)
		}
	}
}

func (r *StackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating stack", map[string]interface{}{"name": data.Name.ValueString()})
	createReq := client.CreateStackRequest{
		Name:   data.Name.ValueString(),
		Config: stackConfigFromModel(ctx, r.client, &data),
	}
	stack, err := r.client.CreateStack(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create stack, got error: %s", err))
		return
	}
	if stack.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Stack creation failed: missing ID",
			"The Komodo API did not return a stack ID. Resource cannot be tracked in state.",
		)
		return
	}
	plannedTags := data.Tags
	stackToModel(ctx, r.client, stack, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Stack", ID: stack.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on stack, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	tflog.Trace(ctx, "Created stack resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stack, err := r.client.GetStack(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack, got error: %s", err))
		return
	}
	if stack == nil {
		// Resource may have been externally recreated with a new ID — try name lookup before removing from state.
		stack, err = r.client.GetStack(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack by name, got error: %s", err))
			return
		}
		if stack == nil {
			tflog.Debug(ctx, "Stack not found by ID or name, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Debug(ctx, "Stack adopted by name after ID lookup failed", map[string]interface{}{"name": stack.Name, "new_id": stack.ID.OID})
	}
	stackToModel(ctx, r.client, stack, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state StackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameStack(ctx, client.RenameStackRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename stack, got error: %s", err))
			return
		}
	}
	updateReq := client.UpdateStackRequest{
		ID:     data.ID.ValueString(),
		Config: stackConfigFromModel(ctx, r.client, &data),
	}
	stack, err := r.client.UpdateStack(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update stack, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	stackToModel(ctx, r.client, stack, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Stack", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on stack, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting stack", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteStack(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete stack, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted stack resource")
}

func (r *StackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// stackConfigFromModel converts the Terraform model into a StackConfig.
func stackConfigFromModel(ctx context.Context, c *client.Client, data *StackResourceModel) client.StackConfig {
	cfg := client.StackConfig{
		ServerID:    data.ServerID.ValueString(),
		SwarmID:     data.SwarmID.ValueString(),
		ProjectName: data.ProjectName.ValueString(),

		FileContents: func() string {
			if data.Compose != nil {
				return strings.ReplaceAll(data.Compose.Contents.ValueString(), "\r\n", "\n")
			}
			return ""
		}(),
		FilesOnHost: data.Compose != nil && data.Compose.LocalEnabled.ValueBool(),

		AutoPull: data.AutoPullEnabled.ValueBool(),
		RunBuild: func() bool {
			if data.Build != nil {
				return data.Build.Enabled.ValueBool()
			}
			return false
		}(),
		DestroyBeforeDeploy:   data.DestroyEnforced.ValueBool(),
		AutoUpdate:            data.AutoUpdateEnabled.ValueBool(),
		AutoUpdateAllServices: data.AutoUpdateScope.ValueString() == "stack",
		PollForUpdates:        data.PollUpdatesEnabled.ValueBool(),
		SendAlerts:            data.AlertsEnabled.ValueBool(),

		RegistryProvider: func() string {
			if data.Registry != nil {
				if acc, err := c.GetDockerRegistryAccount(ctx, data.Registry.AccountID.ValueString()); err == nil && acc != nil {
					return acc.Domain
				}
			}
			return ""
		}(),
		RegistryAccount: func() string {
			if data.Registry != nil {
				if acc, err := c.GetDockerRegistryAccount(ctx, data.Registry.AccountID.ValueString()); err == nil && acc != nil {
					return acc.Username
				}
			}
			return ""
		}(),

		ComposeCmdWrapper: func() string {
			if data.Wrapper != nil {
				return data.Wrapper.Command.ValueString()
			}
			return ""
		}(),

		WebhookEnabled: func() bool {
			if data.Webhook != nil {
				return data.Webhook.Enabled.ValueBool()
			}
			return false
		}(),
		WebhookSecret: func() string {
			if data.Webhook != nil {
				return data.Webhook.Secret.ValueString()
			}
			return ""
		}(),
		WebhookForceDeploy: func() bool {
			if data.Webhook != nil {
				return data.Webhook.ForceDeploy.ValueBool()
			}
			return false
		}(),

		Environment: func() string {
			if data.Environment != nil {
				return envMapToString(ctx, data.Environment.Variables)
			}
			return ""
		}(),
		EnvFilePath: func() string {
			if data.Environment != nil {
				return data.Environment.FilePath.ValueString()
			}
			return ""
		}(),
	}

	if data.Source != nil {
		cfg.LinkedRepo = data.Source.RepoID.ValueString()
		url := data.Source.URL.ValueString()
		if strings.HasPrefix(url, "https://") {
			cfg.GitProvider = strings.TrimPrefix(url, "https://")
			cfg.GitHttps = true
		} else if strings.HasPrefix(url, "http://") {
			cfg.GitProvider = strings.TrimPrefix(url, "http://")
			cfg.GitHttps = false
		} else if url != "" {
			cfg.GitProvider = url
			cfg.GitHttps = true
		} else if acc := c.ResolveGitAccountFull(ctx, data.Source.AccountID.ValueString()); acc != nil {
			// No URL set: derive the provider domain from the account's registered domain
			// so the API stores the correct domain instead of defaulting to "github.com".
			cfg.GitProvider = acc.Domain
			cfg.GitHttps = true
		}
		account, err := c.ResolveGitAccountUsername(ctx, data.Source.AccountID.ValueString())
		if err != nil {
			account = data.Source.AccountID.ValueString()
		}
		cfg.GitAccount = account
		cfg.Repo = data.Source.Path.ValueString()
		cfg.Branch = data.Source.Branch.ValueString()
		cfg.Commit = data.Source.Commit.ValueString()
		cfg.Reclone = data.Source.CloneEnforced.ValueBool()
	}

	if data.Compose != nil {
		cfg.RunDirectory = data.Compose.Directory.ValueString()
		if !data.Compose.FilePaths.IsNull() && !data.Compose.FilePaths.IsUnknown() {
			var vals []string
			data.Compose.FilePaths.ElementsAs(ctx, &vals, false)
			cfg.FilePaths = vals
		}
	}

	if data.PreDeploy != nil {
		cfg.PreDeploy = client.SystemCommand{
			Path:    data.PreDeploy.Path.ValueString(),
			Command: data.PreDeploy.Command.ValueString(),
		}
	}
	if data.PostDeploy != nil {
		cfg.PostDeploy = client.SystemCommand{
			Path:    data.PostDeploy.Path.ValueString(),
			Command: data.PostDeploy.Command.ValueString(),
		}
	}

	if !data.ExtraArguments.IsNull() && !data.ExtraArguments.IsUnknown() {
		var vals []string
		data.ExtraArguments.ElementsAs(ctx, &vals, false)
		cfg.ExtraArgs = vals
	}
	if data.Build != nil {
		if !data.Build.ExtraArguments.IsNull() && !data.Build.ExtraArguments.IsUnknown() {
			var vals []string
			data.Build.ExtraArguments.ElementsAs(ctx, &vals, false)
			cfg.BuildExtraArgs = vals
		} else {
			// extra_arguments omitted — explicitly clear so the API removes any previously set value
			cfg.BuildExtraArgs = []string{}
		}
	}
	if data.Wrapper != nil {
		if !data.Wrapper.Include.IsNull() && !data.Wrapper.Include.IsUnknown() {
			var vals []string
			data.Wrapper.Include.ElementsAs(ctx, &vals, false)
			cfg.ComposeCmdWrapperInclude = vals
		}
	}
	if !data.IgnoreServices.IsNull() && !data.IgnoreServices.IsUnknown() {
		var vals []string
		data.IgnoreServices.ElementsAs(ctx, &vals, false)
		cfg.IgnoreServices = vals
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var vals []string
		data.Links.ElementsAs(ctx, &vals, false)
		cfg.Links = vals
	}

	return cfg
}

// stackToModel populates the Terraform model from a Stack API response.
func stackToModel(ctx context.Context, c *client.Client, stack *client.Stack, data *StackResourceModel) {
	data.ID = types.StringValue(stack.ID.OID)
	data.Name = types.StringValue(stack.Name)
	tagsSlice := stack.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	data.Tags = tags

	strOrNull := func(s string) types.String {
		if s != "" {
			return types.StringValue(s)
		}
		return types.StringNull()
	}

	data.ServerID = strOrNull(stack.Config.ServerID)
	data.SwarmID = strOrNull(stack.Config.SwarmID)
	data.ProjectName = strOrNull(stack.Config.ProjectName)
	data.AutoPullEnabled = types.BoolValue(stack.Config.AutoPull)
	data.DestroyEnforced = types.BoolValue(stack.Config.DestroyBeforeDeploy)
	data.AutoUpdateEnabled = types.BoolValue(stack.Config.AutoUpdate)
	if stack.Config.AutoUpdateAllServices {
		data.AutoUpdateScope = types.StringValue("stack")
	} else {
		data.AutoUpdateScope = types.StringValue("service")
	}
	data.PollUpdatesEnabled = types.BoolValue(stack.Config.PollForUpdates)
	data.AlertsEnabled = types.BoolValue(stack.Config.SendAlerts)
	if stack.Config.RegistryProvider != "" || stack.Config.RegistryAccount != "" {
		accountID := c.ResolveDockerRegistryAccountID(ctx, stack.Config.RegistryProvider, stack.Config.RegistryAccount)
		data.Registry = &RegistryConfigModel{
			AccountID: types.StringValue(accountID),
		}
	} else {
		data.Registry = nil
	}
	// Git block: keep nil when all git fields are empty/default and caller didn't set it
	if data.Source != nil || stack.Config.Repo != "" || stack.Config.Branch != "" ||
		stack.Config.GitProvider != "" || stack.Config.GitAccount != "" || stack.Config.Commit != "" ||
		stack.Config.LinkedRepo != "" || stack.Config.Reclone {
		// When repo_id is set, URL/account_id/path/branch/commit are derived from
		// the linked repo by the API and must not be persisted to state (they were
		// never in config), otherwise plan == null but state == derived value
		// causes a perpetual diff.
		var urlVal types.String
		accountIDVal := types.StringNull()
		pathVal := types.StringNull()
		branchVal := types.StringNull()
		commitVal := types.StringNull()
		if stack.Config.LinkedRepo == "" {
			// Only reconstruct url if the prior state/plan had it explicitly set.
			// If user omitted url (relying on account_id for domain), keep it null.
			priorURLSet := data.Source != nil && !data.Source.URL.IsNull()
			if stack.Config.GitProvider != "" && priorURLSet {
				if stack.Config.GitHttps {
					urlVal = types.StringValue("https://" + stack.Config.GitProvider)
				} else {
					urlVal = types.StringValue("http://" + stack.Config.GitProvider)
				}
			} else {
				urlVal = types.StringNull()
			}
			id := c.ResolveGitAccountID(ctx, stack.Config.GitProvider, stack.Config.GitAccount)
			if id != "" {
				accountIDVal = types.StringValue(id)
			}
			pathVal = strOrNull(stack.Config.Repo)
			branchVal = strOrNull(stack.Config.Branch)
			commitVal = strOrNull(stack.Config.Commit)
		} else {
			urlVal = types.StringNull()
		}
		data.Source = &StackSourceModel{
			RepoID:        strOrNull(stack.Config.LinkedRepo),
			URL:           urlVal,
			AccountID:     accountIDVal,
			Path:          pathVal,
			Branch:        branchVal,
			Commit:        commitVal,
			CloneEnforced: types.BoolValue(stack.Config.Reclone),
		}
	} else {
		data.Source = nil
	}

	// Webhook block: only set when non-default
	if stack.Config.WebhookEnabled || stack.Config.WebhookSecret != "" || stack.Config.WebhookForceDeploy {
		data.Webhook = &StackWebhookModel{
			Enabled:     types.BoolValue(stack.Config.WebhookEnabled),
			Secret:      strOrNull(stack.Config.WebhookSecret),
			ForceDeploy: types.BoolValue(stack.Config.WebhookForceDeploy),
		}
	} else {
		data.Webhook = nil
	}

	// pre_deploy / post_deploy
	// Use strOrNull so that empty-string fields from the API are stored as null,
	// matching a config that omits the field (avoids persistent diffs).
	if stack.Config.PreDeploy.Path != "" || stack.Config.PreDeploy.Command != "" {
		data.PreDeploy = &SystemCommandModel{
			Path:    strOrNull(stack.Config.PreDeploy.Path),
			Command: strOrNull(strings.TrimRight(stack.Config.PreDeploy.Command, "\n")),
		}
	} else {
		data.PreDeploy = nil
	}
	if stack.Config.PostDeploy.Path != "" || stack.Config.PostDeploy.Command != "" {
		data.PostDeploy = &SystemCommandModel{
			Path:    strOrNull(stack.Config.PostDeploy.Path),
			Command: strOrNull(strings.TrimRight(stack.Config.PostDeploy.Command, "\n")),
		}
	} else {
		data.PostDeploy = nil
	}

	// Environment block
	envVars := envStringToMap(strings.TrimRight(stack.Config.Environment, "\n"))
	if len(envVars.Elements()) > 0 || stack.Config.EnvFilePath != "" {
		data.Environment = &EnvironmentModel{
			FilePath:  strOrNull(stack.Config.EnvFilePath),
			Variables: envVars,
		}
	} else if data.Environment != nil {
		data.Environment = &EnvironmentModel{
			FilePath:  types.StringNull(),
			Variables: envVars,
		}
	} else {
		data.Environment = nil
	}

	// Lists
	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ExtraArgs)
	data.ExtraArguments = extraArgs

	if data.Compose != nil || stack.Config.FileContents != "" || stack.Config.FilesOnHost ||
		len(stack.Config.FilePaths) > 0 || stack.Config.RunDirectory != "" {
		filePaths, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.FilePaths)
		normalize := func(s string) string {
			return strings.TrimRight(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
		}
		apiNorm := normalize(stack.Config.FileContents)
		var contents types.String
		if data.Compose != nil && !data.Compose.Contents.IsNull() && !data.Compose.Contents.IsUnknown() {
			if normalize(data.Compose.Contents.ValueString()) == apiNorm {
				contents = data.Compose.Contents
			} else {
				contents = strOrNull(apiNorm)
			}
		} else {
			contents = strOrNull(apiNorm)
		}
		data.Compose = &FilesConfigModel{
			Contents:     contents,
			LocalEnabled: types.BoolValue(stack.Config.FilesOnHost),
			Directory:    strOrNull(stack.Config.RunDirectory),
			FilePaths:    filePaths,
		}
	} else {
		data.Compose = nil
	}

	// Build block
	if stack.Config.RunBuild || len(stack.Config.BuildExtraArgs) > 0 || data.Build != nil {
		var buildExtraArgs types.List
		if len(stack.Config.BuildExtraArgs) > 0 {
			buildExtraArgs, _ = types.ListValueFrom(ctx, types.StringType, stack.Config.BuildExtraArgs)
		} else {
			buildExtraArgs = types.ListNull(types.StringType)
		}
		data.Build = &BuildConfigModel{
			Enabled:        types.BoolValue(stack.Config.RunBuild),
			ExtraArguments: buildExtraArgs,
		}
	} else {
		data.Build = nil
	}

	ignoreServices, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.IgnoreServices)
	data.IgnoreServices = ignoreServices

	// Wrapper block
	if stack.Config.ComposeCmdWrapper != "" || len(stack.Config.ComposeCmdWrapperInclude) > 0 {
		wrapperInclude, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ComposeCmdWrapperInclude)
		data.Wrapper = &StackCmdWrapperModel{
			Command: strOrNull(stack.Config.ComposeCmdWrapper),
			Include: wrapperInclude,
		}
	} else {
		data.Wrapper = nil
	}

	links, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.Links)
	data.Links = links
}

// trimTrailingNewlinePlanModifier strips the trailing newline from a string plan
// value. The Komodo API removes trailing newlines from command strings when
// returning them, which would otherwise cause "inconsistent result" errors.
type trimTrailingNewlinePlanModifier struct{}

func (m trimTrailingNewlinePlanModifier) Description(_ context.Context) string {
	return "Trims the trailing newline from the planned value."
}

func (m trimTrailingNewlinePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Trims the trailing newline from the planned value."
}

func (m trimTrailingNewlinePlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.StringValue(strings.TrimRight(req.PlanValue.ValueString(), "\n"))
}
