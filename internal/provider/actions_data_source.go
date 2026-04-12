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

var _ datasource.DataSource = &ActionsDataSource{}

func NewActionsDataSource() datasource.DataSource {
	return &ActionsDataSource{}
}

type ActionsDataSource struct {
	client *client.Client
}

type ActionsDataSourceModel struct {
	Actions []ActionListItemModel `tfsdk:"actions"`
}

type ActionListItemModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	FileContents        types.String `tfsdk:"file_contents"`
	RunOnStartupEnabled types.Bool   `tfsdk:"run_on_startup_enabled"`
	Schedule            types.String `tfsdk:"schedule"`
	ScheduleEnabled     types.Bool   `tfsdk:"schedule_enabled"`
	WebhookEnabled      types.Bool   `tfsdk:"webhook_enabled"`
}

func (d *ActionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_actions"
}

func (d *ActionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo actions visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"actions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of actions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The action identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the action.",
						},
						"file_contents": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The Deno TypeScript file contents of the action.",
						},
						"run_on_startup_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the action runs at Komodo startup.",
						},
						"schedule": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The cron schedule for the action.",
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

func (d *ActionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ActionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing actions")

	actions, err := d.client.ListActions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list actions, got error: %s", err))
		return
	}

	items := make([]ActionListItemModel, 0, len(actions))
	for _, a := range actions {
		items = append(items, ActionListItemModel{
			ID:                  types.StringValue(a.ID.OID),
			Name:                types.StringValue(a.Name),
			FileContents:        types.StringValue(a.Config.FileContents),
			RunOnStartupEnabled: types.BoolValue(a.Config.RunAtStartup),
			Schedule:            types.StringValue(a.Config.Schedule),
			ScheduleEnabled:     types.BoolValue(a.Config.ScheduleEnabled),
			WebhookEnabled:      types.BoolValue(a.Config.WebhookEnabled),
		})
	}
	data.Actions = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
