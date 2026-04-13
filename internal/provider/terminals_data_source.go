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

var _ datasource.DataSource = &TerminalsDataSource{}

func NewTerminalsDataSource() datasource.DataSource {
	return &TerminalsDataSource{}
}

type TerminalsDataSource struct {
	client *client.Client
}

type TerminalsDataSourceModel struct {
	Terminals []TerminalDataSourceModel `tfsdk:"terminals"`
}

func (d *TerminalsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminals"
}

func (d *TerminalsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	terminalAttrs := map[string]schema.Attribute{
		"target_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The type of target the terminal belongs to.",
		},
		"target_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The primary target ID (server, stack, or deployment).",
		},
		"container": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The container name (only set when type is Container).",
		},
		"service": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The stack service name (only set when type is Stack).",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the terminal session.",
		},
		"command": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The shell command used to initialise the terminal.",
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Creation timestamp in RFC3339 format.",
		},
		"stored_size_kb": schema.Float64Attribute{
			Computed:            true,
			MarkdownDescription: "The size of stored terminal output in kilobytes.",
		},
	}

	desc := "Lists all Komodo terminal sessions registered in the core."
	resp.Schema = schema.Schema{
		MarkdownDescription: desc,
		Attributes: map[string]schema.Attribute{
			"terminals": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of terminal sessions on the target.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: terminalAttrs,
				},
			},
		},
	}
}

func (d *TerminalsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerminalsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerminalsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing all terminals", nil)

	terminals, err := d.client.ListTerminals(ctx, client.ListTerminalsRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list terminals, got error: %s", err))
		return
	}

	items := make([]TerminalDataSourceModel, len(terminals))
	for i, t := range terminals {
		items[i] = TerminalDataSourceModel{
			TargetType:   types.StringValue(t.Target.Type),
			TargetID:     types.StringValue(t.Target.TargetID()),
			Container:    types.StringValue(t.Target.ContainerName()),
			Service:      types.StringValue(t.Target.ServiceName()),
			Name:         types.StringValue(t.Name),
			Command:      types.StringValue(t.Command),
			CreatedAt:    types.StringValue(msToRFC3339(t.CreatedAt)),
			StoredSizeKB: types.Float64Value(t.StoredSizeKB),
		}
	}
	data.Terminals = items

	tflog.Trace(ctx, "Listed terminals")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
