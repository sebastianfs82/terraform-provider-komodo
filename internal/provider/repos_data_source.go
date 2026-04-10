// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ReposDataSource{}

func NewReposDataSource() datasource.DataSource {
	return &ReposDataSource{}
}

type ReposDataSource struct {
	client *client.Client
}

type ReposDataSourceModel struct {
	Repositories []RepoDataSourceModel `tfsdk:"repositories"`
}

func (d *ReposDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repos"
}

func (d *ReposDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
		MarkdownDescription: "Lists all Komodo git repositories visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of git repositories.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git repository identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the git repository.",
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
							MarkdownDescription: "Git source configuration.",
							Attributes: map[string]schema.Attribute{
								"url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The URL of the git provider, e.g. `https://github.com`.",
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
							MarkdownDescription: "Webhook configuration for the repository.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether webhooks trigger an action on this repository.",
								},
								"secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The alternate webhook secret.",
								}, "url_pull": schema.StringAttribute{
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
								}},
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
							MarkdownDescription: "Environment variable configuration.",
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
				},
			},
		},
	}
}

func (d *ReposDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ReposDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ReposDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing git repositories")

	repos, err := d.client.ListGitRepositories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list git repositories, got error: %s", err))
		return
	}

	items := make([]RepoDataSourceModel, 0, len(repos))
	for _, repo := range repos {
		gitAccount := types.StringNull()
		if repo.Config.GitAccount != "" {
			gitAccount = types.StringValue(repo.Config.GitAccount)
		}

		secret := types.StringNull()
		if repo.Config.WebhookSecret != "" {
			secret = types.StringValue(repo.Config.WebhookSecret)
		}
		var webhook *RepoDataSourceWebhookModel
		if repo.Config.WebhookEnabled || repo.Config.WebhookSecret != "" {
			webhook = &RepoDataSourceWebhookModel{
				Enabled:  types.BoolValue(repo.Config.WebhookEnabled),
				Secret:   secret,
				UrlPull:  types.StringNull(),
				UrlClone: types.StringNull(),
				UrlBuild: types.StringNull(),
			}
		}

		envVars := envStringToMap(strings.TrimRight(repo.Config.Environment, "\n"))
		var environment *EnvironmentModel
		if !envVars.IsNull() && len(envVars.Elements()) > 0 || repo.Config.EnvFilePath != "" {
			filePath := types.StringNull()
			if repo.Config.EnvFilePath != "" {
				filePath = types.StringValue(repo.Config.EnvFilePath)
			}
			environment = &EnvironmentModel{
				FilePath:  filePath,
				Variables: envVars,
			}
		}

		links, linksDiags := types.ListValueFrom(ctx, types.StringType, repo.Config.Links)
		resp.Diagnostics.Append(linksDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var repoURLVal types.String
		if repo.Config.GitProvider != "" {
			if repo.Config.GitHttps {
				repoURLVal = types.StringValue("https://" + repo.Config.GitProvider)
			} else {
				repoURLVal = types.StringValue("http://" + repo.Config.GitProvider)
			}
		} else {
			repoURLVal = types.StringNull()
		}

		items = append(items, RepoDataSourceModel{
			ID:        types.StringValue(repo.ID.OID),
			Name:      types.StringValue(repo.Name),
			ServerID:  types.StringValue(repo.Config.ServerID),
			BuilderID: types.StringValue(repo.Config.BuilderID),
			Source: &RepositoryProviderModel{
				URL:       repoURLVal,
				AccountID: gitAccount,
				Path:      types.StringValue(repo.Config.Repo),
				Branch:    types.StringValue(repo.Config.Branch),
				Commit:    types.StringValue(repo.Config.Commit),
			},
			Path:    types.StringValue(repo.Config.Path),
			Webhook: webhook,
			OnClone: &SystemCommandModel{
				Path:    types.StringValue(repo.Config.OnClone.Path),
				Command: types.StringValue(strings.TrimRight(repo.Config.OnClone.Command, "\n")),
			},
			OnPull: &SystemCommandModel{
				Path:    types.StringValue(repo.Config.OnPull.Path),
				Command: types.StringValue(strings.TrimRight(repo.Config.OnPull.Command, "\n")),
			},
			Links:       links,
			Environment: environment,
		})
	}

	data.Repositories = items
	tflog.Trace(ctx, "Listed git repositories", map[string]interface{}{"count": len(items)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
