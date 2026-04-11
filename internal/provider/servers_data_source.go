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

var _ datasource.DataSource = &ServersDataSource{}

func NewServersDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

type ServersDataSource struct {
	client *client.Client
}

type ServersDataSourceModel struct {
	Servers []ServerDataSourceModel `tfsdk:"servers"`
}

func (d *ServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	serverAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The server identifier (ObjectId).",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the server.",
		},
		"address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The ws/s address of the periphery client.",
		},
		"tls_ignored": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether periphery TLS certificate validation is skipped.",
		},
		"external_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The address used for container links on this server.",
		},
		"region": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "An optional region label.",
		},
		"enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the server is enabled.",
		},
		"auto_rotate_keys_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether to automatically rotate server keys.",
		},
		"auto_prune_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether to run `docker image prune -a -f` every 24 hours.",
		},
		"mounts_ignored": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Mount paths filtered from system stats reports.",
		},
		"links": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Quick links displayed in the Komodo UI for this server.",
		},
		"alerts": schema.SingleNestedAttribute{
			Computed:            true,
			MarkdownDescription: "Alert configuration for this server.",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Computed:            true,
					MarkdownDescription: "Whether server stats monitoring and alerting is enabled.",
				},
				"types": schema.SetAttribute{
					Computed:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Enabled alert types.",
				},
				"thresholds": schema.SingleNestedAttribute{
					Computed:            true,
					MarkdownDescription: "Alert threshold percentages.",
					Attributes: map[string]schema.Attribute{
						"cpu_critical": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "CPU percentage threshold for CRITICAL state.",
						},
						"cpu_warning": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "CPU percentage threshold for WARNING state.",
						},
						"disk_critical": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Disk percentage threshold for CRITICAL state.",
						},
						"disk_warning": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Disk percentage threshold for WARNING state.",
						},
						"memory_critical": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Memory percentage threshold for CRITICAL state.",
						},
						"memory_warning": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Memory percentage threshold for WARNING state.",
						},
					},
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Komodo servers visible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of servers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: serverAttrs,
				},
			},
		},
	}
}

func (d *ServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing servers")

	servers, err := d.client.ListServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list servers, got error: %s", err))
		return
	}

	data.Servers = make([]ServerDataSourceModel, len(servers))
	for i := range servers {
		serverToDataSourceModel(ctx, &servers[i], &data.Servers[i])
	}

	tflog.Trace(ctx, "Listed servers", map[string]interface{}{"count": len(servers)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
