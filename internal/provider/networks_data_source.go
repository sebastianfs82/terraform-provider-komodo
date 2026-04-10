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

var _ datasource.DataSource = &NetworksDataSource{}

func NewNetworksDataSource() datasource.DataSource {
	return &NetworksDataSource{}
}

type NetworksDataSource struct {
	client *client.Client
}

type NetworksDataSourceModel struct {
	ServerID types.String             `tfsdk:"server_id"`
	Networks []NetworkDataSourceModel `tfsdk:"networks"`
}

func (d *NetworksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *NetworksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	networkAttrs := map[string]schema.Attribute{
		"server_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The server ID or name the network belongs to.",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the docker network.",
		},
		"network_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The docker-assigned network ID.",
		},
		"created": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The timestamp when the network was created.",
		},
		"scope": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The scope of the network (e.g. `local`, `swarm`).",
		},
		"driver": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The network driver (e.g. `bridge`, `overlay`).",
		},
		"enable_ipv6": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether IPv6 is enabled on the network.",
		},
		"ipam_driver": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPAM driver used by the network.",
		},
		"ipam_subnet": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPAM subnet configured for the network.",
		},
		"ipam_gateway": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPAM gateway configured for the network.",
		},
		"internal": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the network is internal (not connected to the external network).",
		},
		"attachable": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether manual container attachment is allowed.",
		},
		"ingress": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the network is an ingress network (used for swarm routing mesh).",
		},
		"in_use": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the network is currently attached to one or more containers.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all docker networks on a Komodo-managed server.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The server ID or name to list networks on.",
			},
			"networks": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of docker networks on the server.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: networkAttrs,
				},
			},
		},
	}
}

func (d *NetworksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ServerID.ValueString()
	tflog.Debug(ctx, "Listing docker networks", map[string]interface{}{"server_id": serverID})

	networks, err := d.client.ListDockerNetworks(ctx, serverID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list docker networks, got error: %s", err))
		return
	}

	data.Networks = make([]NetworkDataSourceModel, len(networks))
	for i, n := range networks {
		item := &data.Networks[i]
		item.ServerID = types.StringValue(serverID)

		if n.Name != nil {
			item.Name = types.StringValue(*n.Name)
		} else {
			item.Name = types.StringValue("")
		}
		if n.ID != nil {
			item.NetworkID = types.StringValue(*n.ID)
		} else {
			item.NetworkID = types.StringValue("")
		}
		if n.Created != nil {
			item.Created = types.StringValue(*n.Created)
		} else {
			item.Created = types.StringValue("")
		}
		if n.Scope != nil {
			item.Scope = types.StringValue(*n.Scope)
		} else {
			item.Scope = types.StringValue("")
		}
		if n.Driver != nil {
			item.Driver = types.StringValue(*n.Driver)
		} else {
			item.Driver = types.StringValue("")
		}
		if n.EnableIPv6 != nil {
			item.EnableIPv6 = types.BoolValue(*n.EnableIPv6)
		} else {
			item.EnableIPv6 = types.BoolValue(false)
		}
		if n.IPAMDriver != nil {
			item.IPAMDriver = types.StringValue(*n.IPAMDriver)
		} else {
			item.IPAMDriver = types.StringValue("")
		}
		if n.IPAMSubnet != nil {
			item.IPAMSubnet = types.StringValue(*n.IPAMSubnet)
		} else {
			item.IPAMSubnet = types.StringValue("")
		}
		if n.IPAMGateway != nil {
			item.IPAMGateway = types.StringValue(*n.IPAMGateway)
		} else {
			item.IPAMGateway = types.StringValue("")
		}
		if n.Internal != nil {
			item.Internal = types.BoolValue(*n.Internal)
		} else {
			item.Internal = types.BoolValue(false)
		}
		if n.Attachable != nil {
			item.Attachable = types.BoolValue(*n.Attachable)
		} else {
			item.Attachable = types.BoolValue(false)
		}
		if n.Ingress != nil {
			item.Ingress = types.BoolValue(*n.Ingress)
		} else {
			item.Ingress = types.BoolValue(false)
		}
		item.InUse = types.BoolValue(n.InUse)
	}

	tflog.Trace(ctx, "Listed docker networks", map[string]interface{}{"count": len(networks)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
