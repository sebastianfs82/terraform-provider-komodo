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

var _ datasource.DataSource = &StacksDataSource{}

func NewStacksDataSource() datasource.DataSource {
	return &StacksDataSource{}
}

type StacksDataSource struct {
	client *client.Client
}

type StacksDataSourceModel struct {
	ServerID types.String           `tfsdk:"server_id"`
	Stacks   []StackDataSourceModel `tfsdk:"stacks"`
}

func (d *StacksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stacks"
}

func (d *StacksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
		MarkdownDescription: "Lists all Komodo stacks visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "When set, only stacks deployed on this server ID are returned.",
			},
			"stacks": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of stacks.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The stack identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the stack.",
						},
						"server_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the server the stack runs on.",
						},
						"swarm_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the swarm the stack runs on.",
						},
						"project_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Custom project name for `docker compose -p`.",
						},
						"source": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Git source configuration.",
							Attributes: map[string]schema.Attribute{
								"repo_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Name or ID of the linked `komodo_repo` resource.",
								},
								"url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The git provider URL, e.g. `https://github.com`.",
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
								"reclone_enforced": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether the repo folder is deleted and recloned instead of `git pull`.",
								},
							},
						},
						"compose": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Compose file configuration.",
							Attributes: map[string]schema.Attribute{
								"contents": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Inline compose file contents.",
								},
								"local_enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether compose files are sourced from the host filesystem.",
								},
								"directory": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Directory to `cd` into before running `docker compose up`.",
								},
								"file_paths": schema.ListAttribute{
									Computed:            true,
									ElementType:         types.StringType,
									MarkdownDescription: "Paths to compose files relative to `directory`.",
								},
							},
						},
						"environment": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Environment variable configuration written to an env file before deploying.",
							Attributes: map[string]schema.Attribute{
								"file_path": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Relative path for the written env file.",
								},
								"variables": schema.MapAttribute{
									Computed:            true,
									ElementType:         types.StringType,
									MarkdownDescription: "Environment variables injected.",
								},
							},
						},
						"auto_pull_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether `compose pull` is run before every deploy.",
						},
						"build": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Build configuration for the stack.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether `docker compose build` is run before deploying.",
								},
								"extra_arguments": schema.ListAttribute{
									Computed:            true,
									ElementType:         types.StringType,
									MarkdownDescription: "Extra arguments appended to `docker compose build`.",
								},
							},
						},
						"destroy_enforced": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether `docker compose down` is run before `compose up`.",
						},
						"auto_update_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the stack is automatically redeployed when newer images are found.",
						},
						"auto_update_scope": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "How services are redeployed when `auto_update_enabled` is active. Either `\"stack\"` or `\"service\"`.",
						},
						"poll_updates_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether to poll for image updates.",
						},
						"alerts_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether stack-state-change alerts are sent.",
						},
						"webhook": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Webhook configuration for the stack.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether incoming webhooks trigger a deployment.",
								},
								"secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Alternate webhook secret.",
								},
								"force_deploy": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether the webhook always runs `DeployStack` instead of `DeployStackIfChanged`.",
								},
							},
						},
						"pre_deploy": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Command to run before the stack is deployed.",
							Attributes:          systemCommandAttrs,
						},
						"post_deploy": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Command to run after the stack is deployed.",
							Attributes:          systemCommandAttrs,
						},
						"registry": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Registry login configuration for the stack.",
							Attributes: map[string]schema.Attribute{
								"account_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The ID of the `komodo_registry_account` used to authenticate.",
								},
							},
						},
						"extra_arguments": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Extra arguments appended to the compose/stack deploy command.",
						},
						"compose_cmd_wrapper": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A command prefix to wrap the compose command.",
						},
						"compose_cmd_wrapper_include": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Which compose subcommands get wrapped by `compose_cmd_wrapper`.",
						},
						"ignore_services": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Services ignored when checking stack health status.",
						},
						"links": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Quick links displayed in the Komodo UI for this stack.",
						},
					},
				},
			},
		},
	}
}

func (d *StacksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StacksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing stacks")

	stacks, err := d.client.ListStacks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list stacks, got error: %s", err))
		return
	}

	serverIDFilter := data.ServerID.ValueString()

	strOrNull := func(s string) types.String {
		if s != "" {
			return types.StringValue(s)
		}
		return types.StringNull()
	}

	items := make([]StackDataSourceModel, 0, len(stacks))
	for _, stack := range stacks {
		if serverIDFilter != "" && stack.Config.ServerID != serverIDFilter {
			continue
		}
		envVars := envStringToMap(strings.TrimRight(stack.Config.Environment, "\n"))
		filePath := types.StringNull()
		if stack.Config.EnvFilePath != "" {
			filePath = types.StringValue(stack.Config.EnvFilePath)
		}
		var environment *EnvironmentModel
		if len(envVars.Elements()) > 0 || stack.Config.EnvFilePath != "" {
			environment = &EnvironmentModel{
				FilePath:  filePath,
				Variables: envVars,
			}
		}

		var urlVal types.String
		if stack.Config.GitProvider != "" {
			if stack.Config.GitHttps {
				urlVal = types.StringValue("https://" + stack.Config.GitProvider)
			} else {
				urlVal = types.StringValue("http://" + stack.Config.GitProvider)
			}
		} else {
			urlVal = types.StringNull()
		}
		extraArgs, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ExtraArgs)
		buildExtraArgs, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.BuildExtraArgs)
		wrapperInclude, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ComposeCmdWrapperInclude)
		ignoreServices, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.IgnoreServices)
		links, linksDiags := types.ListValueFrom(ctx, types.StringType, stack.Config.Links)
		resp.Diagnostics.Append(linksDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		items = append(items, StackDataSourceModel{
			ID:          types.StringValue(stack.ID.OID),
			Name:        types.StringValue(stack.Name),
			ServerID:    strOrNull(stack.Config.ServerID),
			SwarmID:     strOrNull(stack.Config.SwarmID),
			ProjectName: strOrNull(stack.Config.ProjectName),
			Files: func() *FilesConfigModel {
				f, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.FilePaths)
				return &FilesConfigModel{
					Contents:     strOrNull(stack.Config.FileContents),
					LocalEnabled: types.BoolValue(stack.Config.FilesOnHost),
					Directory:    strOrNull(stack.Config.RunDirectory),
					FilePaths:    f,
				}
			}(),
			Environment:     environment,
			AutoPullEnabled: types.BoolValue(stack.Config.AutoPull),
			Build: &BuildConfigModel{
				Enabled:        types.BoolValue(stack.Config.RunBuild),
				ExtraArguments: buildExtraArgs,
			},
			DestroyEnforced:   types.BoolValue(stack.Config.DestroyBeforeDeploy),
			AutoUpdateEnabled: types.BoolValue(stack.Config.AutoUpdate),
			AutoUpdateScope: func() types.String {
				if stack.Config.AutoUpdateAllServices {
					return types.StringValue("stack")
				}
				return types.StringValue("service")
			}(),
			PollUpdatesEnabled: types.BoolValue(stack.Config.PollForUpdates),
			AlertsEnabled:      types.BoolValue(stack.Config.SendAlerts),
			Registry: &RegistryConfigModel{
				AccountID: types.StringValue(d.client.ResolveDockerRegistryAccountID(ctx, stack.Config.RegistryProvider, stack.Config.RegistryAccount)),
			},
			ComposeCmdWrapper:        strOrNull(stack.Config.ComposeCmdWrapper),
			ExtraArguments:           extraArgs,
			ComposeCmdWrapperInclude: wrapperInclude,
			IgnoreServices:           ignoreServices,
			Links:                    links,
			Source: &StackSourceModel{
				RepoID:        strOrNull(stack.Config.LinkedRepo),
				URL:           urlVal,
				AccountID:     strOrNull(stack.Config.GitAccount),
				Path:          strOrNull(stack.Config.Repo),
				Branch:        strOrNull(stack.Config.Branch),
				Commit:        strOrNull(stack.Config.Commit),
				CloneEnforced: types.BoolValue(stack.Config.Reclone),
			},
			Webhook: &StackDataSourceWebhookModel{
				Enabled:     types.BoolValue(stack.Config.WebhookEnabled),
				Secret:      strOrNull(stack.Config.WebhookSecret),
				ForceDeploy: types.BoolValue(stack.Config.WebhookForceDeploy),
			},
			PreDeploy: &SystemCommandModel{
				Path:    types.StringValue(stack.Config.PreDeploy.Path),
				Command: types.StringValue(strings.TrimRight(stack.Config.PreDeploy.Command, "\n")),
			},
			PostDeploy: &SystemCommandModel{
				Path:    types.StringValue(stack.Config.PostDeploy.Path),
				Command: types.StringValue(strings.TrimRight(stack.Config.PostDeploy.Command, "\n")),
			},
		})
	}

	data.Stacks = items
	tflog.Trace(ctx, "Listed stacks", map[string]interface{}{"count": len(items)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
