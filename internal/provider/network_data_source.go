// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ datasource.DataSource = &NetworkDataSource{}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

type NetworkDataSource struct {
	client *client.Client
}

type NetworkDataSourceModel struct {
	ServerID    types.String `tfsdk:"server_id"`
	Name        types.String `tfsdk:"name"`
	NetworkID   types.String `tfsdk:"network_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Scope       types.String `tfsdk:"scope"`
	Driver      types.String `tfsdk:"driver"`
	IPv6Enabled types.Bool   `tfsdk:"ipv6_enabled"`
	IPAMDriver  types.String `tfsdk:"ipam_driver"`
	IPAMSubnet  types.String `tfsdk:"ipam_subnet"`
	IPAMGateway types.String `tfsdk:"ipam_gateway"`
	Internal    types.Bool   `tfsdk:"internal"`
	Attachable  types.Bool   `tfsdk:"attachable"`
	Ingress     types.Bool   `tfsdk:"ingress"`
	InUse       types.Bool   `tfsdk:"in_use"`
}

func (d *NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing docker network on a Komodo-managed server.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The server ID or name to query networks on.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the docker network to look up.",
			},
			"network_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The docker-assigned network ID.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the network was created (RFC3339 format, e.g. `2006-01-02T15:04:05Z`).",
			},
			"scope": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The scope of the network (e.g. `local`, `swarm`).",
			},
			"driver": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The network driver (e.g. `bridge`, `overlay`).",
			},
			"ipv6_enabled": schema.BoolAttribute{
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
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading docker network", map[string]interface{}{
		"server_id": data.ServerID.ValueString(),
		"name":      data.Name.ValueString(),
	})

	networks, err := d.client.ListDockerNetworks(ctx, data.ServerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list docker networks, got error: %s", err))
		return
	}

	var found *client.NetworkListItem
	for i := range networks {
		if networks[i].Name != nil && *networks[i].Name == data.Name.ValueString() {
			found = &networks[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Not Found",
			fmt.Sprintf("Docker network %q not found on server %q", data.Name.ValueString(), data.ServerID.ValueString()),
		)
		return
	}

	if found.ID != nil {
		data.NetworkID = types.StringValue(*found.ID)
	} else {
		data.NetworkID = types.StringValue("")
	}
	if found.Created != nil {
		if t, err := time.Parse(time.RFC3339Nano, *found.Created); err == nil {
			data.CreatedAt = types.StringValue(t.UTC().Format(time.RFC3339))
		} else {
			data.CreatedAt = types.StringValue(*found.Created)
		}
	} else {
		data.CreatedAt = types.StringValue("")
	}
	if found.Scope != nil {
		data.Scope = types.StringValue(*found.Scope)
	} else {
		data.Scope = types.StringValue("")
	}
	if found.Driver != nil {
		data.Driver = types.StringValue(*found.Driver)
	} else {
		data.Driver = types.StringValue("")
	}
	if found.EnableIPv6 != nil {
		data.IPv6Enabled = types.BoolValue(*found.EnableIPv6)
	} else {
		data.IPv6Enabled = types.BoolValue(false)
	}
	if found.IPAMDriver != nil {
		data.IPAMDriver = types.StringValue(*found.IPAMDriver)
	} else {
		data.IPAMDriver = types.StringValue("")
	}
	if found.IPAMSubnet != nil {
		data.IPAMSubnet = types.StringValue(*found.IPAMSubnet)
	} else {
		data.IPAMSubnet = types.StringValue("")
	}
	if found.IPAMGateway != nil {
		data.IPAMGateway = types.StringValue(*found.IPAMGateway)
	} else {
		data.IPAMGateway = types.StringValue("")
	}
	if found.Internal != nil {
		data.Internal = types.BoolValue(*found.Internal)
	} else {
		data.Internal = types.BoolValue(false)
	}
	if found.Attachable != nil {
		data.Attachable = types.BoolValue(*found.Attachable)
	} else {
		data.Attachable = types.BoolValue(false)
	}
	if found.Ingress != nil {
		data.Ingress = types.BoolValue(*found.Ingress)
	} else {
		data.Ingress = types.BoolValue(false)
	}
	data.InUse = types.BoolValue(found.InUse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
