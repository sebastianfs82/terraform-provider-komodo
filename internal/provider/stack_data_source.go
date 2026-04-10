// Copyright (c) HashiCorp, Inc.
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

var _ datasource.DataSource = &StackDataSource{}

func NewStackDataSource() datasource.DataSource {
	return &StackDataSource{}
}

type StackDataSource struct {
	client *client.Client
}

type StackDataSourceWebhookModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Secret      types.String `tfsdk:"secret"`
	ForceDeploy types.Bool   `tfsdk:"force_deploy"`
}

type StackDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ServerID    types.String `tfsdk:"server_id"`
	SwarmID     types.String `tfsdk:"swarm_id"`
	ProjectName types.String `tfsdk:"project_name"`

	Source *StackSourceModel `tfsdk:"source"`
	Files  *FilesConfigModel `tfsdk:"files"`

	Environment *EnvironmentModel `tfsdk:"environment"`

	AutoPullEnabled    types.Bool        `tfsdk:"auto_pull_enabled"`
	Build              *BuildConfigModel `tfsdk:"build"`
	DestroyEnforced    types.Bool        `tfsdk:"destroy_enforced"`
	AutoUpdateEnabled  types.Bool        `tfsdk:"auto_update_enabled"`
	AutoUpdateScope    types.String      `tfsdk:"auto_update_scope"`
	PollUpdatesEnabled types.Bool        `tfsdk:"poll_updates_enabled"`
	AlertsEnabled      types.Bool        `tfsdk:"alerts_enabled"`

	Webhook    *StackDataSourceWebhookModel `tfsdk:"webhook"`
	PreDeploy  *SystemCommandModel          `tfsdk:"pre_deploy"`
	PostDeploy *SystemCommandModel          `tfsdk:"post_deploy"`

	Registry *RegistryConfigModel `tfsdk:"registry"`

	ExtraArguments           types.List   `tfsdk:"extra_arguments"`
	ComposeCmdWrapper        types.String `tfsdk:"compose_cmd_wrapper"`
	ComposeCmdWrapperInclude types.List   `tfsdk:"compose_cmd_wrapper_include"`
	IgnoreServices           types.List   `tfsdk:"ignore_services"`
	Links                    types.List   `tfsdk:"links"`
}

func (d *StackDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (d *StackDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
		MarkdownDescription: "Reads an existing Komodo stack.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The stack identifier (ObjectId).",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the stack to look up.",
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
						MarkdownDescription: "Id or name of the linked `komodo_repo` resource.",
					},
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
					"reclone_enforced": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the repo folder is deleted and recloned instead of `git pull`.",
					},
				},
			},
			"files": schema.SingleNestedAttribute{
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
					"paths": schema.ListAttribute{
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
					"provider": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Registry provider domain.",
					},
					"account": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Registry account name.",
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
	}
}

func (d *StackDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading stack data source", map[string]interface{}{"name": data.Name.ValueString()})

	stack, err := d.client.GetStack(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack, got error: %s", err))
		return
	}
	if stack == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Stack with name %q not found", data.Name.ValueString()))
		return
	}

	stackToDataSourceModel(ctx, stack, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// stackToDataSourceModel populates a StackDataSourceModel from an API Stack response.
func stackToDataSourceModel(ctx context.Context, stack *client.Stack, data *StackDataSourceModel) {
	strOrNull := func(s string) types.String {
		if s != "" {
			return types.StringValue(s)
		}
		return types.StringNull()
	}

	data.ID = types.StringValue(stack.ID.OID)
	data.Name = types.StringValue(stack.Name)
	data.ServerID = strOrNull(stack.Config.ServerID)
	data.SwarmID = strOrNull(stack.Config.SwarmID)
	data.ProjectName = strOrNull(stack.Config.ProjectName)
	data.AutoPullEnabled = types.BoolValue(stack.Config.AutoPull)
	buildExtraArgs, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.BuildExtraArgs)
	data.Build = &BuildConfigModel{
		Enabled:        types.BoolValue(stack.Config.RunBuild),
		ExtraArguments: buildExtraArgs,
	}
	data.DestroyEnforced = types.BoolValue(stack.Config.DestroyBeforeDeploy)
	data.AutoUpdateEnabled = types.BoolValue(stack.Config.AutoUpdate)
	if stack.Config.AutoUpdateAllServices {
		data.AutoUpdateScope = types.StringValue("stack")
	} else {
		data.AutoUpdateScope = types.StringValue("service")
	}
	data.PollUpdatesEnabled = types.BoolValue(stack.Config.PollForUpdates)
	data.AlertsEnabled = types.BoolValue(stack.Config.SendAlerts)
	data.Registry = &RegistryConfigModel{
		Provider: strOrNull(stack.Config.RegistryProvider),
		Account:  strOrNull(stack.Config.RegistryAccount),
	}
	data.ComposeCmdWrapper = strOrNull(stack.Config.ComposeCmdWrapper)

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
	dsFilePaths, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.FilePaths)
	data.Source = &StackSourceModel{
		RepoID:        strOrNull(stack.Config.LinkedRepo),
		URL:           urlVal,
		AccountID:     strOrNull(stack.Config.GitAccount),
		Path:          strOrNull(stack.Config.Repo),
		Branch:        strOrNull(stack.Config.Branch),
		Commit:        strOrNull(stack.Config.Commit),
		CloneEnforced: types.BoolValue(stack.Config.Reclone),
	}
	data.Files = &FilesConfigModel{
		Contents:     strOrNull(stack.Config.FileContents),
		LocalEnabled: types.BoolValue(stack.Config.FilesOnHost),
		Directory:    strOrNull(stack.Config.RunDirectory),
		Paths:        dsFilePaths,
	}

	data.Webhook = &StackDataSourceWebhookModel{
		Enabled:     types.BoolValue(stack.Config.WebhookEnabled),
		Secret:      strOrNull(stack.Config.WebhookSecret),
		ForceDeploy: types.BoolValue(stack.Config.WebhookForceDeploy),
	}

	data.PreDeploy = &SystemCommandModel{
		Path:    types.StringValue(stack.Config.PreDeploy.Path),
		Command: types.StringValue(strings.TrimRight(stack.Config.PreDeploy.Command, "\n")),
	}
	data.PostDeploy = &SystemCommandModel{
		Path:    types.StringValue(stack.Config.PostDeploy.Path),
		Command: types.StringValue(strings.TrimRight(stack.Config.PostDeploy.Command, "\n")),
	}

	envVars := envStringToMap(strings.TrimRight(stack.Config.Environment, "\n"))
	filePath := types.StringNull()
	if stack.Config.EnvFilePath != "" {
		filePath = types.StringValue(stack.Config.EnvFilePath)
	}
	data.Environment = &EnvironmentModel{
		FilePath:  filePath,
		Variables: envVars,
	}

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ExtraArgs)
	data.ExtraArguments = extraArgs

	wrapperInclude, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.ComposeCmdWrapperInclude)
	data.ComposeCmdWrapperInclude = wrapperInclude

	ignoreServices, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.IgnoreServices)
	data.IgnoreServices = ignoreServices

	links, _ := types.ListValueFrom(ctx, types.StringType, stack.Config.Links)
	data.Links = links
}
