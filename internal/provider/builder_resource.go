// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &BuilderResource{}
var _ resource.ResourceWithImportState = &BuilderResource{}

func NewBuilderResource() resource.Resource {
	return &BuilderResource{}
}

type BuilderResource struct {
	client *client.Client
}

type BuilderResourceModel struct {
	ID           types.String       `tfsdk:"id"`
	Name         types.String       `tfsdk:"name"`
	BuilderType  types.String       `tfsdk:"builder_type"`
	UrlConfig    *UrlConfigModel    `tfsdk:"url_config"`
	ServerConfig *ServerConfigModel `tfsdk:"server_config"`
	AwsConfig    *AwsConfigModel    `tfsdk:"aws_config"`
}

type UrlConfigModel struct {
	Address            types.String `tfsdk:"address"`
	PeripheryPublicKey types.String `tfsdk:"periphery_public_key"`
	InsecureTls        types.Bool   `tfsdk:"insecure_tls"`
	Passkey            types.String `tfsdk:"passkey"`
}

type ServerConfigModel struct {
	ServerID types.String `tfsdk:"server_id"`
}

type AwsConfigModel struct {
	Region             types.String `tfsdk:"region"`
	InstanceType       types.String `tfsdk:"instance_type"`
	VolumeGb           types.Int64  `tfsdk:"volume_gb"`
	AmiID              types.String `tfsdk:"ami_id"`
	SubnetID           types.String `tfsdk:"subnet_id"`
	KeyPairName        types.String `tfsdk:"key_pair_name"`
	AssignPublicIP     types.Bool   `tfsdk:"assign_public_ip"`
	UsePublicIP        types.Bool   `tfsdk:"use_public_ip"`
	SecurityGroupIDs   types.List   `tfsdk:"security_group_ids"`
	UserData           types.String `tfsdk:"user_data"`
	Port               types.Int64  `tfsdk:"port"`
	UseHttps           types.Bool   `tfsdk:"use_https"`
	PeripheryPublicKey types.String `tfsdk:"periphery_public_key"`
	InsecureTls        types.Bool   `tfsdk:"insecure_tls"`
	Secrets            types.List   `tfsdk:"secrets"`
}

func (r *BuilderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_builder"
}

func (r *BuilderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo builder resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The builder identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the builder. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"builder_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The builder type. One of `Url`, `Server`, `Aws`. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a URL builder. Required when `builder_type` is `Url`.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The address of the Periphery agent.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"periphery_public_key": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "An expected public key associated with the Periphery private key. If empty, the public key is not validated.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"insecure_tls": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to skip TLS certificate validation for the Periphery connection.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"passkey": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Deprecated. Optional passkey for Periphery authentication.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"server_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a server builder. Required when `builder_type` is `Server`.",
				Attributes: map[string]schema.Attribute{
					"server_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The ID of the Komodo server to use as the builder.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"aws_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for an AWS builder. Required when `builder_type` is `Aws`.",
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The AWS region to create the instance in.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"instance_type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The EC2 instance type to use for the build.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"volume_gb": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The size of the builder volume in GiB.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"ami_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The EC2 AMI ID to use. The AMI must have the Periphery client configured to start at boot.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"subnet_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The subnet ID to create the instance in.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"key_pair_name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The key pair name to attach to the instance.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"assign_public_ip": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to assign the instance a public IP address.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"use_public_ip": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether Komodo core should communicate with the instance using its public IP address.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group_ids": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The security group IDs to attach to the instance.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"user_data": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The user data script to run when the instance starts.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"port": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The port that Periphery is running on. Default: `8120`.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"use_https": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to use HTTPS to communicate with the Periphery agent.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"periphery_public_key": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "An expected public key associated with the Periphery private key. If empty, the public key is not validated.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"insecure_tls": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to skip TLS certificate validation for the Periphery connection.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"secrets": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The secret names available on the AMI.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *BuilderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BuilderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BuilderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating builder", map[string]interface{}{
		"name":         data.Name.ValueString(),
		"builder_type": data.BuilderType.ValueString(),
	})
	configInput, err := builderConfigInputFromModel(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Config Error", fmt.Sprintf("Unable to build builder config: %s", err))
		return
	}
	createReq := client.CreateBuilderRequest{
		Name:   data.Name.ValueString(),
		Config: configInput,
	}
	b, err := r.client.CreateBuilder(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create builder, got error: %s", err))
		return
	}
	if b.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Builder creation failed: missing ID",
			"The Komodo API did not return a builder ID. Resource cannot be tracked in state.",
		)
		return
	}
	resp.Diagnostics.Append(builderToModel(ctx, b, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "Created builder resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuilderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BuilderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	b, err := r.client.GetBuilder(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read builder, got error: %s", err))
		return
	}
	if b == nil {
		tflog.Debug(ctx, "Builder not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(builderToModel(ctx, b, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuilderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BuilderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state BuilderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	configInput, err := builderConfigInputFromModel(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Config Error", fmt.Sprintf("Unable to build builder config: %s", err))
		return
	}
	updateReq := client.UpdateBuilderRequest{
		ID:     data.ID.ValueString(),
		Config: configInput,
	}
	b, err := r.client.UpdateBuilder(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update builder, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(builderToModel(ctx, b, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuilderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BuilderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting builder", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteBuilder(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete builder, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted builder resource")
}

func (r *BuilderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// builderConfigInputFromModel converts the Terraform model into a BuilderConfigInput
// suitable for the Komodo Create/Update API.
func builderConfigInputFromModel(ctx context.Context, data *BuilderResourceModel) (client.BuilderConfigInput, error) {
	switch data.BuilderType.ValueString() {
	case "Url":
		if data.UrlConfig == nil {
			return client.BuilderConfigInput{}, fmt.Errorf("url_config is required when builder_type is Url")
		}
		return client.BuilderConfigInput{
			Type: "Url",
			Params: client.UrlBuilderConfig{
				Address:            data.UrlConfig.Address.ValueString(),
				PeripheryPublicKey: data.UrlConfig.PeripheryPublicKey.ValueString(),
				InsecureTls:        data.UrlConfig.InsecureTls.ValueBool(),
				Passkey:            data.UrlConfig.Passkey.ValueString(),
			},
		}, nil
	case "Server":
		if data.ServerConfig == nil {
			return client.BuilderConfigInput{}, fmt.Errorf("server_config is required when builder_type is Server")
		}
		return client.BuilderConfigInput{
			Type: "Server",
			Params: client.ServerBuilderConfig{
				ServerID: data.ServerConfig.ServerID.ValueString(),
			},
		}, nil
	case "Aws":
		if data.AwsConfig == nil {
			return client.BuilderConfigInput{}, fmt.Errorf("aws_config is required when builder_type is Aws")
		}
		var securityGroupIDs []string
		if !data.AwsConfig.SecurityGroupIDs.IsNull() && !data.AwsConfig.SecurityGroupIDs.IsUnknown() {
			data.AwsConfig.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)
		}
		var secrets []string
		if !data.AwsConfig.Secrets.IsNull() && !data.AwsConfig.Secrets.IsUnknown() {
			data.AwsConfig.Secrets.ElementsAs(ctx, &secrets, false)
		}
		return client.BuilderConfigInput{
			Type: "Aws",
			Params: client.AwsBuilderConfig{
				Region:             data.AwsConfig.Region.ValueString(),
				InstanceType:       data.AwsConfig.InstanceType.ValueString(),
				VolumeGb:           data.AwsConfig.VolumeGb.ValueInt64(),
				AmiID:              data.AwsConfig.AmiID.ValueString(),
				SubnetID:           data.AwsConfig.SubnetID.ValueString(),
				KeyPairName:        data.AwsConfig.KeyPairName.ValueString(),
				AssignPublicIP:     data.AwsConfig.AssignPublicIP.ValueBool(),
				UsePublicIP:        data.AwsConfig.UsePublicIP.ValueBool(),
				SecurityGroupIDs:   securityGroupIDs,
				UserData:           data.AwsConfig.UserData.ValueString(),
				Port:               data.AwsConfig.Port.ValueInt64(),
				UseHttps:           data.AwsConfig.UseHttps.ValueBool(),
				PeripheryPublicKey: data.AwsConfig.PeripheryPublicKey.ValueString(),
				InsecureTls:        data.AwsConfig.InsecureTls.ValueBool(),
				Secrets:            secrets,
			},
		}, nil
	default:
		return client.BuilderConfigInput{}, fmt.Errorf("unknown builder_type: %q; must be one of Url, Server, Aws", data.BuilderType.ValueString())
	}
}

// builderToModel reads a Builder API response into the Terraform resource model.
func builderToModel(ctx context.Context, b *client.Builder, data *BuilderResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(b.ID.OID)
	data.Name = types.StringValue(b.Name)
	data.BuilderType = types.StringValue(b.Config.Type)

	switch b.Config.Type {
	case "Url":
		urlCfg, err := b.Config.GetUrlConfig()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode URL builder config: %s", err))
			return diags
		}
		if urlCfg != nil {
			data.UrlConfig = &UrlConfigModel{
				Address:            types.StringValue(urlCfg.Address),
				PeripheryPublicKey: types.StringValue(urlCfg.PeripheryPublicKey),
				InsecureTls:        types.BoolValue(urlCfg.InsecureTls),
				Passkey:            types.StringValue(urlCfg.Passkey),
			}
		}
		data.ServerConfig = nil
		data.AwsConfig = nil
	case "Server":
		serverCfg, err := b.Config.GetServerConfig()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Server builder config: %s", err))
			return diags
		}
		if serverCfg != nil {
			data.ServerConfig = &ServerConfigModel{
				ServerID: types.StringValue(serverCfg.ServerID),
			}
		}
		data.UrlConfig = nil
		data.AwsConfig = nil
	case "Aws":
		awsCfg, err := b.Config.GetAwsConfig()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode AWS builder config: %s", err))
			return diags
		}
		if awsCfg != nil {
			securityGroupIDs, sgDiags := types.ListValueFrom(ctx, types.StringType, awsCfg.SecurityGroupIDs)
			diags.Append(sgDiags...)
			if diags.HasError() {
				return diags
			}
			secrets, secretsDiags := types.ListValueFrom(ctx, types.StringType, awsCfg.Secrets)
			diags.Append(secretsDiags...)
			if diags.HasError() {
				return diags
			}
			data.AwsConfig = &AwsConfigModel{
				Region:             types.StringValue(awsCfg.Region),
				InstanceType:       types.StringValue(awsCfg.InstanceType),
				VolumeGb:           types.Int64Value(awsCfg.VolumeGb),
				AmiID:              types.StringValue(awsCfg.AmiID),
				SubnetID:           types.StringValue(awsCfg.SubnetID),
				KeyPairName:        types.StringValue(awsCfg.KeyPairName),
				AssignPublicIP:     types.BoolValue(awsCfg.AssignPublicIP),
				UsePublicIP:        types.BoolValue(awsCfg.UsePublicIP),
				SecurityGroupIDs:   securityGroupIDs,
				UserData:           types.StringValue(awsCfg.UserData),
				Port:               types.Int64Value(awsCfg.Port),
				UseHttps:           types.BoolValue(awsCfg.UseHttps),
				PeripheryPublicKey: types.StringValue(awsCfg.PeripheryPublicKey),
				InsecureTls:        types.BoolValue(awsCfg.InsecureTls),
				Secrets:            secrets,
			}
		}
		data.UrlConfig = nil
		data.ServerConfig = nil
	default:
		diags.AddError("Unknown Builder Type", fmt.Sprintf("Unknown builder type from API: %q", b.Config.Type))
	}

	return diags
}
