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

var _ datasource.DataSource = &RepoDataSource{}
var _ datasource.DataSourceWithValidateConfig = &RepoDataSource{}

func NewRepoDataSource() datasource.DataSource {
	return &RepoDataSource{}
}

type RepoDataSource struct {
	client *client.Client
}

type RepoDataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	ServerID    types.String                `tfsdk:"server_id"`
	BuilderID   types.String                `tfsdk:"builder_id"`
	Source      *RepositoryProviderModel    `tfsdk:"source"`
	Path        types.String                `tfsdk:"path"`
	Webhook     *RepoDataSourceWebhookModel `tfsdk:"webhook"`
	OnClone     *SystemCommandModel         `tfsdk:"on_clone"`
	OnPull      *SystemCommandModel         `tfsdk:"on_pull"`
	Links       types.List                  `tfsdk:"links"`
	Environment *EnvironmentModel           `tfsdk:"environment"`
}

func (d *RepoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo"
}

func (d *RepoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	systemCommandAttrs := map[string]schema.Attribute{
		"path": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The working directory for the command.",
		},
		"command": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The shell command to run.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo git repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The git repository identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the git repository. One of `name` or `id` must be set.",
			},
			"server_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the server the repo is cloned on.",
			},
			"builder_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the attached builder.",
			},
			"source": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Git provider configuration.",
				Attributes: map[string]schema.Attribute{
					"domain": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The git provider domain without protocol prefix (e.g. `github.com`).",
					},
					"https_enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether HTTPS is used for cloning.",
					},
					"account_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The git account used for private repositories.",
					},
					"path": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The repository path, e.g. `owner/repo`.",
					},
					"branch": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The branch checked out.",
					},
					"commit": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The specific commit hash checked out.",
					},
				},
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The folder on the server the repo is cloned into.",
			},
			"webhook": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook configuration and URLs for the repository.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether webhooks trigger an action on this repository.",
					},
					"secret": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The alternate webhook secret.",
					},
					"url_pull": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Webhook URL to trigger a git pull.",
					},
					"url_clone": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Webhook URL to trigger a git clone.",
					},
					"url_build": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Webhook URL to trigger a build.",
					},
				},
			},
			"on_clone": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The command run after the repository is cloned.",
				Attributes:          systemCommandAttrs,
			},
			"on_pull": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The command run after the repository is pulled.",
				Attributes:          systemCommandAttrs,
			},
			"links": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links associated with this repository.",
			},
			"environment": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Environment configuration for the repository.",
				Attributes: map[string]schema.Attribute{
					"file_path": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Path to the environment file.",
					},
					"variables": schema.MapAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Environment variables injected. Keys are uppercased.",
					},
				},
			},
		},
	}
}

func (d *RepoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepoDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data RepoDataSourceModel
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

func (d *RepoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading git repository data source", map[string]interface{}{"lookup": lookup})

	repo, err := d.client.GetGitRepository(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read git repository, got error: %s", err))
		return
	}
	if repo == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Git repository %q not found", lookup))
		return
	}

	data.ID = types.StringValue(repo.ID.OID)
	data.Name = types.StringValue(repo.Name)
	data.ServerID = types.StringValue(repo.Config.ServerID)
	data.BuilderID = types.StringValue(repo.Config.BuilderID)
	gitAccount := types.StringNull()
	if repo.Config.GitAccount != "" {
		gitAccount = types.StringValue(repo.Config.GitAccount)
	}
	domainVal := types.StringNull()
	if repo.Config.GitProvider != "" {
		domainVal = types.StringValue(repo.Config.GitProvider)
	}
	data.Source = &RepositoryProviderModel{
		Domain:       domainVal,
		HttpsEnabled: types.BoolValue(repo.Config.GitHttps),
		AccountID:    gitAccount,
		Path:         types.StringValue(repo.Config.Repo),
		Branch:       types.StringValue(repo.Config.Branch),
		Commit:       types.StringValue(repo.Config.Commit),
	}
	data.Path = types.StringValue(repo.Config.Path)
	secret := types.StringNull()
	if repo.Config.WebhookSecret != "" {
		secret = types.StringValue(repo.Config.WebhookSecret)
	}
	data.Webhook = &RepoDataSourceWebhookModel{
		Enabled:  types.BoolValue(repo.Config.WebhookEnabled),
		Secret:   secret,
		UrlPull:  types.StringNull(),
		UrlClone: types.StringNull(),
		UrlBuild: types.StringNull(),
	}
	envVars := envStringToMap(strings.TrimRight(repo.Config.Environment, "\n"))
	filePath := types.StringNull()
	if repo.Config.EnvFilePath != "" {
		filePath = types.StringValue(repo.Config.EnvFilePath)
	}
	data.Environment = &EnvironmentModel{
		FilePath:  filePath,
		Variables: envVars,
	}

	data.OnClone = &SystemCommandModel{
		Path:    types.StringValue(repo.Config.OnClone.Path),
		Command: types.StringValue(strings.TrimRight(repo.Config.OnClone.Command, "\n")),
	}
	data.OnPull = &SystemCommandModel{
		Path:    types.StringValue(repo.Config.OnPull.Path),
		Command: types.StringValue(strings.TrimRight(repo.Config.OnPull.Command, "\n")),
	}

	links, linksDiags := types.ListValueFrom(ctx, types.StringType, repo.Config.Links)
	resp.Diagnostics.Append(linksDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Links = links

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
