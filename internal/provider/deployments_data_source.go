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

var _ datasource.DataSource = &DeploymentsDataSource{}

func NewDeploymentsDataSource() datasource.DataSource {
	return &DeploymentsDataSource{}
}

type DeploymentsDataSource struct {
	client *client.Client
}

type DeploymentsDataSourceModel struct {
	Deployments []DeploymentListItemModel `tfsdk:"deployments"`
}

type DeploymentListItemModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ServerID   types.String `tfsdk:"server_id"`
	Image      types.String `tfsdk:"image"`
	SendAlerts types.Bool   `tfsdk:"send_alerts"`
}

func (d *DeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *DeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo deployments visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"deployments": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of deployments.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The deployment identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the deployment.",
						},
						"server_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the server the deployment runs on.",
						},
						"image": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The container image.",
						},
						"send_alerts": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether alerts are enabled for this deployment.",
						},
					},
				},
			},
		},
	}
}

func (d *DeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Listing deployments")

	deployments, err := d.client.ListDeployments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list deployments, got error: %s", err))
		return
	}

	items := make([]DeploymentListItemModel, 0, len(deployments))
	for _, dep := range deployments {
		imageStr := ""
		if dep.Config.Image.Image != nil {
			imageStr = dep.Config.Image.Image.Image
		} else if dep.Config.Image.Build != nil {
			imageStr = dep.Config.Image.Build.BuildID
		}
		items = append(items, DeploymentListItemModel{
			ID:         types.StringValue(dep.ID.OID),
			Name:       types.StringValue(dep.Name),
			ServerID:   types.StringValue(dep.Config.ServerID),
			Image:      types.StringValue(imageStr),
			SendAlerts: types.BoolValue(dep.Config.SendAlerts),
		})
	}
	data := DeploymentsDataSourceModel{Deployments: items}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
