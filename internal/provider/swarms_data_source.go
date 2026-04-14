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

var _ datasource.DataSource = &SwarmsDataSource{}

func NewSwarmsDataSource() datasource.DataSource {
	return &SwarmsDataSource{}
}

type SwarmsDataSource struct {
	client *client.Client
}

type SwarmsDataSourceModel struct {
	Swarms []SwarmListItemModel `tfsdk:"swarms"`
}

// SwarmListItemModel is the per-item model for the swarms list data source.
// Maintenance windows are excluded from list results (use komodo_swarm for full details).
type SwarmListItemModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Tags          types.List   `tfsdk:"tags"`
	ServerIDs     types.List   `tfsdk:"server_ids"`
	Links         types.List   `tfsdk:"links"`
	AlertsEnabled types.Bool   `tfsdk:"alerts_enabled"`
}

func (d *SwarmsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_swarms"
}

func (d *SwarmsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	swarmAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The swarm identifier (ObjectId).",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the swarm.",
		},
		"tags": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Tag IDs attached to this swarm.",
		},
		"server_ids": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "IDs of the servers that are manager nodes of this swarm.",
		},
		"links": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Quick links displayed in the Komodo UI for this swarm.",
		},
		"alerts_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether alerts are sent when the swarm is unhealthy.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo Docker Swarms visible to the authenticated user. Use `komodo_swarm` for full details including maintenance windows.",
		Attributes: map[string]schema.Attribute{
			"swarms": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of swarms.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: swarmAttrs,
				},
			},
		},
	}
}

func (d *SwarmsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SwarmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SwarmsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing swarms")

	swarms, err := d.client.ListSwarms(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list swarms, got error: %s", err))
		return
	}

	data.Swarms = make([]SwarmListItemModel, len(swarms))
	for i := range swarms {
		s := &swarms[i]
		item := &data.Swarms[i]
		item.ID = types.StringValue(s.ID.OID)
		item.Name = types.StringValue(s.Name)
		tagsSlice := s.Tags
		if tagsSlice == nil {
			tagsSlice = []string{}
		}
		item.Tags, _ = types.ListValueFrom(ctx, types.StringType, tagsSlice)
		if s.Config.ServerIDs != nil {
			item.ServerIDs, _ = types.ListValueFrom(ctx, types.StringType, s.Config.ServerIDs)
		} else {
			item.ServerIDs, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		}
		if s.Config.Links != nil {
			item.Links, _ = types.ListValueFrom(ctx, types.StringType, s.Config.Links)
		} else {
			item.Links, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		}
		item.AlertsEnabled = types.BoolValue(s.Config.AlertsEnabled)
	}

	tflog.Trace(ctx, "Listed swarms", map[string]interface{}{"count": len(swarms)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
