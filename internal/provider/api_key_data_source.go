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

var _ datasource.DataSource = &ApiKeyDataSource{}

func NewApiKeyDataSource() datasource.DataSource {
	return &ApiKeyDataSource{}
}

type ApiKeyDataSource struct {
	client *client.Client
}

type ApiKeyDataSourceModel struct {
	Key           types.String `tfsdk:"key"`
	Name          types.String `tfsdk:"name"`
	UserID        types.String `tfsdk:"user_id"`
	ServiceUserID types.String `tfsdk:"service_user_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
}

func (d *ApiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *ApiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo API key by key ID or name.",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The API key identifier (e.g. `K-...`). If set alongside `name`, takes precedence.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The human-friendly name of the API key.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who owns this key.",
			},
			"service_user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "When set, searches within the API keys of the specified service user.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation time in RFC3339 format.",
			},
			"expires_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Expiration time in RFC3339 format. Empty string means no expiration.",
			},
		},
	}
}

func (d *ApiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := data.Key.ValueString()
	name := data.Name.ValueString()
	serviceUserID := data.ServiceUserID.ValueString()

	if keyID == "" && name == "" {
		resp.Diagnostics.AddError("Missing Query Attribute", "Either key or name must be set to look up an API key.")
		return
	}

	var apiKey *client.ApiKey
	var err error

	if serviceUserID != "" {
		tflog.Debug(ctx, "Listing API keys for service user", map[string]interface{}{"service_user_id": serviceUserID})
		var keys []client.ApiKey
		keys, err = d.client.ListApiKeysForServiceUser(ctx, serviceUserID)
		if err == nil {
			for i := range keys {
				if (keyID != "" && keys[i].Key == keyID) || (keyID == "" && keys[i].Name == name) {
					apiKey = &keys[i]
					break
				}
			}
		}
	} else if keyID != "" {
		tflog.Debug(ctx, "Looking up API key by key ID", map[string]interface{}{"key": keyID})
		apiKey, err = d.client.GetApiKey(ctx, keyID)
	} else {
		tflog.Debug(ctx, "Listing API keys to find by name", map[string]interface{}{"name": name})
		var keys []client.ApiKey
		keys, err = d.client.ListApiKeys(ctx)
		if err == nil {
			for i := range keys {
				if keys[i].Name == name {
					apiKey = &keys[i]
					break
				}
			}
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API key, got error: %s", err))
		return
	}
	if apiKey == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("API key not found (key=%q, name=%q, service_user_id=%q)", keyID, name, serviceUserID))
		return
	}

	data.Key = types.StringValue(apiKey.Key)
	data.Name = types.StringValue(apiKey.Name)
	data.UserID = types.StringValue(apiKey.UserID)
	data.CreatedAt = types.StringValue(msToRFC3339(apiKey.CreatedAt))
	data.ExpiresAt = types.StringValue(msToRFC3339(apiKey.Expires))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
