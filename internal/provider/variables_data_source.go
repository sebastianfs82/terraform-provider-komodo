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

var _ datasource.DataSource = &VariablesDataSource{}

func NewVariablesDataSource() datasource.DataSource {
	return &VariablesDataSource{}
}

type VariablesDataSource struct {
	client *client.Client
}

type VariablesDataSourceModel struct {
	Variables []VariableDataSourceModel `tfsdk:"variables"`
}

func (d *VariablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variables"
}

func (d *VariablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo variables visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"variables": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of variables.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The variable name.",
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The variable value (empty string for secret variables).",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "An optional description of the variable.",
						},
						"is_secret": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the variable is treated as a secret.",
						},
					},
				},
			},
		},
	}
}

func (d *VariablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VariablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VariablesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing variables")

	variables, err := d.client.ListVariables(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list variables, got error: %s", err))
		return
	}

	items := make([]VariableDataSourceModel, 0, len(variables))
	for _, v := range variables {
		items = append(items, VariableDataSourceModel{
			Name:        types.StringValue(v.Name),
			Value:       types.StringValue(v.Value),
			Description: types.StringValue(v.Description),
			IsSecret:    types.BoolValue(v.IsSecret),
		})
	}
	data.Variables = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
