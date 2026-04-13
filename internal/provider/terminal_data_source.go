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

var _ datasource.DataSource = &TerminalDataSource{}

func NewTerminalDataSource() datasource.DataSource {
	return &TerminalDataSource{}
}

type TerminalDataSource struct {
	client *client.Client
}

type TerminalDataSourceModel struct {
	TargetType   types.String  `tfsdk:"target_type"`
	TargetID     types.String  `tfsdk:"target_id"`
	Container    types.String  `tfsdk:"container"`
	Service      types.String  `tfsdk:"service"`
	Name         types.String  `tfsdk:"name"`
	Command      types.String  `tfsdk:"command"`
	CreatedAt    types.String  `tfsdk:"created_at"`
	StoredSizeKB types.Float64 `tfsdk:"stored_size_kb"`
}

func (d *TerminalDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminal"
}

func (d *TerminalDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Komodo terminal session on a target resource.",
		Attributes: map[string]schema.Attribute{
			"target_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of target to look up terminals on. One of `Server`, `Container`, `Stack`, `Deployment`.",
			},
			"target_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The primary target ID or name. For `Server`: the server. For `Container`: the server hosting the container. For `Stack`: the stack. For `Deployment`: the deployment.",
			},
			"container": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The container name. Only used when `target_type` is `Container`.",
			},
			"service": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The stack service name. Only used when `target_type` is `Stack`.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the terminal session to look up.",
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
		},
	}
}

func (d *TerminalDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerminalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerminalDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading terminal", map[string]interface{}{
		"target_id": data.TargetID.ValueString(),
		"name":      data.Name.ValueString(),
	})

	targetType := data.TargetType.ValueString()
	targetID := data.TargetID.ValueString()
	readTarget := client.NewTerminalTarget(targetType, targetID, data.Container.ValueString(), data.Service.ValueString())
	terminals, err := d.client.ListTerminals(ctx, client.ListTerminalsRequest{
		Target: &readTarget,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list terminals, got error: %s", err))
		return
	}

	var found *client.Terminal
	for i := range terminals {
		if terminals[i].Name == data.Name.ValueString() {
			found = &terminals[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Terminal Not Found",
			fmt.Sprintf("No terminal named %q found on target %q.", data.Name.ValueString(), targetID),
		)
		return
	}

	data.Command = types.StringValue(found.Command)
	data.CreatedAt = types.StringValue(msToRFC3339(found.CreatedAt))
	data.StoredSizeKB = types.Float64Value(found.StoredSizeKB)
	data.Container = types.StringValue(found.Target.ContainerName())
	data.Service = types.StringValue(found.Target.ServiceName())

	tflog.Trace(ctx, "Read terminal data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
