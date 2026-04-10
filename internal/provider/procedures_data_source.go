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

var _ datasource.DataSource = &ProceduresDataSource{}

func NewProceduresDataSource() datasource.DataSource {
	return &ProceduresDataSource{}
}

type ProceduresDataSource struct {
	client *client.Client
}

type ProceduresDataSourceModel struct {
	Procedures []ProcedureListItemModel `tfsdk:"procedures"`
}

type ProcedureListItemModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Schedule        types.String `tfsdk:"schedule"`
	ScheduleEnabled types.Bool   `tfsdk:"schedule_enabled"`
	WebhookEnabled  types.Bool   `tfsdk:"webhook_enabled"`
}

func (d *ProceduresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedures"
}

func (d *ProceduresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo procedures visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"procedures": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of procedures.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The procedure identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the procedure.",
						},
						"schedule": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The cron schedule for the procedure.",
						},
						"schedule_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the schedule is enabled.",
						},
						"webhook_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether webhook triggers are enabled.",
						},
					},
				},
			},
		},
	}
}

func (d *ProceduresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProceduresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProceduresDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing procedures")

	procedures, err := d.client.ListProcedures(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list procedures, got error: %s", err))
		return
	}

	items := make([]ProcedureListItemModel, 0, len(procedures))
	for _, p := range procedures {
		items = append(items, ProcedureListItemModel{
			ID:              types.StringValue(p.ID.OID),
			Name:            types.StringValue(p.Name),
			Schedule:        types.StringValue(p.Config.Schedule),
			ScheduleEnabled: types.BoolValue(p.Config.ScheduleEnabled),
			WebhookEnabled:  types.BoolValue(p.Config.WebhookEnabled),
		})
	}
	data.Procedures = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
