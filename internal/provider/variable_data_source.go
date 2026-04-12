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

var _ datasource.DataSource = &VariableDataSource{}
var _ datasource.DataSourceWithValidateConfig = &VariableDataSource{}

func NewVariableDataSource() datasource.DataSource {
	return &VariableDataSource{}
}

type VariableDataSource struct {
	client *client.Client
}

type VariableDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	Description   types.String `tfsdk:"description"`
	SecretEnabled types.Bool   `tfsdk:"secret_enabled"`
}

func (d *VariableDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

func (d *VariableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The variable identifier (same as name). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The variable name. One of `name` or `id` must be set.",
			},
			"value": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The variable value.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The variable description.",
			},
			"secret_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the variable is secret.",
			},
		},
	}
}

func (d *VariableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VariableDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data VariableDataSourceModel
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

func (d *VariableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VariableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}
	tflog.Debug(ctx, "Reading variable data source", map[string]interface{}{"lookup": lookup})

	// Find variable by name (case-insensitive)
	variables, err := d.client.ListVariables(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list variables, got error: %s", err))
		return
	}
	var variable *client.Variable
	for _, v := range variables {
		if strings.EqualFold(v.Name, lookup) {
			variable = &v
			break
		}
	}
	if variable == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Variable %q not found", lookup))
		return
	}
	data.ID = types.StringValue(variable.Name)
	data.Name = types.StringValue(variable.Name)
	data.Value = types.StringValue(variable.Value)
	data.Description = types.StringValue(variable.Description)
	data.SecretEnabled = types.BoolValue(variable.IsSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
