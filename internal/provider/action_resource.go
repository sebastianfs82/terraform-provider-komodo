// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &ActionResource{}
var _ resource.ResourceWithImportState = &ActionResource{}

func NewActionResource() resource.Resource {
	return &ActionResource{}
}

type ActionResource struct {
	client *client.Client
}

type ActionResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	RunAtStartup     types.Bool   `tfsdk:"run_at_startup"`
	ScheduleFormat   types.String `tfsdk:"schedule_format"`
	Schedule         types.String `tfsdk:"schedule"`
	ScheduleEnabled  types.Bool   `tfsdk:"schedule_enabled"`
	ScheduleTimezone types.String `tfsdk:"schedule_timezone"`
	ScheduleAlert    types.Bool   `tfsdk:"schedule_alert"`
	FailureAlert     types.Bool   `tfsdk:"failure_alert"`
	WebhookEnabled   types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret    types.String `tfsdk:"webhook_secret"`
	ReloadDenoDeps   types.Bool   `tfsdk:"reload_deno_deps"`
	FileContents     types.String `tfsdk:"file_contents"`
	ArgumentsFormat  types.String `tfsdk:"arguments_format"`
	Arguments        types.String `tfsdk:"arguments"`
}

func (r *ActionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (r *ActionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo action resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The action identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the action. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"run_at_startup": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to run the action at Komodo startup.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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
				MarkdownDescription: "Whether to send an alert when the action runs on schedule.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_alert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send an alert when the action fails.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to allow triggering the action via webhook.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_secret": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Override the default webhook secret for this action.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reload_deno_deps": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to reload Deno dependencies on each run.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"file_contents": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "TypeScript file contents using the Komodo client.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arguments_format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The format for action arguments (e.g. `KeyValue` or `Toml`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arguments": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default arguments passed to the action as the `ARGS` variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ActionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating action", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := client.CreateActionRequest{
		Name:   data.Name.ValueString(),
		Config: partialActionConfigFromModel(&data),
	}
	a, err := r.client.CreateAction(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create action, got error: %s", err))
		return
	}
	if a.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Action creation failed: missing ID",
			"The Komodo API did not return an action ID. Resource cannot be tracked in state.",
		)
		return
	}
	actionToModel(a, &data)
	tflog.Trace(ctx, "Created action resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := r.client.GetAction(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read action, got error: %s", err))
		return
	}
	if a == nil {
		tflog.Debug(ctx, "Action not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	actionToModel(a, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	updateReq := client.UpdateActionRequest{
		ID:     data.ID.ValueString(),
		Config: partialActionConfigFromModel(&data),
	}
	a, err := r.client.UpdateAction(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update action, got error: %s", err))
		return
	}
	actionToModel(a, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting action", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteAction(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete action, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted action resource")
}

func (r *ActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// partialActionConfigFromModel converts the Terraform model into a PartialActionConfig.
func partialActionConfigFromModel(data *ActionResourceModel) client.PartialActionConfig {
	cfg := client.PartialActionConfig{}
	if !data.RunAtStartup.IsNull() && !data.RunAtStartup.IsUnknown() {
		v := data.RunAtStartup.ValueBool()
		cfg.RunAtStartup = &v
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
	if !data.ReloadDenoDeps.IsNull() && !data.ReloadDenoDeps.IsUnknown() {
		v := data.ReloadDenoDeps.ValueBool()
		cfg.ReloadDenoDeps = &v
	}
	if !data.FileContents.IsNull() && !data.FileContents.IsUnknown() {
		v := data.FileContents.ValueString()
		cfg.FileContents = &v
	}
	if !data.ArgumentsFormat.IsNull() && !data.ArgumentsFormat.IsUnknown() {
		v := data.ArgumentsFormat.ValueString()
		cfg.ArgumentsFormat = &v
	}
	if !data.Arguments.IsNull() && !data.Arguments.IsUnknown() {
		v := data.Arguments.ValueString()
		cfg.Arguments = &v
	}
	return cfg
}

// actionToModel populates a ActionResourceModel from an Action API response.
func actionToModel(a *client.Action, data *ActionResourceModel) {
	data.ID = types.StringValue(a.ID.OID)
	data.Name = types.StringValue(a.Name)
	data.RunAtStartup = types.BoolValue(a.Config.RunAtStartup)
	data.ScheduleFormat = types.StringValue(a.Config.ScheduleFormat)
	data.Schedule = types.StringValue(a.Config.Schedule)
	data.ScheduleEnabled = types.BoolValue(a.Config.ScheduleEnabled)
	data.ScheduleTimezone = types.StringValue(a.Config.ScheduleTimezone)
	data.ScheduleAlert = types.BoolValue(a.Config.ScheduleAlert)
	data.FailureAlert = types.BoolValue(a.Config.FailureAlert)
	data.WebhookEnabled = types.BoolValue(a.Config.WebhookEnabled)
	data.WebhookSecret = types.StringValue(a.Config.WebhookSecret)
	data.ReloadDenoDeps = types.BoolValue(a.Config.ReloadDenoDeps)
	data.FileContents = types.StringValue(strings.TrimRight(a.Config.FileContents, "\n"))
	data.ArgumentsFormat = types.StringValue(a.Config.ArgumentsFormat)
	data.Arguments = types.StringValue(a.Config.Arguments)
}
