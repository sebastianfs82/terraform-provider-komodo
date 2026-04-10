// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &OnboardingKeyDataSource{}

func NewOnboardingKeyDataSource() datasource.DataSource {
	return &OnboardingKeyDataSource{}
}

type OnboardingKeyDataSource struct {
	client *client.Client
}

type OnboardingKeyDataSourceModel struct {
	PublicKey     types.String `tfsdk:"public_key"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Expires       types.Int64  `tfsdk:"expires"`
	Privileged    types.Bool   `tfsdk:"privileged"`
	CopyServer    types.String `tfsdk:"copy_server"`
	CreateBuilder types.Bool   `tfsdk:"create_builder"`
}

func (d *OnboardingKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_onboarding_key"
}

func (d *OnboardingKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo onboarding key by name or public key.",
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The public key used to identify the onboarding key. If set alongside name, takes precedence.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the onboarding key.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the onboarding key is enabled.",
			},
			"expires": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The expiry timestamp (Unix ms). 0 means no expiry.",
			},
			"privileged": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the onboarding key grants privileged access.",
			},
			"copy_server": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID or name of a server to copy configuration from.",
			},
			"create_builder": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to create a builder for the onboarded server.",
			},
		},
	}
}

func (d *OnboardingKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OnboardingKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnboardingKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicKey := data.PublicKey.ValueString()
	name := data.Name.ValueString()

	if publicKey == "" && name == "" {
		resp.Diagnostics.AddError("Missing Query Attribute", "Either public_key or name must be set to query an onboarding key.")
		return
	}

	var key *client.OnboardingKey
	var err error

	if publicKey != "" {
		tflog.Debug(ctx, "Looking up onboarding key by public_key", map[string]interface{}{"public_key": publicKey})
		key, err = d.client.GetOnboardingKey(ctx, publicKey)
	} else {
		tflog.Debug(ctx, "Listing onboarding keys to find by name", map[string]interface{}{"name": name})
		var keys []client.OnboardingKey
		keys, err = d.client.ListOnboardingKeys(ctx)
		if err == nil {
			for i := range keys {
				if keys[i].Name == name {
					key = &keys[i]
					break
				}
			}
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read onboarding key, got error: %s", err))
		return
	}
	if key == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Onboarding key not found (public_key=%q, name=%q)", publicKey, name))
		return
	}

	data.PublicKey = types.StringValue(key.PublicKey)
	data.Name = types.StringValue(key.Name)
	data.Enabled = types.BoolValue(key.Enabled)
	data.Expires = types.Int64Value(key.Expires)
	data.Privileged = types.BoolValue(key.Privileged)
	data.CopyServer = types.StringValue(key.CopyServer)
	data.CreateBuilder = types.BoolValue(key.CreateBuilder)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
