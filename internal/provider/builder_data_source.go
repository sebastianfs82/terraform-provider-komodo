// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &BuilderDataSource{}
var _ datasource.DataSourceWithValidateConfig = &BuilderDataSource{}

func NewBuilderDataSource() datasource.DataSource {
	return &BuilderDataSource{}
}

type BuilderDataSource struct {
	client *client.Client
}

type BuilderDataSourceModel struct {
	ID           types.String       `tfsdk:"id"`
	Name         types.String       `tfsdk:"name"`
	BuilderType  types.String       `tfsdk:"builder_type"`
	UrlConfig    *UrlConfigModel    `tfsdk:"url_config"`
	ServerConfig *ServerConfigModel `tfsdk:"server_config"`
	AwsConfig    *AwsConfigModel    `tfsdk:"aws_config"`
}

func (d *BuilderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_builder"
}

func (d *BuilderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo builder resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The builder identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The builder name. One of `name` or `id` must be set.",
			},
			"builder_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The builder type: `Url`, `Server`, or `Aws`.",
			},
			"url_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a URL builder. Populated when `builder_type` is `Url`.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The address of the Periphery agent.",
					},
					"periphery_public_key": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "An expected public key associated with the Periphery private key.",
					},
					"insecure_tls": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether TLS certificate validation is skipped.",
					},
					"passkey": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Deprecated. Optional passkey for Periphery authentication.",
					},
				},
			},
			"server_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for a server builder. Populated when `builder_type` is `Server`.",
				Attributes: map[string]schema.Attribute{
					"server_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The ID of the Komodo server used as the builder.",
					},
				},
			},
			"aws_config": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Configuration for an AWS builder. Populated when `builder_type` is `Aws`.",
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS region.",
					},
					"instance_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The EC2 instance type.",
					},
					"volume_gb": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The volume size in GiB.",
					},
					"ami_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The EC2 AMI ID.",
					},
					"subnet_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The subnet ID.",
					},
					"key_pair_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The key pair name.",
					},
					"assign_public_ip": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the instance is assigned a public IP.",
					},
					"use_public_ip": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether to use the public IP for Komodo-to-Periphery communication.",
					},
					"security_group_ids": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The security group IDs.",
					},
					"user_data": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The user data script.",
					},
					"port": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The Periphery port.",
					},
					"use_https": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether HTTPS is used for the Periphery connection.",
					},
					"periphery_public_key": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "An expected public key associated with the Periphery private key.",
					},
					"insecure_tls": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether TLS certificate validation is skipped.",
					},
					"secrets": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The secret names available on the AMI.",
					},
				},
			},
		},
	}
}

func (d *BuilderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BuilderDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data BuilderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// If either value is unknown (referenced from another resource not yet applied),
	// skip validation — it will be re-checked once values are known.
	if data.Name.IsUnknown() || data.ID.IsUnknown() {
		return
	}
	nameSet := !data.Name.IsNull()
	idSet := !data.ID.IsNull()
	if nameSet && idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of `name` or `id` may be set, not both.",
		)
		return
	}
	if !nameSet && !idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of `name` or `id` must be set.",
		)
	}
}

func (d *BuilderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BuilderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading builder data source", map[string]interface{}{"lookup": lookup})

	b, err := d.client.GetBuilder(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read builder, got error: %s", err))
		return
	}
	if b == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Builder %q not found.", lookup))
		return
	}

	// Reuse the resource model helper via an adapter.
	resourceModel := &BuilderResourceModel{}
	resp.Diagnostics.Append(builderToModel(ctx, b, resourceModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = resourceModel.ID
	data.Name = resourceModel.Name
	data.BuilderType = resourceModel.BuilderType
	data.UrlConfig = resourceModel.UrlConfig
	data.ServerConfig = resourceModel.ServerConfig
	data.AwsConfig = resourceModel.AwsConfig

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
