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

// ─── shared helper ───────────────────────────────────────────────────────────

func serverActionConfigure(providerData any, addError func(string, string)) *client.Client {
	if providerData == nil {
		return nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		addError("Unexpected Action Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}
	return c
}

// ─── shared model ────────────────────────────────────────────────────────────

type ServerPruneModel struct {
	Server types.String `tfsdk:"server"`
}

// ─── PruneBuildx ─────────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneBuildxAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneBuildxAction)(nil)

func NewServerPruneBuildxAction() action.Action { return &ServerPruneBuildxAction{} }

type ServerPruneBuildxAction struct{ client *client.Client }

func (a *ServerPruneBuildxAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_buildx"
}

func (a *ServerPruneBuildxAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker buildx cache on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune the buildx cache.",
			},
		},
	}
}

func (a *ServerPruneBuildxAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneBuildxAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneBuildx", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneBuildx(ctx, client.PruneBuildxRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune buildx cache, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneBuildx action completed")
}

// ─── PruneContainers ─────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneContainersAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneContainersAction)(nil)

func NewServerPruneContainersAction() action.Action { return &ServerPruneContainersAction{} }

type ServerPruneContainersAction struct{ client *client.Client }

func (a *ServerPruneContainersAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_containers"
}

func (a *ServerPruneContainersAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker containers on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune containers.",
			},
		},
	}
}

func (a *ServerPruneContainersAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneContainersAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneContainers", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneContainers(ctx, client.PruneContainersRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune containers, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneContainers action completed")
}

// ─── PruneBuilders ───────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneBuildersAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneBuildersAction)(nil)

func NewServerPruneBuildersAction() action.Action { return &ServerPruneBuildersAction{} }

type ServerPruneBuildersAction struct{ client *client.Client }

func (a *ServerPruneBuildersAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_builders"
}

func (a *ServerPruneBuildersAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker builders on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune docker builders.",
			},
		},
	}
}

func (a *ServerPruneBuildersAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneBuildersAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneDockerBuilders", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneDockerBuilders(ctx, client.PruneDockerBuildersRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune docker builders, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneDockerBuilders action completed")
}

// ─── PruneImages ─────────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneImagesAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneImagesAction)(nil)

func NewServerPruneImagesAction() action.Action { return &ServerPruneImagesAction{} }

type ServerPruneImagesAction struct{ client *client.Client }

func (a *ServerPruneImagesAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_images"
}

func (a *ServerPruneImagesAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker images on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune images.",
			},
		},
	}
}

func (a *ServerPruneImagesAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneImagesAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneImages", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneImages(ctx, client.PruneImagesRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune images, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneImages action completed")
}

// ─── PruneNetworks ───────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneNetworksAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneNetworksAction)(nil)

func NewServerPruneNetworksAction() action.Action { return &ServerPruneNetworksAction{} }

type ServerPruneNetworksAction struct{ client *client.Client }

func (a *ServerPruneNetworksAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_networks"
}

func (a *ServerPruneNetworksAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker networks on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune networks.",
			},
		},
	}
}

func (a *ServerPruneNetworksAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneNetworksAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneNetworks", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneNetworks(ctx, client.PruneNetworksRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune networks, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneNetworks action completed")
}

// ─── PruneSystem ─────────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneSystemAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneSystemAction)(nil)

func NewServerPruneSystemAction() action.Action { return &ServerPruneSystemAction{} }

type ServerPruneSystemAction struct{ client *client.Client }

func (a *ServerPruneSystemAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_system"
}

func (a *ServerPruneSystemAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker system on the target server, including volumes.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune the docker system.",
			},
		},
	}
}

func (a *ServerPruneSystemAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneSystemAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneSystem", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneSystem(ctx, client.PruneSystemRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune docker system, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneSystem action completed")
}

// ─── PruneVolumes ────────────────────────────────────────────────────────────

var _ action.Action = (*ServerPruneVolumesAction)(nil)
var _ action.ActionWithConfigure = (*ServerPruneVolumesAction)(nil)

func NewServerPruneVolumesAction() action.Action { return &ServerPruneVolumesAction{} }

type ServerPruneVolumesAction struct{ client *client.Client }

func (a *ServerPruneVolumesAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_prune_volumes"
}

func (a *ServerPruneVolumesAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Prunes the docker volumes on the target server.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id or name of the server on which to prune volumes.",
			},
		},
	}
}

func (a *ServerPruneVolumesAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.client = serverActionConfigure(req.ProviderData, func(s1, s2 string) {
		resp.Diagnostics.AddError(s1, s2)
	})
}

func (a *ServerPruneVolumesAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerPruneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Executing PruneVolumes", map[string]interface{}{"server": data.Server.ValueString()})
	if err := a.client.PruneVolumes(ctx, client.PruneVolumesRequest{Server: data.Server.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to prune volumes, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "PruneVolumes action completed")
}
