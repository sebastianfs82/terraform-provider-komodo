// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// ProcedureExecutionModel is the Terraform model for a single execution within a stage.
type ProcedureExecutionModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	Type       types.String `tfsdk:"type"`
	Parameters types.Map    `tfsdk:"parameters"`
}

// ProcedureStageModel is the Terraform model for one procedure stage.
type ProcedureStageModel struct {
	Name       types.String              `tfsdk:"name"`
	Executions []ProcedureExecutionModel `tfsdk:"execution"`
}

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
	ID           types.String          `tfsdk:"id"`
	Name         types.String          `tfsdk:"name"`
	Tags         types.List            `tfsdk:"tags"`
	Stages       []ProcedureStageModel `tfsdk:"stage"`
	Schedule     *ScheduleModel        `tfsdk:"schedule"`
	FailureAlert types.Bool            `tfsdk:"failure_alert_enabled"`
	Webhook      *WebhookModel         `tfsdk:"webhook"`
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
				MarkdownDescription: "The unique name of the procedure.",
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
			"schedule": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Schedule configuration for the procedure.",
				Attributes: map[string]schema.Attribute{
					"format": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The schedule format. One of `Cron` or `English`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"expression": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The schedule expression (cron string or English description).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether the schedule is enabled.",
						Default:             booldefault.StaticBool(true),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"timezone": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Timezone for the schedule (IANA TZ identifier, e.g. `America/New_York`). Defaults to `\"\"` (Core local timezone).",
						Default:             stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"alert_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether to send an alert when the procedure runs on schedule.",
						Default:             booldefault.StaticBool(true),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"failure_alert_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send an alert when the procedure fails.",
				Default:             booldefault.StaticBool(true),
			},
			"webhook": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Webhook configuration for the procedure.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to allow triggering the procedure via webhook.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Override the default webhook secret for this procedure.",
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"stage": schema.ListNestedBlock{
				MarkdownDescription: "Ordered list of procedure stages. Stages run sequentially; executions within a stage run in parallel.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The name of the stage.",
						},
					},
					Blocks: map[string]schema.Block{
						"execution": schema.ListNestedBlock{
							MarkdownDescription: "Ordered list of executions in this stage.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "The execution type (e.g. `DeployStack`, `RunBuild`, `RunProcedure`).",
									},
									"parameters": schema.MapAttribute{
										Optional:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Parameters specific to the execution type as key-value string pairs.",
									},
									"enabled": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Whether this execution is enabled.",
										Default:             booldefault.StaticBool(true),
									},
								},
							},
						},
					},
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
	plannedTags := data.Tags
	procedureToModel(proc, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Procedure", ID: proc.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on procedure, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
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
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameProcedure(ctx, client.RenameProcedureRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename procedure, got error: %s", err))
			return
		}
	}

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
	plannedTags := data.Tags
	procedureToModel(proc, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Procedure", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on procedure, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
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

	if len(data.Stages) > 0 {
		stages := make([]client.ProcedureStage, len(data.Stages))
		for i, s := range data.Stages {
			executions := make([]client.ProcedureStageExecution, len(s.Executions))
			for j, e := range s.Executions {
				var params json.RawMessage
				if e.Parameters.IsNull() || e.Parameters.IsUnknown() || len(e.Parameters.Elements()) == 0 {
					params = json.RawMessage("{}")
				} else {
					m := make(map[string]interface{}, len(e.Parameters.Elements()))
					for k, v := range e.Parameters.Elements() {
						if sv, ok := v.(types.String); ok {
							raw := sv.ValueString()
							// Attempt to decode as a native JSON number or boolean so the
							// API receives the correct type (e.g. stop_time expects i32,
							// not a string). Plain strings (IDs, names) are not valid JSON
							// tokens and will remain as strings.
							var native interface{}
							if json.Unmarshal([]byte(raw), &native) == nil {
								switch native.(type) {
								case float64, bool:
									m[k] = native
									continue
								}
							}
							m[k] = raw
						}
					}
					b, err := json.Marshal(m)
					if err != nil {
						diags.AddAttributeError(
							path.Root("stage").AtListIndex(i).AtName("execution").AtListIndex(j).AtName("parameters"),
							"Parameter Serialization Error",
							fmt.Sprintf("Unable to serialize parameters to JSON: %s", err),
						)
						return cfg, diags
					}
					params = json.RawMessage(b)
				}
				executions[j] = client.ProcedureStageExecution{
					Enabled: e.Enabled.ValueBool(),
					Execution: client.ProcedureExecution{
						Type:   e.Type.ValueString(),
						Params: params,
					},
				}
			}
			stages[i] = client.ProcedureStage{
				Name:       s.Name.ValueString(),
				Executions: executions,
			}
		}
		cfg.Stages = stages
	}
	if data.Schedule != nil {
		if !data.Schedule.Format.IsNull() && !data.Schedule.Format.IsUnknown() {
			v := data.Schedule.Format.ValueString()
			cfg.ScheduleFormat = &v
		}
		if !data.Schedule.Expression.IsNull() && !data.Schedule.Expression.IsUnknown() {
			v := data.Schedule.Expression.ValueString()
			cfg.Schedule = &v
		}
		if !data.Schedule.Enabled.IsNull() && !data.Schedule.Enabled.IsUnknown() {
			v := data.Schedule.Enabled.ValueBool()
			cfg.ScheduleEnabled = &v
		}
		if !data.Schedule.Timezone.IsNull() && !data.Schedule.Timezone.IsUnknown() {
			v := data.Schedule.Timezone.ValueString()
			cfg.ScheduleTimezone = &v
		}
		if !data.Schedule.AlertEnabled.IsNull() && !data.Schedule.AlertEnabled.IsUnknown() {
			v := data.Schedule.AlertEnabled.ValueBool()
			cfg.ScheduleAlert = &v
		}
	} else {
		f, s := false, ""
		cfg.ScheduleEnabled = &f
		cfg.Schedule = &s
		cfg.ScheduleTimezone = &s
		cfg.ScheduleAlert = &f
	}
	if !data.FailureAlert.IsNull() && !data.FailureAlert.IsUnknown() {
		v := data.FailureAlert.ValueBool()
		cfg.FailureAlert = &v
	}
	if data.Webhook != nil {
		if !data.Webhook.Enabled.IsNull() && !data.Webhook.Enabled.IsUnknown() {
			v := data.Webhook.Enabled.ValueBool()
			cfg.WebhookEnabled = &v
		}
		if !data.Webhook.Secret.IsNull() && !data.Webhook.Secret.IsUnknown() {
			v := data.Webhook.Secret.ValueString()
			cfg.WebhookSecret = &v
		}
	} else {
		f, s := false, ""
		cfg.WebhookEnabled = &f
		cfg.WebhookSecret = &s
	}
	return cfg, diags
}

// procedureToModel populates a ProcedureResourceModel from a Procedure API response.
func procedureToModel(proc *client.Procedure, data *ProcedureResourceModel) {
	data.ID = types.StringValue(proc.ID.OID)
	data.Name = types.StringValue(proc.Name)
	tagVals := make([]attr.Value, len(proc.Tags))
	for i, t := range proc.Tags {
		tagVals[i] = types.StringValue(t)
	}
	data.Tags = types.ListValueMust(types.StringType, tagVals)

	// stages: convert []client.ProcedureStage → []ProcedureStageModel
	if len(proc.Config.Stages) > 0 {
		stages := make([]ProcedureStageModel, len(proc.Config.Stages))
		for i, s := range proc.Config.Stages {
			execs := make([]ProcedureExecutionModel, len(s.Executions))
			for j, e := range s.Executions {
				// Read parameters from the API response, but only keep keys the user
				// configured. This detects out-of-band drift while ignoring extra
				// API-injected fields (e.g. services, stop_time).
				var parameters types.Map
				if i < len(data.Stages) && j < len(data.Stages[i].Executions) {
					existing := data.Stages[i].Executions[j].Parameters
					if existing.IsNull() || existing.IsUnknown() {
						parameters = existing
					} else {
						// Decode the raw API params JSON into a flat map.
						var apiParams map[string]interface{}
						if err := json.Unmarshal(e.Execution.Params, &apiParams); err != nil {
							apiParams = nil
						}
						elems := make(map[string]attr.Value, len(existing.Elements()))
						for k := range existing.Elements() {
							if apiParams != nil {
								if v, ok := apiParams[k]; ok {
									elems[k] = types.StringValue(fmt.Sprintf("%v", v))
								} else {
									// key removed on server — keep current value so user is aware
									elems[k] = existing.Elements()[k]
								}
							} else {
								elems[k] = existing.Elements()[k]
							}
						}
						parameters, _ = types.MapValue(types.StringType, elems)
					}
				} else {
					parameters = types.MapNull(types.StringType)
				}
				execs[j] = ProcedureExecutionModel{
					Enabled:    types.BoolValue(e.Enabled),
					Type:       types.StringValue(e.Execution.Type),
					Parameters: parameters,
				}
			}
			stages[i] = ProcedureStageModel{
				Name:       types.StringValue(s.Name),
				Executions: execs,
			}
		}
		data.Stages = stages
	} else {
		data.Stages = nil
	}

	if proc.Config.ScheduleEnabled || proc.Config.Schedule != "" || proc.Config.ScheduleTimezone != "" || proc.Config.ScheduleAlert {
		data.Schedule = &ScheduleModel{
			Format:       types.StringValue(proc.Config.ScheduleFormat),
			Expression:   types.StringValue(proc.Config.Schedule),
			Enabled:      types.BoolValue(proc.Config.ScheduleEnabled),
			Timezone:     types.StringValue(proc.Config.ScheduleTimezone),
			AlertEnabled: types.BoolValue(proc.Config.ScheduleAlert),
		}
	} else {
		data.Schedule = nil
	}
	data.FailureAlert = types.BoolValue(proc.Config.FailureAlert)
	webhookSecret := types.StringNull()
	if proc.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(proc.Config.WebhookSecret)
	}
	if proc.Config.WebhookEnabled || proc.Config.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(proc.Config.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
}
