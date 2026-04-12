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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &AlerterResource{}
var _ resource.ResourceWithImportState = &AlerterResource{}
var _ resource.ResourceWithValidateConfig = &AlerterResource{}

func NewAlerterResource() resource.Resource {
	return &AlerterResource{}
}

type AlerterResource struct {
	client *client.Client
}

type AlerterResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	Name        types.String             `tfsdk:"name"`
	Tags        types.List               `tfsdk:"tags"`
	Enabled     types.Bool               `tfsdk:"enabled"`
	AlertTypes  types.List               `tfsdk:"types"`
	Resources   []ResourceTargetModel    `tfsdk:"resource"`
	Endpoint    *AlerterEndpointModel    `tfsdk:"endpoint"`
	Maintenance []MaintenanceWindowModel `tfsdk:"maintenance"`
}

type ResourceTargetModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Type    types.String `tfsdk:"type"`
	ID      types.String `tfsdk:"id"`
}

type AlerterEndpointModel struct {
	Type  types.String `tfsdk:"type"`
	URL   types.String `tfsdk:"url"`
	Email types.String `tfsdk:"email"`
}

func (r *AlerterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerter"
}

func (r *AlerterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo alerter resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The alerter identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the alerter.",
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
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the alerter is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"types": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Only send specific alert types. If empty, all alert types are sent.",
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The alerter endpoint configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The endpoint type. One of `Custom`, `Slack`, `Discord`, `Ntfy`, `Pushover`.",
					},
					"url": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The webhook or endpoint URL.",
					},
					"email": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Email address. Only valid when `type` is `Ntfy`.",
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource": schema.ListNestedBlock{
				MarkdownDescription: "Filter alerts to specific resources. Set `enabled = true` to include a resource, `enabled = false` to exclude it.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "If `true`, only send alerts for this resource (include). If `false`, never send alerts for this resource (exclude).",
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The resource type, e.g. `Server`, `Stack`, `Deployment`, `Build`, `Repo`, `Procedure`, `Action`, `Builder`, `Alerter`, `ResourceSync`.",
						},
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The name or ID of the resource.",
						},
					},
				},
			},
			"maintenance": schema.ListNestedBlock{
				MarkdownDescription: "Scheduled maintenance windows during which alerts from this alerter will be suppressed.",
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

func (r *AlerterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data AlerterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Endpoint != nil && !data.Endpoint.Email.IsNull() && !data.Endpoint.Email.IsUnknown() {
		if data.Endpoint.Type.ValueString() != "Ntfy" {
			resp.Diagnostics.AddAttributeError(
				path.Root("endpoint").AtName("email"),
				"Invalid Configuration",
				"`email` can only be set when `type` is `Ntfy`.",
			)
		}
	}
}

func (r *AlerterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AlerterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlerterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating alerter", map[string]interface{}{
		"name": data.Name.ValueString(),
	})
	configInput, diags := alerterConfigInputFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq := client.CreateAlerterRequest{
		Name:   data.Name.ValueString(),
		Config: configInput,
	}
	a, err := r.client.CreateAlerter(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alerter, got error: %s", err))
		return
	}
	if a.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Alerter creation failed: missing ID",
			"The Komodo API did not return an alerter ID. Resource cannot be tracked in state.",
		)
		return
	}
	plannedTags := data.Tags
	resp.Diagnostics.Append(alerterToModel(ctx, a, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Alerter", ID: a.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on alerter, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	tflog.Trace(ctx, "Created alerter resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlerterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlerterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := r.client.GetAlerter(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alerter, got error: %s", err))
		return
	}
	if a == nil {
		tflog.Debug(ctx, "Alerter not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(alerterToModel(ctx, a, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlerterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlerterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AlerterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	// rename if name changed
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameAlerter(ctx, client.RenameAlerterRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename alerter, got error: %s", err))
			return
		}
	}
	configInput, diags := alerterConfigInputFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq := client.UpdateAlerterRequest{
		ID:     data.ID.ValueString(),
		Config: configInput,
	}
	a, err := r.client.UpdateAlerter(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alerter, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	resp.Diagnostics.Append(alerterToModel(ctx, a, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Alerter", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on alerter, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlerterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlerterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting alerter", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteAlerter(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alerter, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted alerter resource")
}

func (r *AlerterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// alerterConfigInputFromModel converts the Terraform model into a PartialAlerterConfigInput
// suitable for the Komodo Create/Update API.
func alerterConfigInputFromModel(ctx context.Context, data *AlerterResourceModel) (client.PartialAlerterConfigInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	enabled := data.Enabled.ValueBool()
	cfg := client.PartialAlerterConfigInput{
		Enabled: &enabled,
	}

	// types
	if !data.AlertTypes.IsNull() && !data.AlertTypes.IsUnknown() {
		var alertTypes []string
		diags.Append(data.AlertTypes.ElementsAs(ctx, &alertTypes, false)...)
		if diags.HasError() {
			return cfg, diags
		}
		cfg.AlertTypes = &alertTypes
	}

	// resource blocks — split by enabled flag
	var include, exclude []client.ResourceTarget
	for _, m := range data.Resources {
		rt := client.ResourceTarget{Type: m.Type.ValueString(), ID: m.ID.ValueString()}
		if m.Enabled.ValueBool() {
			include = append(include, rt)
		} else {
			exclude = append(exclude, rt)
		}
	}
	if include == nil {
		include = []client.ResourceTarget{}
	}
	if exclude == nil {
		exclude = []client.ResourceTarget{}
	}
	cfg.Resources = &include
	cfg.ExceptResources = &exclude

	// maintenance windows
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

	// endpoint
	if data.Endpoint == nil {
		diags.AddError("Config Error", "An endpoint block is required.")
		return cfg, diags
	}
	switch data.Endpoint.Type.ValueString() {
	case "Custom":
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Custom", Params: client.CustomAlerterEndpoint{URL: data.Endpoint.URL.ValueString()}}
	case "Slack":
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Slack", Params: client.SlackAlerterEndpoint{URL: data.Endpoint.URL.ValueString()}}
	case "Discord":
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Discord", Params: client.DiscordAlerterEndpoint{URL: data.Endpoint.URL.ValueString()}}
	case "Ntfy":
		params := client.NtfyAlerterEndpoint{URL: data.Endpoint.URL.ValueString()}
		if !data.Endpoint.Email.IsNull() && !data.Endpoint.Email.IsUnknown() {
			email := data.Endpoint.Email.ValueString()
			params.Email = &email
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Ntfy", Params: params}
	case "Pushover":
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Pushover", Params: client.PushoverAlerterEndpoint{URL: data.Endpoint.URL.ValueString()}}
	default:
		diags.AddError("Config Error", fmt.Sprintf("Unknown endpoint type: %q; must be one of Custom, Slack, Discord, Ntfy, Pushover", data.Endpoint.Type.ValueString()))
	}

	return cfg, diags
}

// alerterToModel reads an Alerter API response into the Terraform resource model.
func alerterToModel(ctx context.Context, a *client.Alerter, data *AlerterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(a.ID.OID)
	data.Name = types.StringValue(a.Name)
	tagsSlice := a.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, tagsDiags := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	diags.Append(tagsDiags...)
	if diags.HasError() {
		return diags
	}
	data.Tags = tags
	data.Enabled = types.BoolValue(a.Config.Enabled)

	// types
	alertTypes, atDiags := types.ListValueFrom(ctx, types.StringType, a.Config.AlertTypes)
	diags.Append(atDiags...)
	if diags.HasError() {
		return diags
	}
	data.AlertTypes = alertTypes

	// resource blocks — merge include (enabled=true) and exclude (enabled=false)
	resources := make([]ResourceTargetModel, 0, len(a.Config.Resources)+len(a.Config.ExceptResources))
	for _, rt := range a.Config.Resources {
		resources = append(resources, ResourceTargetModel{Enabled: types.BoolValue(true), Type: types.StringValue(rt.Type), ID: types.StringValue(rt.ID)})
	}
	for _, rt := range a.Config.ExceptResources {
		resources = append(resources, ResourceTargetModel{Enabled: types.BoolValue(false), Type: types.StringValue(rt.Type), ID: types.StringValue(rt.ID)})
	}
	data.Resources = resources

	// endpoint block
	ep := &AlerterEndpointModel{
		Type:  types.StringValue(a.Config.Endpoint.Type),
		URL:   types.StringValue(""),
		Email: types.StringNull(),
	}
	switch a.Config.Endpoint.Type {
	case "Custom":
		p, err := a.Config.Endpoint.GetCustomParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Custom alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ep.URL = types.StringValue(p.URL)
		}
	case "Slack":
		p, err := a.Config.Endpoint.GetSlackParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Slack alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ep.URL = types.StringValue(p.URL)
		}
	case "Discord":
		p, err := a.Config.Endpoint.GetDiscordParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Discord alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ep.URL = types.StringValue(p.URL)
		}
	case "Ntfy":
		p, err := a.Config.Endpoint.GetNtfyParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Ntfy alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ep.URL = types.StringValue(p.URL)
			if p.Email != nil {
				ep.Email = types.StringValue(*p.Email)
			}
		}
	case "Pushover":
		p, err := a.Config.Endpoint.GetPushoverParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Pushover alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ep.URL = types.StringValue(p.URL)
		}
	default:
		diags.AddError("Unknown Endpoint Type", fmt.Sprintf("Unknown alerter endpoint type from API: %q", a.Config.Endpoint.Type))
		return diags
	}
	data.Endpoint = ep

	// Maintenance windows
	if len(a.Config.MaintenanceWindows) > 0 {
		windows := make([]MaintenanceWindowModel, len(a.Config.MaintenanceWindows))
		for i, w := range a.Config.MaintenanceWindows {
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
			if w.DayOfWeek == "" {
				windows[i].DayOfWeek = types.StringNull()
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
