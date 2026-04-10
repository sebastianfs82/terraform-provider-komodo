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
	InsecureTLS     types.Bool   `tfsdk:"insecure_tls"`
	ExternalAddress types.String `tfsdk:"external_address"`
	Region          types.String `tfsdk:"region"`

	// Behaviour
	Enabled         types.Bool `tfsdk:"enabled"`
	AutoRotateKeys  types.Bool `tfsdk:"auto_rotate_keys"`
	AutoPrune       types.Bool `tfsdk:"auto_prune"`
	StatsMonitoring types.Bool `tfsdk:"stats_monitoring"`

	// Mounts / links
	IgnoreMounts types.List `tfsdk:"ignore_mounts"`
	Links        types.List `tfsdk:"links"`

	// Alert flags
	SendUnreachableAlerts     types.Bool `tfsdk:"send_unreachable_alerts"`
	SendCPUAlerts             types.Bool `tfsdk:"send_cpu_alerts"`
	SendMemAlerts             types.Bool `tfsdk:"send_mem_alerts"`
	SendDiskAlerts            types.Bool `tfsdk:"send_disk_alerts"`
	SendVersionMismatchAlerts types.Bool `tfsdk:"send_version_mismatch_alerts"`

	// Alert thresholds
	CPUWarning   types.Float64 `tfsdk:"cpu_warning"`
	CPUCritical  types.Float64 `tfsdk:"cpu_critical"`
	MemWarning   types.Float64 `tfsdk:"mem_warning"`
	MemCritical  types.Float64 `tfsdk:"mem_critical"`
	DiskWarning  types.Float64 `tfsdk:"disk_warning"`
	DiskCritical types.Float64 `tfsdk:"disk_critical"`
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
			"insecure_tls": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to skip periphery TLS certificate validation.",
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
			"auto_rotate_keys": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to automatically rotate server keys.",
			},
			"auto_prune": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to run `docker image prune -a -f` every 24 hours.",
			},
			"stats_monitoring": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to monitor server stats beyond health checks.",
			},
			"ignore_mounts": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Mount paths to filter from system stats reports.",
			},
			"links": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links displayed in the Komodo UI for this server.",
			},
			"send_unreachable_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server reachability.",
			},
			"send_cpu_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server CPU status.",
			},
			"send_mem_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server memory status.",
			},
			"send_disk_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server disk status.",
			},
			"send_version_mismatch_alerts": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about version mismatches with core.",
			},
			"cpu_warning": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "CPU percentage threshold for WARNING state.",
			},
			"cpu_critical": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "CPU percentage threshold for CRITICAL state.",
			},
			"mem_warning": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Memory percentage threshold for WARNING state.",
			},
			"mem_critical": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Memory percentage threshold for CRITICAL state.",
			},
			"disk_warning": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Disk percentage threshold for WARNING state.",
			},
			"disk_critical": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Disk percentage threshold for CRITICAL state.",
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
	data.InsecureTLS = types.BoolValue(cfg.InsecureTLS)
	data.ExternalAddress = types.StringValue(cfg.ExternalAddress)
	data.Region = types.StringValue(cfg.Region)
	data.Enabled = types.BoolValue(cfg.Enabled)
	data.AutoRotateKeys = types.BoolValue(cfg.AutoRotateKeys)
	data.AutoPrune = types.BoolValue(cfg.AutoPrune)
	data.StatsMonitoring = types.BoolValue(cfg.StatsMonitoring)

	if cfg.IgnoreMounts != nil {
		elems := make([]types.String, len(cfg.IgnoreMounts))
		for i, v := range cfg.IgnoreMounts {
			elems[i] = types.StringValue(v)
		}
		listVal, _ := types.ListValueFrom(ctx, types.StringType, elems)
		data.IgnoreMounts = listVal
	} else {
		data.IgnoreMounts, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	if cfg.Links != nil {
		elems := make([]types.String, len(cfg.Links))
		for i, v := range cfg.Links {
			elems[i] = types.StringValue(v)
		}
		listVal, _ := types.ListValueFrom(ctx, types.StringType, elems)
		data.Links = listVal
	} else {
		data.Links, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	data.SendUnreachableAlerts = types.BoolValue(cfg.SendUnreachableAlerts)
	data.SendCPUAlerts = types.BoolValue(cfg.SendCPUAlerts)
	data.SendMemAlerts = types.BoolValue(cfg.SendMemAlerts)
	data.SendDiskAlerts = types.BoolValue(cfg.SendDiskAlerts)
	data.SendVersionMismatchAlerts = types.BoolValue(cfg.SendVersionMismatchAlerts)

	data.CPUWarning = types.Float64Value(cfg.CPUWarning)
	data.CPUCritical = types.Float64Value(cfg.CPUCritical)
	data.MemWarning = types.Float64Value(cfg.MemWarning)
	data.MemCritical = types.Float64Value(cfg.MemCritical)
	data.DiskWarning = types.Float64Value(cfg.DiskWarning)
	data.DiskCritical = types.Float64Value(cfg.DiskCritical)
}
