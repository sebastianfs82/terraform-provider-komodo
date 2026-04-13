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

var _ datasource.DataSource = &ServerDataSource{}
var _ datasource.DataSourceWithValidateConfig = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

type ServerDataSource struct {
	client *client.Client
}

type ServerDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	// Connection
	Address         types.String `tfsdk:"address"`
	TLSIgnored      types.Bool   `tfsdk:"tls_ignored"`
	ExternalAddress types.String `tfsdk:"external_address"`
	Region          types.String `tfsdk:"region"`

	// Behaviour
	Enabled               types.Bool `tfsdk:"enabled"`
	AutoRotateKeysEnabled types.Bool `tfsdk:"auto_rotate_keys_enabled"`
	AutoPruneEnabled      types.Bool `tfsdk:"auto_prune_enabled"`

	// Mounts / links
	MountsIgnored types.List `tfsdk:"mounts_ignored"`
	Links         types.List `tfsdk:"links"`

	// Alerts
	Alerts *ServerAlertsModel `tfsdk:"alerts"`
}

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Komodo server by name or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The server identifier (ObjectId). One of `name` or `id` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The server name. One of `name` or `id` must be set.",
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
		},
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data ServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nameSet := !data.Name.IsNull() && !data.Name.IsUnknown()
	idSet := !data.ID.IsNull() && !data.ID.IsUnknown()
	// If either value is unknown (e.g. referenced from another data source),
	// skip validation — it will be enforced during Read once values are known.
	if data.Name.IsUnknown() || data.ID.IsUnknown() {
		return
	}
	if nameSet && idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of `name` or `id` may be set, not both.",
		)
		return
	}
	if !nameSet && !idSet {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of `name` or `id` must be set.",
		)
	}
}

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lookup := data.Name.ValueString()
	if lookup == "" {
		lookup = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Reading server", map[string]interface{}{"lookup": lookup})

	server, err := d.client.GetServer(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server, got error: %s", err))
		return
	}
	if server == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Server %q not found", lookup))
		return
	}

	serverToDataSourceModel(ctx, server, &data)

	tflog.Trace(ctx, "Read server data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func serverToDataSourceModel(ctx context.Context, s *client.Server, data *ServerDataSourceModel) {
	data.ID = types.StringValue(s.ID.OID)
	data.Name = types.StringValue(s.Name)

	cfg := s.Config
	data.Address = types.StringValue(cfg.Address)
	data.TLSIgnored = types.BoolValue(cfg.InsecureTLS)
	data.ExternalAddress = types.StringValue(cfg.ExternalAddress)
	data.Region = types.StringValue(cfg.Region)
	data.Enabled = types.BoolValue(cfg.Enabled)
	data.AutoRotateKeysEnabled = types.BoolValue(cfg.AutoRotateKeys)
	data.AutoPruneEnabled = types.BoolValue(cfg.AutoPrune)

	if cfg.IgnoreMounts != nil {
		data.MountsIgnored, _ = types.ListValueFrom(ctx, types.StringType, cfg.IgnoreMounts)
	} else {
		data.MountsIgnored, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	if cfg.Links != nil {
		data.Links, _ = types.ListValueFrom(ctx, types.StringType, cfg.Links)
	} else {
		data.Links, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	var alertTypes []string
	if cfg.SendCPUAlerts {
		alertTypes = append(alertTypes, "cpu")
	}
	if cfg.SendDiskAlerts {
		alertTypes = append(alertTypes, "disk")
	}
	if cfg.SendMemAlerts {
		alertTypes = append(alertTypes, "memory")
	}
	if cfg.SendUnreachableAlerts {
		alertTypes = append(alertTypes, "unreachable")
	}
	if cfg.SendVersionMismatchAlerts {
		alertTypes = append(alertTypes, "version")
	}
	typesSet, _ := types.SetValueFrom(ctx, types.StringType, alertTypes)

	data.Alerts = &ServerAlertsModel{
		Enabled: types.BoolValue(cfg.StatsMonitoring),
		Types:   typesSet,
		Thresholds: &ServerAlertsThresholdsModel{
			CPUCritical:    types.Int64Value(int64(cfg.CPUCritical)),
			CPUWarning:     types.Int64Value(int64(cfg.CPUWarning)),
			DiskCritical:   types.Int64Value(int64(cfg.DiskCritical)),
			DiskWarning:    types.Int64Value(int64(cfg.DiskWarning)),
			MemoryCritical: types.Int64Value(int64(cfg.MemCritical)),
			MemoryWarning:  types.Int64Value(int64(cfg.MemWarning)),
		},
	}
}
