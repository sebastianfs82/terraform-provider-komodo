// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

type ServerResource struct {
	client *client.Client
}

type ServerResourceModel struct {
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

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The server identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the server. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ws/s address of the periphery client. If unset, server expects Periphery → Core connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"insecure_tls": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to skip periphery TLS certificate validation. Defaults to `true` because periphery generates self-signed certificates by default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"external_address": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The address used for container links on this server. If empty, uses `address`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "An optional region label.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the server is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_rotate_keys": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to automatically rotate server keys when `RotateAllServerKeys` is called.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_prune": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to run `docker image prune -a -f` every 24 hours.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"stats_monitoring": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to monitor server stats beyond health checks.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_mounts": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Mount paths to filter from system stats reports.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links displayed in the Komodo UI for this server.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"send_unreachable_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server reachability.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_cpu_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server CPU status.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_mem_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server memory status.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_disk_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about server disk status.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_version_mismatch_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send alerts about version mismatches with core.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu_warning": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CPU percentage threshold for WARNING state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"cpu_critical": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CPU percentage threshold for CRITICAL state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"mem_warning": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Memory percentage threshold for WARNING state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"mem_critical": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Memory percentage threshold for CRITICAL state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"disk_warning": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disk percentage threshold for WARNING state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"disk_critical": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disk percentage threshold for CRITICAL state.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating server", map[string]interface{}{"name": data.Name.ValueString()})

	cfg, diags := serverConfigFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.CreateServer(ctx, client.CreateServerRequest{
		Name:   data.Name.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server, got error: %s", err))
		return
	}
	if server.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Server creation failed: missing ID",
			"The Komodo API did not return a server ID. Resource cannot be tracked in state.",
		)
		return
	}

	resp.Diagnostics.Append(serverToResourceModel(ctx, server, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Created server resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server, got error: %s", err))
		return
	}
	if server == nil {
		tflog.Debug(ctx, "Server not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(serverToResourceModel(ctx, server, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := serverConfigFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.UpdateServer(ctx, client.UpdateServerRequest{
		ID:     data.ID.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(serverToResourceModel(ctx, server, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting server", map[string]interface{}{"id": data.ID.ValueString()})

	if err := r.client.DeleteServer(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete server, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted server resource")
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// serverConfigFromModel converts the resource model to a ServerConfig for API calls.
func serverConfigFromModel(ctx context.Context, data *ServerResourceModel) (client.ServerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	cfg := client.ServerConfig{
		Address:                   data.Address.ValueString(),
		InsecureTLS:               data.InsecureTLS.ValueBool(),
		ExternalAddress:           data.ExternalAddress.ValueString(),
		Region:                    data.Region.ValueString(),
		Enabled:                   data.Enabled.ValueBool(),
		AutoRotateKeys:            data.AutoRotateKeys.ValueBool(),
		AutoPrune:                 data.AutoPrune.ValueBool(),
		StatsMonitoring:           data.StatsMonitoring.ValueBool(),
		SendUnreachableAlerts:     data.SendUnreachableAlerts.ValueBool(),
		SendCPUAlerts:             data.SendCPUAlerts.ValueBool(),
		SendMemAlerts:             data.SendMemAlerts.ValueBool(),
		SendDiskAlerts:            data.SendDiskAlerts.ValueBool(),
		SendVersionMismatchAlerts: data.SendVersionMismatchAlerts.ValueBool(),
		CPUWarning:                data.CPUWarning.ValueFloat64(),
		CPUCritical:               data.CPUCritical.ValueFloat64(),
		MemWarning:                data.MemWarning.ValueFloat64(),
		MemCritical:               data.MemCritical.ValueFloat64(),
		DiskWarning:               data.DiskWarning.ValueFloat64(),
		DiskCritical:              data.DiskCritical.ValueFloat64(),
	}

	if !data.IgnoreMounts.IsNull() && !data.IgnoreMounts.IsUnknown() {
		var mounts []string
		diags.Append(data.IgnoreMounts.ElementsAs(ctx, &mounts, false)...)
		cfg.IgnoreMounts = mounts
	}

	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		diags.Append(data.Links.ElementsAs(ctx, &links, false)...)
		cfg.Links = links
	}

	return cfg, diags
}

// serverToResourceModel maps a client.Server to the resource model.
func serverToResourceModel(ctx context.Context, s *client.Server, data *ServerResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

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

	if cfg.IgnoreMounts != nil {
		elems := make([]types.String, len(cfg.IgnoreMounts))
		for i, v := range cfg.IgnoreMounts {
			elems[i] = types.StringValue(v)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, elems)
		diags.Append(d...)
		data.IgnoreMounts = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []types.String{})
		diags.Append(d...)
		data.IgnoreMounts = listVal
	}

	if cfg.Links != nil {
		elems := make([]types.String, len(cfg.Links))
		for i, v := range cfg.Links {
			elems[i] = types.StringValue(v)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, elems)
		diags.Append(d...)
		data.Links = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []types.String{})
		diags.Append(d...)
		data.Links = listVal
	}

	return diags
}
