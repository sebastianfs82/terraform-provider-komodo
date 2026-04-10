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

var _ datasource.DataSource = &BuildsDataSource{}

func NewBuildsDataSource() datasource.DataSource {
	return &BuildsDataSource{}
}

type BuildsDataSource struct {
	client *client.Client
}

type BuildsDataSourceModel struct {
	RepoID    types.String         `tfsdk:"repo_id"`
	BuilderID types.String         `tfsdk:"builder_id"`
	Builds    []BuildListItemModel `tfsdk:"builds"`
}

type BuildListItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	BuilderID      types.String `tfsdk:"builder_id"`
	ImageName      types.String `tfsdk:"image_name"`
	ImageTag       types.String `tfsdk:"image_tag"`
	Repo           types.String `tfsdk:"repo"`
	Branch         types.String `tfsdk:"branch"`
	WebhookEnabled types.Bool   `tfsdk:"webhook_enabled"`
}

func (d *BuildsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_builds"
}

func (d *BuildsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo builds visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"repo_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter builds by linked repo ID. When set, only builds sourced from this repo are returned.",
			},
			"builder_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter builds by builder ID. When set, only builds using this builder are returned.",
			},
			"builds": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of builds.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The build identifier (ObjectId).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the build.",
						},
						"builder_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the builder used by this build.",
						},
						"image_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The target image name.",
						},
						"image_tag": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The target image tag.",
						},
						"repo": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git repository path (owner/repo).",
						},
						"branch": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The git branch.",
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

func (d *BuildsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BuildsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BuildsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing builds")

	builds, err := d.client.ListBuilds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list builds, got error: %s", err))
		return
	}

	items := make([]BuildListItemModel, 0, len(builds))
	for _, b := range builds {
		if !data.RepoID.IsNull() && !data.RepoID.IsUnknown() && b.Config.LinkedRepo != data.RepoID.ValueString() {
			continue
		}
		if !data.BuilderID.IsNull() && !data.BuilderID.IsUnknown() && b.Config.BuilderID != data.BuilderID.ValueString() {
			continue
		}
		items = append(items, BuildListItemModel{
			ID:             types.StringValue(b.ID.OID),
			Name:           types.StringValue(b.Name),
			BuilderID:      types.StringValue(b.Config.BuilderID),
			ImageName:      types.StringValue(b.Config.ImageName),
			ImageTag:       types.StringValue(b.Config.ImageTag),
			Repo:           types.StringValue(b.Config.Repo),
			Branch:         types.StringValue(b.Config.Branch),
			WebhookEnabled: types.BoolValue(b.Config.WebhookEnabled),
		})
	}
	data.Builds = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
