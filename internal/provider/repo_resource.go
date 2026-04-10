// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &RepoResource{}
var _ resource.ResourceWithImportState = &RepoResource{}
var _ resource.ResourceWithValidateConfig = &RepoResource{}

func NewRepoResource() resource.Resource {
	return &RepoResource{}
}

type RepoResource struct {
	client *client.Client
}

type SystemCommandModel struct {
	Path    types.String `tfsdk:"path"`
	Command types.String `tfsdk:"command"`
}

type RepositoryProviderModel struct {
	Domain       types.String `tfsdk:"domain"`
	HttpsEnabled types.Bool   `tfsdk:"https_enabled"`
	AccountID    types.String `tfsdk:"account_id"`
	Path         types.String `tfsdk:"path"`
	Branch       types.String `tfsdk:"branch"`
	Commit       types.String `tfsdk:"commit"`
}

type EnvironmentModel struct {
	FilePath  types.String `tfsdk:"file_path"`
	Variables types.Map    `tfsdk:"variables"`
}

type WebhookModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Secret  types.String `tfsdk:"secret"`
}

// RepoDataSourceWebhookModel extends WebhookModel with read-only webhook URL fields
// exposed by the repo and repos data sources.
type RepoDataSourceWebhookModel struct {
	Enabled  types.Bool   `tfsdk:"enabled"`
	Secret   types.String `tfsdk:"secret"`
	UrlPull  types.String `tfsdk:"url_pull"`
	UrlClone types.String `tfsdk:"url_clone"`
	UrlBuild types.String `tfsdk:"url_build"`
}

type BuildConfigModel struct {
	Enabled        types.Bool `tfsdk:"enabled"`
	ExtraArguments types.List `tfsdk:"extra_arguments"`
}

type RepoResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	Name        types.String             `tfsdk:"name"`
	ServerID    types.String             `tfsdk:"server_id"`
	BuilderID   types.String             `tfsdk:"builder_id"`
	Source      *RepositoryProviderModel `tfsdk:"source"`
	Path        types.String             `tfsdk:"path"`
	Webhook     *WebhookModel            `tfsdk:"webhook"`
	OnClone     *SystemCommandModel      `tfsdk:"on_clone"`
	OnPull      *SystemCommandModel      `tfsdk:"on_pull"`
	Links       types.List               `tfsdk:"links"`
	Environment *EnvironmentModel        `tfsdk:"environment"`
}

func (r *RepoResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo"
}

func (r *RepoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The working directory for the command.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"command": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The shell command to run.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo git repository resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The git repository identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the git repository. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the server to clone the repo on. Omit or set to empty string to disconnect.",
			},
			"builder_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the builder to attach. Omit or set to empty string to disconnect.",
			},
			"source": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Git provider configuration.",
				Attributes: map[string]schema.Attribute{
					"domain": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The git provider domain without protocol prefix (e.g. `github.com`).",
					},
					"https_enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to use HTTPS (true) or HTTP (false) for cloning. Defaults to true.",
					},
					"account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The git account to use for private repositories. Omit or set to empty string to disconnect.",
					},
					"path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The repository to clone, e.g. `owner/repo`. Omit or set to empty string to clear.",
					},
					"branch": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("main"),
						MarkdownDescription: "The branch to check out. Defaults to `main`.",
					},
					"commit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "A specific commit hash to check out. Omit or set to empty string to clear.",
					},
				},
			},
			"path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The folder on the server to clone into. Omit or set to empty string to clear.",
			},
			"webhook": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Webhook configuration for the repository.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether webhooks should trigger an action on this repository.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "An alternate webhook secret for this repository.",
					},
				},
			},
			"on_clone": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A command to run after the repository is cloned.",
				Attributes:          systemCommandAttrs,
			},
			"on_pull": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "A command to run after the repository is pulled.",
				Attributes:          systemCommandAttrs,
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links associated with this repository.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"environment": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Environment configuration for the repository.",
				Attributes: map[string]schema.Attribute{
					"file_path": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Path to the environment file. Defaults to `.env`.",
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
		},
	}
}

func (r *RepoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepoResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data RepoResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Source == nil {
		return
	}
	accountIDSet := !data.Source.AccountID.IsNull() && !data.Source.AccountID.IsUnknown() && data.Source.AccountID.ValueString() != ""
	if accountIDSet {
		if !data.Source.Domain.IsNull() && !data.Source.Domain.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source").AtName("domain"),
				"Conflicting source attributes",
				"source.domain cannot be set when source.account_id is set. The domain is derived automatically from the linked provider account.",
			)
		}
		if !data.Source.HttpsEnabled.IsNull() && !data.Source.HttpsEnabled.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source").AtName("https_enabled"),
				"Conflicting source attributes",
				"source.https_enabled cannot be set when source.account_id is set. The value is derived automatically from the linked provider account.",
			)
		}
	}
}

func (r *RepoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating git repository", map[string]interface{}{
		"name": data.Name.ValueString(),
	})
	createReq := client.CreateGitRepositoryRequest{
		Name:   data.Name.ValueString(),
		Config: repoConfigFromModel(ctx, r.client, &data),
	}
	repo, err := r.client.CreateGitRepository(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create git repository, got error: %s", err))
		return
	}
	if repo.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Git repository creation failed: missing ID",
			"The Komodo API did not return a git repository ID. Resource cannot be tracked in state.",
		)
		return
	}
	repoToModel(ctx, r.client, repo, &data)
	tflog.Trace(ctx, "Created git repository resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	repo, err := r.client.GetGitRepository(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read git repository, got error: %s", err))
		return
	}
	if repo == nil {
		tflog.Debug(ctx, "Git repository not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	repoToModel(ctx, r.client, repo, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state RepoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	updateReq := client.UpdateGitRepositoryRequest{
		ID:     data.ID.ValueString(),
		Config: repoConfigFromModel(ctx, r.client, &data),
	}
	repo, err := r.client.UpdateGitRepository(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update git repository, got error: %s", err))
		return
	}
	repoToModel(ctx, r.client, repo, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RepoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting git repository", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteGitRepository(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete git repository, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted git repository resource")
}

func (r *RepoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// envMapToString converts a Terraform map of env vars to a newline-separated KEY=VALUE string.
// Keys are uppercased. Returns empty string for null/unknown maps.
func envMapToString(ctx context.Context, m types.Map) string {
	if m.IsNull() || m.IsUnknown() {
		return ""
	}
	elems := make(map[string]string, len(m.Elements()))
	_ = m.ElementsAs(ctx, &elems, false)
	keys := make([]string, 0, len(elems))
	for k := range elems {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(strings.ToUpper(k))
		sb.WriteString("=")
		sb.WriteString(elems[k])
		sb.WriteString("\n")
	}
	return sb.String()
}

// envStringToMap parses a newline-separated KEY=VALUE string into a Terraform map.
// Keys are uppercased. Returns an empty map for blank input.
func envStringToMap(s string) types.Map {
	elems := map[string]attr.Value{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			elems[strings.ToUpper(line)] = types.StringValue("")
			continue
		}
		elems[strings.ToUpper(line[:idx])] = types.StringValue(line[idx+1:])
	}
	m, _ := types.MapValue(types.StringType, elems)
	if len(elems) == 0 {
		return types.MapNull(types.StringType)
	}
	return m
}

// repoConfigFromModel converts the Terraform model into a GitRepositoryConfig.
func repoConfigFromModel(ctx context.Context, c *client.Client, data *RepoResourceModel) client.GitRepositoryConfig {
	cfg := client.GitRepositoryConfig{
		ServerID:  data.ServerID.ValueString(),
		BuilderID: data.BuilderID.ValueString(),
		Path:      data.Path.ValueString(),
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
		SkipSecretInterp: false,
	}
	if data.Source != nil {
		// Resolve account_id → full GitProviderAccount so we can derive domain/https from it.
		linkedAccount := c.ResolveGitAccountFull(ctx, data.Source.AccountID.ValueString())

		// domain: explicit value wins; fall back to the linked account's domain.
		if !data.Source.Domain.IsNull() && !data.Source.Domain.IsUnknown() {
			cfg.GitProvider = data.Source.Domain.ValueString()
		} else if linkedAccount != nil {
			cfg.GitProvider = linkedAccount.Domain
		}

		// https_enabled: explicit value wins; fall back to the linked account's setting.
		if !data.Source.HttpsEnabled.IsNull() && !data.Source.HttpsEnabled.IsUnknown() {
			cfg.GitHttps = data.Source.HttpsEnabled.ValueBool()
		} else if linkedAccount != nil {
			cfg.GitHttps = linkedAccount.HttpsEnabled
		} else {
			cfg.GitHttps = true // safe default
		}

		account, err := c.ResolveGitAccountUsername(ctx, data.Source.AccountID.ValueString())
		if err != nil {
			account = data.Source.AccountID.ValueString()
		}
		cfg.GitAccount = account
		cfg.Repo = data.Source.Path.ValueString()
		cfg.Branch = data.Source.Branch.ValueString()
		cfg.Commit = data.Source.Commit.ValueString()
	}

	if data.OnClone != nil {
		cfg.OnClone = client.SystemCommand{
			Path:    data.OnClone.Path.ValueString(),
			Command: data.OnClone.Command.ValueString(),
		}
	}

	if data.OnPull != nil {
		cfg.OnPull = client.SystemCommand{
			Path:    data.OnPull.Path.ValueString(),
			Command: data.OnPull.Command.ValueString(),
		}
	}

	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		data.Links.ElementsAs(ctx, &links, false)
		cfg.Links = links
	}

	return cfg
}

// repoToModel populates the Terraform model from a GitRepository API response.
func repoToModel(ctx context.Context, c *client.Client, repo *client.GitRepository, data *RepoResourceModel) {
	data.ID = types.StringValue(repo.ID.OID)
	data.Name = types.StringValue(repo.Name)
	// Store null when the API returns empty string so removing the attribute from
	// config fully clears it without causing a perpetual diff.
	if repo.Config.ServerID != "" {
		data.ServerID = types.StringValue(repo.Config.ServerID)
	} else {
		data.ServerID = types.StringNull()
	}
	if repo.Config.BuilderID != "" {
		data.BuilderID = types.StringValue(repo.Config.BuilderID)
	} else {
		data.BuilderID = types.StringNull()
	}
	// Populate git block. If the block was nil in config and all git fields are
	// empty/default, keep it nil so no spurious diff is produced.
	gitAccountID := ""
	if repo.Config.GitAccount != "" {
		gitAccountID = c.ResolveGitAccountID(ctx, repo.Config.GitProvider, repo.Config.GitAccount)
	}
	gitAccount := types.StringNull()
	if gitAccountID != "" {
		gitAccount = types.StringValue(gitAccountID)
	}
	if data.Source != nil || repo.Config.Repo != "" || repo.Config.Branch != "" ||
		repo.Config.GitProvider != "" || repo.Config.GitAccount != "" || repo.Config.Commit != "" {
		gitPath := types.StringNull()
		if repo.Config.Repo != "" {
			gitPath = types.StringValue(repo.Config.Repo)
		}
		gitBranch := types.StringNull()
		if repo.Config.Branch != "" {
			gitBranch = types.StringValue(repo.Config.Branch)
		}
		gitCommit := types.StringNull()
		if repo.Config.Commit != "" {
			gitCommit = types.StringValue(repo.Config.Commit)
		}
		// When account_id is set, domain/https_enabled are derived from the linked
		// account and should not be persisted to state (they were never in config).
		// Only store them when the user explicitly configured them (account_id absent).
		domainVal := types.StringNull()
		httpsVal := types.BoolNull()
		if gitAccount.IsNull() {
			if repo.Config.GitProvider != "" {
				domainVal = types.StringValue(repo.Config.GitProvider)
			}
			httpsVal = types.BoolValue(repo.Config.GitHttps)
		}
		data.Source = &RepositoryProviderModel{
			Domain:       domainVal,
			HttpsEnabled: httpsVal,
			AccountID:    gitAccount,
			Path:         gitPath,
			Branch:       gitBranch,
			Commit:       gitCommit,
		}
	} else {
		data.Source = nil
	}
	if repo.Config.Path != "" {
		data.Path = types.StringValue(repo.Config.Path)
	} else {
		data.Path = types.StringNull()
	}
	secret := types.StringNull()
	if repo.Config.WebhookSecret != "" {
		secret = types.StringValue(repo.Config.WebhookSecret)
	}
	if repo.Config.WebhookEnabled || repo.Config.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(repo.Config.WebhookEnabled),
			Secret:  secret,
		}
	} else {
		data.Webhook = nil
	}
	envVars := envStringToMap(strings.TrimRight(repo.Config.Environment, "\n"))
	envFilePath := repo.Config.EnvFilePath
	if !envVars.IsNull() && len(envVars.Elements()) > 0 || envFilePath != "" {
		filePath := types.StringNull()
		if envFilePath != "" {
			filePath = types.StringValue(envFilePath)
		}
		data.Environment = &EnvironmentModel{
			FilePath:  filePath,
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

	// on_clone — only set if non-empty, otherwise leave nil to match Optional schema
	if repo.Config.OnClone.Path != "" || repo.Config.OnClone.Command != "" {
		data.OnClone = &SystemCommandModel{
			Path:    types.StringValue(repo.Config.OnClone.Path),
			Command: types.StringValue(strings.TrimRight(repo.Config.OnClone.Command, "\n")),
		}
	} else {
		data.OnClone = nil
	}

	// on_pull — only set if non-empty, otherwise leave nil to match Optional schema
	if repo.Config.OnPull.Path != "" || repo.Config.OnPull.Command != "" {
		data.OnPull = &SystemCommandModel{
			Path:    types.StringValue(repo.Config.OnPull.Path),
			Command: types.StringValue(strings.TrimRight(repo.Config.OnPull.Command, "\n")),
		}
	} else {
		data.OnPull = nil
	}

	links, _ := types.ListValueFrom(ctx, types.StringType, repo.Config.Links)
	data.Links = links
}
