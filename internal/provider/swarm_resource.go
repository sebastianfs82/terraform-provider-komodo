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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &SwarmResource{}
var _ resource.ResourceWithImportState = &SwarmResource{}

func NewSwarmResource() resource.Resource {
	return &SwarmResource{}
}

type SwarmResource struct {
	client *client.Client
}

type SwarmResourceModel struct {
	ID            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	Tags          types.List               `tfsdk:"tags"`
	ServerIDs     types.List               `tfsdk:"server_ids"`
	Links         types.List               `tfsdk:"links"`
	AlertsEnabled types.Bool               `tfsdk:"alerts_enabled"`
	Maintenance   []MaintenanceWindowModel `tfsdk:"maintenance"`
}

func (r *SwarmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_swarm"
}

func (r *SwarmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo Docker Swarm.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The swarm identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the swarm.",
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
			"server_ids": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "IDs of the servers that are manager nodes of this swarm. Defaults to `[]`.",
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Quick links displayed in the Komodo UI for this swarm. Defaults to `[]`.",
			},
			"alerts_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to send alerts when the swarm is unhealthy. Defaults to `true`.",
			},
		},
		Blocks: map[string]schema.Block{
			"maintenance": schema.ListNestedBlock{
				MarkdownDescription: "Scheduled maintenance windows during which alerts from this swarm will be suppressed.",
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

func (r *SwarmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SwarmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SwarmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating swarm", map[string]interface{}{"name": data.Name.ValueString()})

	cfg, diags := swarmConfigFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	swarm, err := r.client.CreateSwarm(ctx, client.CreateSwarmRequest{
		Name:   data.Name.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create swarm, got error: %s", err))
		return
	}
	if swarm.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Swarm creation failed: missing ID",
			"The Komodo API did not return a swarm ID. Resource cannot be tracked in state.",
		)
		return
	}

	plannedTags := data.Tags
	resp.Diagnostics.Append(swarmToResourceModel(ctx, swarm, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Swarm", ID: swarm.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on swarm, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}

	tflog.Trace(ctx, "Created swarm resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwarmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SwarmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	swarm, err := r.client.GetSwarm(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read swarm, got error: %s", err))
		return
	}
	if swarm == nil {
		tflog.Debug(ctx, "Swarm not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(swarmToResourceModel(ctx, swarm, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwarmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SwarmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SwarmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameSwarm(ctx, client.RenameSwarmRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename swarm, got error: %s", err))
			return
		}
	}

	cfg, diags := swarmConfigFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	swarm, err := r.client.UpdateSwarm(ctx, client.UpdateSwarmRequest{
		ID:     state.ID.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update swarm, got error: %s", err))
		return
	}
	swarm, err = r.client.GetSwarm(ctx, swarm.ID.OID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read swarm after update, got error: %s", err))
		return
	}
	if swarm == nil {
		resp.Diagnostics.AddError("Client Error", "Swarm not found after update")
		return
	}

	plannedTags := data.Tags
	resp.Diagnostics.Append(swarmToResourceModel(ctx, swarm, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Swarm", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on swarm, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwarmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SwarmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting swarm", map[string]interface{}{"id": data.ID.ValueString()})

	if err := r.client.DeleteSwarm(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete swarm, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted swarm resource")
}

func (r *SwarmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// swarmConfigFromModel converts the resource model to a PartialSwarmConfig for API calls.
func swarmConfigFromModel(ctx context.Context, data *SwarmResourceModel) (client.PartialSwarmConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var cfg client.PartialSwarmConfig

	if !data.ServerIDs.IsNull() && !data.ServerIDs.IsUnknown() {
		var ids []string
		diags.Append(data.ServerIDs.ElementsAs(ctx, &ids, false)...)
		cfg.ServerIDs = &ids
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		diags.Append(data.Links.ElementsAs(ctx, &links, false)...)
		cfg.Links = &links
	}
	if !data.AlertsEnabled.IsNull() && !data.AlertsEnabled.IsUnknown() {
		cfg.AlertsEnabled = client.BoolPtr(data.AlertsEnabled.ValueBool())
	}

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

// swarmToResourceModel maps a client.Swarm to the resource model.
func swarmToResourceModel(ctx context.Context, s *client.Swarm, data *SwarmResourceModel) diag.Diagnostics {
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

	if cfg.ServerIDs != nil {
		listVal, d := types.ListValueFrom(ctx, types.StringType, cfg.ServerIDs)
		diags.Append(d...)
		data.ServerIDs = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		data.ServerIDs = listVal
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

	data.AlertsEnabled = types.BoolValue(cfg.AlertsEnabled)

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
