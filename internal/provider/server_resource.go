// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}
var _ resource.ResourceWithValidateConfig = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

type ServerResource struct {
	client *client.Client
}

// MaintenanceWindowModel represents a single scheduled maintenance window.
type MaintenanceWindowModel struct {
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ScheduleType    types.String `tfsdk:"schedule_type"`
	DayOfWeek       types.String `tfsdk:"day_of_week"`
	Date            types.String `tfsdk:"date"`
	Hour            types.Int64  `tfsdk:"hour"`
	Minute          types.Int64  `tfsdk:"minute"`
	DurationMinutes types.Int64  `tfsdk:"duration_minutes"`
	Timezone        types.String `tfsdk:"timezone"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

// ServerAlertsThresholdsModel holds alert threshold percentages.
type ServerAlertsThresholdsModel struct {
	CPUCritical    types.Int64 `tfsdk:"cpu_critical"`
	CPUWarning     types.Int64 `tfsdk:"cpu_warning"`
	DiskCritical   types.Int64 `tfsdk:"disk_critical"`
	DiskWarning    types.Int64 `tfsdk:"disk_warning"`
	MemoryCritical types.Int64 `tfsdk:"memory_critical"`
	MemoryWarning  types.Int64 `tfsdk:"memory_warning"`
}

// ServerAlertsModel holds the alert configuration for a server.
type ServerAlertsModel struct {
	Enabled    types.Bool                   `tfsdk:"enabled"`
	Types      types.Set                    `tfsdk:"types"`
	Thresholds *ServerAlertsThresholdsModel `tfsdk:"thresholds"`
}

type ServerResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Tags types.List   `tfsdk:"tags"`

	// Connection
	Address                        types.String `tfsdk:"address"`
	CertificateVerificationEnabled types.Bool   `tfsdk:"certificate_verification_enabled"`
	ExternalAddress                types.String `tfsdk:"external_address"`
	Region                         types.String `tfsdk:"region"`
	PublicKey                      types.String `tfsdk:"public_key"`

	// Behaviour
	Enabled                           types.Bool `tfsdk:"enabled"`
	AutoRotateKeysEnabled             types.Bool `tfsdk:"auto_rotate_keys_enabled"`
	AutoPruneImagesEnabled            types.Bool `tfsdk:"auto_prune_images_enabled"`
	HistoricalSystemStatisticsEnabled types.Bool `tfsdk:"historical_system_statistics_enabled"`

	// Mounts / links
	IgnoredDiskMounts types.List `tfsdk:"ignored_disk_mounts"`
	Links             types.List `tfsdk:"links"`

	// Alerts (nested object)
	Alerts *ServerAlertsModel `tfsdk:"alerts"`

	// Maintenance windows
	Maintenance []MaintenanceWindowModel `tfsdk:"maintenance"`
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ServerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Alerts == nil {
		return
	}
	if !data.Alerts.Enabled.IsNull() && !data.Alerts.Enabled.IsUnknown() && data.Alerts.Enabled.ValueBool() {
		if !data.Alerts.Types.IsUnknown() && (data.Alerts.Types.IsNull() || len(data.Alerts.Types.Elements()) == 0) {
			resp.Diagnostics.AddAttributeError(
				path.Root("alerts").AtName("types"),
				"Invalid alerts configuration",
				"alerts.types must not be empty when alerts.enabled is true. Specify at least one of: cpu, disk, memory, unreachable, version.",
			)
		}
	}
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
				MarkdownDescription: "The unique name of the server.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "A list of tag IDs to attach to this resource. Use `komodo_tag.<name>.id` to reference tags.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
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
			"certificate_verification_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to verify the periphery TLS certificate. When `false`, certificate validation is skipped (useful for self-signed certs). Defaults to `true`.",
			},
			"public_key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Custom public key for the Periphery agent. If provided, the associated private key must be set as Periphery `private_key`. Required for Periphery → Core connections unless `public_key` is set in Core config. Note: the API does not return this value, so external changes cannot be detected.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"auto_rotate_keys_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to automatically rotate server keys when `RotateAllServerKeys` is called. Defaults to `true`.",
			},
			"auto_prune_images_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to run `docker image prune -a -f` every 24 hours. Defaults to `false`.",
			},
			"historical_system_statistics_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether server stats monitoring is enabled. Defaults to `true`.",
			},
			"ignored_disk_mounts": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Mount paths to filter from system stats reports. Defaults to `[]`.",
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Quick links displayed in the Komodo UI for this server.",
			},
		},
		Blocks: map[string]schema.Block{
			"alerts": schema.SingleNestedBlock{
				MarkdownDescription: "Alert configuration for this server.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether alerting is enabled. When `false`, all alert types are disabled regardless of `types`. Defaults to `true`.",
					},
					"types": schema.SetAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Alert types to enable when `enabled` is `true`. Valid values: `cpu`, `disk`, `memory`, `unreachable`, `version`. Required (non-empty) when `enabled` is `true`. Defaults to `[]`.",
					},
				},
				Blocks: map[string]schema.Block{
					"thresholds": schema.SingleNestedBlock{
						MarkdownDescription: "Alert threshold percentages.",
						Attributes: map[string]schema.Attribute{
							"cpu_critical": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(99),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "CPU percentage threshold for CRITICAL state. Must be between 0 and 100. Defaults to `99`.",
							},
							"cpu_warning": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(90),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "CPU percentage threshold for WARNING state. Must be between 0 and 100. Defaults to `90`.",
							},
							"disk_critical": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(95),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "Disk percentage threshold for CRITICAL state. Must be between 0 and 100. Defaults to `95`.",
							},
							"disk_warning": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(75),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "Disk percentage threshold for WARNING state. Must be between 0 and 100. Defaults to `75`.",
							},
							"memory_critical": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(95),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "Memory percentage threshold for CRITICAL state. Must be between 0 and 100. Defaults to `95`.",
							},
							"memory_warning": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(75),
								Validators:          []validator.Int64{int64PercentValidator{}},
								MarkdownDescription: "Memory percentage threshold for WARNING state. Must be between 0 and 100. Defaults to `75`.",
							},
						},
					},
				},
			},
			"maintenance": schema.ListNestedBlock{
				MarkdownDescription: "Scheduled maintenance windows during which alerts from this server will be suppressed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Name for the maintenance window.",
						},
						"description": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Description of what maintenance is performed.",
						},
						"schedule_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Schedule type: `Daily`, `Weekly`, or `OneTime`.",
						},
						"day_of_week": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "For `Weekly` schedules: day of the week (e.g. `Monday`, `Tuesday`).",
						},
						"date": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "For `OneTime` windows: ISO 8601 date in `YYYY-MM-DD` format.",
						},
						"hour": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							MarkdownDescription: "Start hour in 24-hour format (0–23). Defaults to `0`.",
						},
						"minute": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							MarkdownDescription: "Start minute (0–59). Defaults to `0`.",
						},
						"duration_minutes": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Duration of the maintenance window in minutes.",
						},
						"timezone": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Timezone for the maintenance window. If empty, uses the Core timezone.",
						},
						"enabled": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							MarkdownDescription: "Whether this maintenance window is active. Defaults to `true`.",
						},
					},
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

	var pubKey *string
	if !data.PublicKey.IsNull() && !data.PublicKey.IsUnknown() && data.PublicKey.ValueString() != "" {
		pubKey = client.StringPtr(data.PublicKey.ValueString())
	}
	server, err := r.client.CreateServer(ctx, client.CreateServerRequest{
		Name:      data.Name.ValueString(),
		Config:    cfg,
		PublicKey: pubKey,
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

	plannedTags := data.Tags
	resp.Diagnostics.Append(serverToResourceModel(ctx, server, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Server", ID: server.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on server, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
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
	var state ServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameServer(ctx, client.RenameServerRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename server, got error: %s", err))
			return
		}
	}

	cfg, diags := serverConfigFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.PublicKey.IsNull() && !data.PublicKey.IsUnknown() && data.PublicKey.ValueString() != "" {
		if err := r.client.UpdateServerPublicKey(ctx, data.ID.ValueString(), data.PublicKey.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server public key, got error: %s", err))
			return
		}
	}
	server, err := r.client.UpdateServer(ctx, client.UpdateServerRequest{
		ID:     data.ID.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server, got error: %s", err))
		return
	}
	// UpdateServer uses _PartialServerConfig (all Option<T> fields). The response
	// may not fully reflect merged state, so re-read to get the authoritative values.
	server, err = r.client.GetServer(ctx, server.ID.OID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server after update, got error: %s", err))
		return
	}
	if server == nil {
		resp.Diagnostics.AddError("Client Error", "Server not found after update")
		return
	}

	plannedTags := data.Tags
	resp.Diagnostics.Append(serverToResourceModel(ctx, server, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Server", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on server, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
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

// serverConfigFromModel converts the resource model to a PartialServerConfig for API calls.
// Only fields that are known (not null/unknown) are set so the API applies its own defaults
// for omitted fields instead of receiving an explicit false/zero.
func serverConfigFromModel(ctx context.Context, data *ServerResourceModel) (client.PartialServerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var cfg client.PartialServerConfig

	if !data.Address.IsNull() && !data.Address.IsUnknown() {
		cfg.Address = client.StringPtr(data.Address.ValueString())
	}
	if !data.CertificateVerificationEnabled.IsNull() && !data.CertificateVerificationEnabled.IsUnknown() {
		cfg.InsecureTLS = client.BoolPtr(!data.CertificateVerificationEnabled.ValueBool())
	}
	if !data.ExternalAddress.IsNull() && !data.ExternalAddress.IsUnknown() {
		cfg.ExternalAddress = client.StringPtr(data.ExternalAddress.ValueString())
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		cfg.Region = client.StringPtr(data.Region.ValueString())
	}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		cfg.Enabled = client.BoolPtr(data.Enabled.ValueBool())
	}
	if !data.AutoRotateKeysEnabled.IsNull() && !data.AutoRotateKeysEnabled.IsUnknown() {
		cfg.AutoRotateKeys = client.BoolPtr(data.AutoRotateKeysEnabled.ValueBool())
	}
	if !data.AutoPruneImagesEnabled.IsNull() && !data.AutoPruneImagesEnabled.IsUnknown() {
		cfg.AutoPrune = client.BoolPtr(data.AutoPruneImagesEnabled.ValueBool())
	}

	if !data.IgnoredDiskMounts.IsNull() && !data.IgnoredDiskMounts.IsUnknown() {
		var mounts []string
		diags.Append(data.IgnoredDiskMounts.ElementsAs(ctx, &mounts, false)...)
		cfg.IgnoreMounts = &mounts
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		diags.Append(data.Links.ElementsAs(ctx, &links, false)...)
		cfg.Links = &links
	}
	if !data.HistoricalSystemStatisticsEnabled.IsNull() && !data.HistoricalSystemStatisticsEnabled.IsUnknown() {
		cfg.StatsMonitoring = client.BoolPtr(data.HistoricalSystemStatisticsEnabled.ValueBool())
	}

	if data.Alerts != nil {
		if !data.Alerts.Enabled.IsNull() && !data.Alerts.Enabled.IsUnknown() {
			if !data.Alerts.Enabled.ValueBool() {
				// alerts.enabled = false → disable all send_*_alerts, ignore types
				cfg.SendCPUAlerts = client.BoolPtr(false)
				cfg.SendDiskAlerts = client.BoolPtr(false)
				cfg.SendMemAlerts = client.BoolPtr(false)
				cfg.SendUnreachableAlerts = client.BoolPtr(false)
				cfg.SendVersionMismatchAlerts = client.BoolPtr(false)
			} else {
				var alertTypes []string
				diags.Append(data.Alerts.Types.ElementsAs(ctx, &alertTypes, false)...)
				wantCPU, wantDisk, wantMem, wantUnreachable, wantVersion := false, false, false, false, false
				for _, t := range alertTypes {
					switch t {
					case "cpu":
						wantCPU = true
					case "disk":
						wantDisk = true
					case "memory":
						wantMem = true
					case "unreachable":
						wantUnreachable = true
					case "version":
						wantVersion = true
					}
				}
				cfg.SendCPUAlerts = client.BoolPtr(wantCPU)
				cfg.SendDiskAlerts = client.BoolPtr(wantDisk)
				cfg.SendMemAlerts = client.BoolPtr(wantMem)
				cfg.SendUnreachableAlerts = client.BoolPtr(wantUnreachable)
				cfg.SendVersionMismatchAlerts = client.BoolPtr(wantVersion)
			}
		}

		if data.Alerts.Thresholds != nil {
			th := data.Alerts.Thresholds
			if !th.CPUCritical.IsNull() && !th.CPUCritical.IsUnknown() {
				cfg.CPUCritical = client.Float64Ptr(float64(th.CPUCritical.ValueInt64()))
			}
			if !th.CPUWarning.IsNull() && !th.CPUWarning.IsUnknown() {
				cfg.CPUWarning = client.Float64Ptr(float64(th.CPUWarning.ValueInt64()))
			}
			if !th.DiskCritical.IsNull() && !th.DiskCritical.IsUnknown() {
				cfg.DiskCritical = client.Float64Ptr(float64(th.DiskCritical.ValueInt64()))
			}
			if !th.DiskWarning.IsNull() && !th.DiskWarning.IsUnknown() {
				cfg.DiskWarning = client.Float64Ptr(float64(th.DiskWarning.ValueInt64()))
			}
			if !th.MemoryCritical.IsNull() && !th.MemoryCritical.IsUnknown() {
				cfg.MemCritical = client.Float64Ptr(float64(th.MemoryCritical.ValueInt64()))
			}
			if !th.MemoryWarning.IsNull() && !th.MemoryWarning.IsUnknown() {
				cfg.MemWarning = client.Float64Ptr(float64(th.MemoryWarning.ValueInt64()))
			}
		}
	}

	// Maintenance windows
	if data.Maintenance != nil {
		windows := make([]client.MaintenanceWindow, len(data.Maintenance))
		for i, w := range data.Maintenance {
			windows[i] = client.MaintenanceWindow{
				Name:            w.Name.ValueString(),
				Description:     w.Description.ValueString(),
				ScheduleType:    w.ScheduleType.ValueString(),
				DayOfWeek:       w.DayOfWeek.ValueString(),
				Date:            w.Date.ValueString(),
				Hour:            w.Hour.ValueInt64(),
				Minute:          w.Minute.ValueInt64(),
				DurationMinutes: w.DurationMinutes.ValueInt64(),
				Timezone:        w.Timezone.ValueString(),
				Enabled:         w.Enabled.ValueBool(),
			}
		}
		cfg.MaintenanceWindows = &windows
	}

	return cfg, diags
}

// serverToResourceModel maps a client.Server to the resource model.
func serverToResourceModel(ctx context.Context, s *client.Server, data *ServerResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(s.ID.OID)
	data.Name = types.StringValue(s.Name)
	tagsSlice := s.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, tagsDiags := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	diags.Append(tagsDiags...)
	if diags.HasError() {
		return diags
	}
	data.Tags = tags

	cfg := s.Config
	data.Address = types.StringValue(cfg.Address)
	data.CertificateVerificationEnabled = types.BoolValue(!cfg.InsecureTLS)
	data.ExternalAddress = types.StringValue(cfg.ExternalAddress)
	data.Region = types.StringValue(cfg.Region)
	// public_key is not returned by the API; preserve any known value from state/plan,
	// but resolve unknown (e.g. first apply with no prior state) to null.
	if data.PublicKey.IsUnknown() {
		data.PublicKey = types.StringNull()
	}
	data.Enabled = types.BoolValue(cfg.Enabled)
	data.AutoRotateKeysEnabled = types.BoolValue(cfg.AutoRotateKeys)
	data.AutoPruneImagesEnabled = types.BoolValue(cfg.AutoPrune)
	data.HistoricalSystemStatisticsEnabled = types.BoolValue(cfg.StatsMonitoring)

	if cfg.IgnoreMounts != nil {
		listVal, d := types.ListValueFrom(ctx, types.StringType, cfg.IgnoreMounts)
		diags.Append(d...)
		data.IgnoredDiskMounts = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		data.IgnoredDiskMounts = listVal
	}

	if cfg.Links != nil {
		listVal, d := types.ListValueFrom(ctx, types.StringType, cfg.Links)
		diags.Append(d...)
		data.Links = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		data.Links = listVal
	}

	// Build alerts.types from individual API boolean flags.
	anyAlertEnabled := cfg.SendCPUAlerts || cfg.SendDiskAlerts || cfg.SendMemAlerts || cfg.SendUnreachableAlerts || cfg.SendVersionMismatchAlerts
	var typesSet types.Set
	if anyAlertEnabled {
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
		var d diag.Diagnostics
		typesSet, d = types.SetValueFrom(ctx, types.StringType, alertTypes)
		diags.Append(d...)
	} else if data.Alerts != nil && !data.Alerts.Types.IsNull() && !data.Alerts.Types.IsUnknown() {
		// All alert flags are disabled on the API. Preserve the configured types so
		// the state stays consistent with the plan (alerts.enabled = false with types
		// is valid config; types are the "desired" set to restore when re-enabled).
		typesSet = data.Alerts.Types
	} else {
		var d diag.Diagnostics
		typesSet, d = types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
	}
	if data.Alerts != nil {
		data.Alerts = &ServerAlertsModel{
			Enabled: types.BoolValue(anyAlertEnabled),
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

	// Maintenance windows
	if len(cfg.MaintenanceWindows) > 0 {
		windows := make([]MaintenanceWindowModel, len(cfg.MaintenanceWindows))
		for i, w := range cfg.MaintenanceWindows {
			windows[i] = MaintenanceWindowModel{
				Name:            types.StringValue(w.Name),
				Description:     types.StringValue(w.Description),
				ScheduleType:    types.StringValue(w.ScheduleType),
				DayOfWeek:       types.StringValue(w.DayOfWeek),
				Date:            types.StringValue(w.Date),
				Hour:            types.Int64Value(w.Hour),
				Minute:          types.Int64Value(w.Minute),
				DurationMinutes: types.Int64Value(w.DurationMinutes),
				Timezone:        types.StringValue(w.Timezone),
				Enabled:         types.BoolValue(w.Enabled),
			}
			// Fix: Map empty string fields to null for optional attributes
			if w.Description == "" {
				windows[i].Description = types.StringNull()
			}
			if w.Date == "" {
				windows[i].Date = types.StringNull()
			}
			if w.Timezone == "" {
				windows[i].Timezone = types.StringNull()
			}
		}
		data.Maintenance = windows
	} else if data.Maintenance != nil && len(data.Maintenance) == 0 {
		data.Maintenance = []MaintenanceWindowModel{}
	} else {
		data.Maintenance = nil
	}

	return diags
}

// int64PercentValidator enforces that an int64 attribute value is between 0 and 100 inclusive.
type int64PercentValidator struct{}

func (v int64PercentValidator) Description(_ context.Context) string {
	return "Value must be an integer between 0 and 100."
}

func (v int64PercentValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v int64PercentValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val < 0 || val > 100 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid percentage value",
			fmt.Sprintf("Expected a value between 0 and 100, got: %d.", val),
		)
	}
}
