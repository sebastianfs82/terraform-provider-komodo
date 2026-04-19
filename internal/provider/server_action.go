// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// ─── ServerAction (unified) ───────────────────────────────────────────────────

var _ action.Action = (*ServerAction)(nil)
var _ action.ActionWithConfigure = (*ServerAction)(nil)

func NewServerAction() action.Action { return &ServerAction{} }

type ServerAction struct{ client *client.Client }

type ServerActionModel struct {
	ID     types.String `tfsdk:"id"`
	Action types.String `tfsdk:"action"`
}

func (a *ServerAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (a *ServerAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invokes any action that belongs to a Komodo server resource. " +
			"Set `action` to select the operation: `prune_buildx`, `prune_containers`, `prune_builders`, " +
			"`prune_images`, `prune_networks`, `prune_system`, or `prune_volumes`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the target server.",
			},
			"action": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The server operation to perform. Valid values: `prune_buildx`, `prune_containers`, " +
					"`prune_builders`, `prune_images`, `prune_networks`, `prune_system`, `prune_volumes`.",
			},
		},
	}
}

func (a *ServerAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = miscActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ID.ValueString()
	act := data.Action.ValueString()

	tflog.Debug(ctx, "Executing unified server action", map[string]interface{}{"server": serverID, "action": act})

	switch act {
	case "prune_buildx":
		if err := a.client.PruneBuildx(ctx, client.PruneBuildxRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune buildx cache, got error: %s", err))
		}
	case "prune_containers":
		if err := a.client.PruneContainers(ctx, client.PruneContainersRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune containers, got error: %s", err))
		}
	case "prune_builders":
		if err := a.client.PruneDockerBuilders(ctx, client.PruneDockerBuildersRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune docker builders, got error: %s", err))
		}
	case "prune_images":
		if err := a.client.PruneImages(ctx, client.PruneImagesRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune images, got error: %s", err))
		}
	case "prune_networks":
		if err := a.client.PruneNetworks(ctx, client.PruneNetworksRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune networks, got error: %s", err))
		}
	case "prune_system":
		if err := a.client.PruneSystem(ctx, client.PruneSystemRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune docker system, got error: %s", err))
		}
	case "prune_volumes":
		if err := a.client.PruneVolumes(ctx, client.PruneVolumesRequest{Server: serverID}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune volumes, got error: %s", err))
		}
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Unknown server action %q. Valid values are: prune_buildx, prune_containers, prune_builders, prune_images, prune_networks, prune_system, prune_volumes.", act),
		)
	}
}
