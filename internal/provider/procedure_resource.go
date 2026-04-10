// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ProcedureResource{}
var _ resource.ResourceWithImportState = &ProcedureResource{}

func NewProcedureResource() resource.Resource {
	return &ProcedureResource{}
}

type ProcedureResource struct {
	client *client.Client
}

// ProcedureResourceModel is the Terraform resource model for komodo_procedure.
type ProcedureResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Stages           types.String `tfsdk:"stages"`
	ScheduleFormat   types.String `tfsdk:"schedule_format"`
	Schedule         types.String `tfsdk:"schedule"`
	ScheduleEnabled  types.Bool   `tfsdk:"schedule_enabled"`
	ScheduleTimezone types.String `tfsdk:"schedule_timezone"`
	ScheduleAlert    types.Bool   `tfsdk:"schedule_alert"`
	FailureAlert     types.Bool   `tfsdk:"failure_alert"`
	WebhookEnabled   types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret    types.String `tfsdk:"webhook_secret"`
}

func (r *ProcedureResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (r *ProcedureResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo procedure resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The procedure identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the procedure. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"stages": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "JSON array of procedure stages. Each stage is a `ProcedureStage` object with `name`, `enabled`, and `executions` fields.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The schedule format. One of `Cron` or `English`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The schedule expression (cron string or English description).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the schedule is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_timezone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Timezone for the schedule (IANA TZ identifier, e.g. `America/New_York`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_alert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send an alert when the procedure runs on schedule.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_alert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send an alert when the procedure fails.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to allow triggering the procedure via webhook.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_secret": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Override the default webhook secret for this procedure.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProcedureResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProcedureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProcedureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating procedure", map[string]interface{}{"name": data.Name.ValueString()})

	cfg, d := partialProcedureConfigFromModel(&data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateProcedureRequest{
		Name:   data.Name.ValueString(),
		Config: cfg,
	}
	proc, err := r.client.CreateProcedure(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create procedure, got error: %s", err))
		return
	}
	if proc.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Procedure creation failed: missing ID",
			"The Komodo API did not return a procedure ID. Resource cannot be tracked in state.",
		)
		return
	}
	procedureToModel(proc, &data)
	tflog.Trace(ctx, "Created procedure resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProcedureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProcedureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	proc, err := r.client.GetProcedure(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read procedure, got error: %s", err))
		return
	}
	if proc == nil {
		tflog.Debug(ctx, "Procedure not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	procedureToModel(proc, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProcedureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProcedureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state ProcedureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	cfg, d := partialProcedureConfigFromModel(&data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateProcedureRequest{
		ID:     data.ID.ValueString(),
		Config: cfg,
	}
	proc, err := r.client.UpdateProcedure(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update procedure, got error: %s", err))
		return
	}
	procedureToModel(proc, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProcedureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProcedureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting procedure", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteProcedure(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete procedure, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted procedure resource")
}

func (r *ProcedureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// partialProcedureConfigFromModel converts the Terraform model into a PartialProcedureConfig.
func partialProcedureConfigFromModel(data *ProcedureResourceModel) (client.PartialProcedureConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := client.PartialProcedureConfig{}

	if !data.Stages.IsNull() && !data.Stages.IsUnknown() {
		raw := json.RawMessage(data.Stages.ValueString())
		// Validate that it's valid JSON.
		if !json.Valid(raw) {
			diags.AddAttributeError(
				path.Root("stages"),
				"Invalid JSON",
				fmt.Sprintf("The stages value is not valid JSON: %s", data.Stages.ValueString()),
			)
			return cfg, diags
		}
		cfg.Stages = raw
	}
	if !data.ScheduleFormat.IsNull() && !data.ScheduleFormat.IsUnknown() {
		v := data.ScheduleFormat.ValueString()
		cfg.ScheduleFormat = &v
	}
	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		v := data.Schedule.ValueString()
		cfg.Schedule = &v
	}
	if !data.ScheduleEnabled.IsNull() && !data.ScheduleEnabled.IsUnknown() {
		v := data.ScheduleEnabled.ValueBool()
		cfg.ScheduleEnabled = &v
	}
	if !data.ScheduleTimezone.IsNull() && !data.ScheduleTimezone.IsUnknown() {
		v := data.ScheduleTimezone.ValueString()
		cfg.ScheduleTimezone = &v
	}
	if !data.ScheduleAlert.IsNull() && !data.ScheduleAlert.IsUnknown() {
		v := data.ScheduleAlert.ValueBool()
		cfg.ScheduleAlert = &v
	}
	if !data.FailureAlert.IsNull() && !data.FailureAlert.IsUnknown() {
		v := data.FailureAlert.ValueBool()
		cfg.FailureAlert = &v
	}
	if !data.WebhookEnabled.IsNull() && !data.WebhookEnabled.IsUnknown() {
		v := data.WebhookEnabled.ValueBool()
		cfg.WebhookEnabled = &v
	}
	if !data.WebhookSecret.IsNull() && !data.WebhookSecret.IsUnknown() {
		v := data.WebhookSecret.ValueString()
		cfg.WebhookSecret = &v
	}
	return cfg, diags
}

// procedureToModel populates a ProcedureResourceModel from a Procedure API response.
func procedureToModel(proc *client.Procedure, data *ProcedureResourceModel) {
	data.ID = types.StringValue(proc.ID.OID)
	data.Name = types.StringValue(proc.Name)

	// stages: set as JSON string if non-empty, otherwise null.
	stagesStr := string(proc.Config.Stages)
	if len(proc.Config.Stages) > 0 && stagesStr != "null" {
		data.Stages = types.StringValue(stagesStr)
	} else {
		data.Stages = types.StringNull()
	}

	data.ScheduleFormat = types.StringValue(proc.Config.ScheduleFormat)
	data.Schedule = types.StringValue(proc.Config.Schedule)
	data.ScheduleEnabled = types.BoolValue(proc.Config.ScheduleEnabled)
	data.ScheduleTimezone = types.StringValue(proc.Config.ScheduleTimezone)
	data.ScheduleAlert = types.BoolValue(proc.Config.ScheduleAlert)
	data.FailureAlert = types.BoolValue(proc.Config.FailureAlert)
	data.WebhookEnabled = types.BoolValue(proc.Config.WebhookEnabled)
	data.WebhookSecret = types.StringValue(proc.Config.WebhookSecret)
}
