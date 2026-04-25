// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var _ resource.Resource = &ActionResource{}
var _ resource.ResourceWithImportState = &ActionResource{}
var _ resource.ResourceWithConfigValidators = &ActionResource{}

func NewActionResource() resource.Resource {
	return &ActionResource{}
}

type ActionResource struct {
	client *client.Client
}

// ArgumentModel holds a single key-value argument for an action.
type ArgumentModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// ScheduleModel holds the nested schedule block attributes shared by Action and Procedure resources.
type ScheduleModel struct {
	Format       types.String `tfsdk:"format"`
	Expression   types.String `tfsdk:"expression"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Timezone     types.String `tfsdk:"timezone"`
	AlertEnabled types.Bool   `tfsdk:"alert_enabled"`
}

type ActionResourceModel struct {
	ID                        types.String       `tfsdk:"id"`
	Name                      types.String       `tfsdk:"name"`
	Tags                      types.List         `tfsdk:"tags"`
	RunOnStartupEnabled       types.Bool         `tfsdk:"run_on_startup_enabled"`
	Schedule                  *ScheduleModel     `tfsdk:"schedule"`
	FailureAlert              types.Bool         `tfsdk:"failure_alert_enabled"`
	Webhook                   *WebhookModel      `tfsdk:"webhook"`
	ReloadDependenciesEnabled types.Bool         `tfsdk:"reload_dependencies_enabled"`
	FileContents              TrimmedStringValue `tfsdk:"file_contents"`
	Arguments                 []ArgumentModel    `tfsdk:"argument"`
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
				MarkdownDescription: "The unique name of the action.",
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
			"run_on_startup_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to run the action at Komodo startup.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_alert_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send an alert when the action fails.",
				Default:             booldefault.StaticBool(true),
			},
			"reload_dependencies_enabled": schema.BoolAttribute{
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
				CustomType:          TrimmedStringType{},
				MarkdownDescription: "TypeScript file contents using the Komodo client.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"schedule": schema.SingleNestedBlock{
				MarkdownDescription: "Schedule configuration for the action.",
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
						MarkdownDescription: "Whether to send an alert when the action runs on schedule.",
						Default:             booldefault.StaticBool(true),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"webhook": schema.SingleNestedBlock{
				MarkdownDescription: "Webhook configuration for the action.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to allow triggering the action via webhook.",
					},
					"secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Override the default webhook secret for this action.",
					},
				},
			},
			"argument": schema.ListNestedBlock{
				MarkdownDescription: "Key-value arguments passed to the action as the `ARGS` variable.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The argument name (environment variable name).",
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The argument value.",
						},
					},
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
	plannedTags := data.Tags
	actionToModel(a, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Action", ID: a.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on action, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
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
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameAction(ctx, client.RenameActionRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename action, got error: %s", err))
			return
		}
	}

	updateReq := client.UpdateActionRequest{
		ID:     data.ID.ValueString(),
		Config: partialActionConfigFromModel(&data),
	}
	a, err := r.client.UpdateAction(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update action, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	actionToModel(a, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Action", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on action, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
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

// ConfigValidators returns validators that run against the whole resource config.
func (r *ActionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		cronScheduleExpressionValidator{},
	}
}

// cronScheduleExpressionValidator rejects Cron schedule expressions that do not
// have 6 or 7 whitespace-separated fields. Komodo requires seconds as the first
// field, so the standard 5-field crontab syntax ("0 * * * *") is invalid.
type cronScheduleExpressionValidator struct{}

func (v cronScheduleExpressionValidator) Description(_ context.Context) string {
	return "When `schedule.format` is `Cron`, `schedule.expression` must have 6 or 7 fields (seconds required), e.g. `\"0 0 * * * *\"`."
}

func (v cronScheduleExpressionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cronScheduleExpressionValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ActionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Schedule == nil {
		return
	}
	if data.Schedule.Format.IsNull() || data.Schedule.Format.IsUnknown() {
		return
	}
	if data.Schedule.Format.ValueString() != "Cron" {
		return
	}
	if data.Schedule.Expression.IsNull() || data.Schedule.Expression.IsUnknown() {
		return
	}
	expr := strings.TrimSpace(data.Schedule.Expression.ValueString())
	fields := strings.Fields(expr)
	if n := len(fields); n != 6 && n != 7 {
		resp.Diagnostics.AddAttributeError(
			path.Root("schedule").AtName("expression"),
			"Invalid Cron expression",
			fmt.Sprintf(
				"Komodo requires cron expressions with 6 or 7 fields (seconds are mandatory as the first field). "+
					"Got %d field(s): %q. Example of a valid expression: \"0 0 * * * *\" (every day at midnight).",
				n, expr,
			),
		)
	}
}

// partialActionConfigFromModel converts the Terraform model into a PartialActionConfig.
func partialActionConfigFromModel(data *ActionResourceModel) client.PartialActionConfig {
	cfg := client.PartialActionConfig{}
	if !data.RunOnStartupEnabled.IsNull() && !data.RunOnStartupEnabled.IsUnknown() {
		v := data.RunOnStartupEnabled.ValueBool()
		cfg.RunAtStartup = &v
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
	if !data.ReloadDependenciesEnabled.IsNull() && !data.ReloadDependenciesEnabled.IsUnknown() {
		v := data.ReloadDependenciesEnabled.ValueBool()
		cfg.ReloadDenoDeps = &v
	}
	if !data.FileContents.IsNull() && !data.FileContents.IsUnknown() {
		v := data.FileContents.ValueString()
		cfg.FileContents = &v
	}
	format := "json"
	cfg.ArgumentsFormat = &format
	if len(data.Arguments) > 0 {
		m := make(map[string]string, len(data.Arguments))
		for _, arg := range data.Arguments {
			m[arg.Name.ValueString()] = arg.Value.ValueString()
		}
		b, _ := json.Marshal(m)
		v := string(b)
		cfg.Arguments = &v
	} else {
		v := "{}"
		cfg.Arguments = &v
	}
	return cfg
}

// actionToModel populates a ActionResourceModel from an Action API response.
func actionToModel(a *client.Action, data *ActionResourceModel) {
	data.ID = types.StringValue(a.ID.OID)
	data.Name = types.StringValue(a.Name)
	tagVals := make([]attr.Value, len(a.Tags))
	for i, t := range a.Tags {
		tagVals[i] = types.StringValue(t)
	}
	data.Tags = types.ListValueMust(types.StringType, tagVals)
	data.RunOnStartupEnabled = types.BoolValue(a.Config.RunAtStartup)
	if a.Config.ScheduleEnabled || a.Config.Schedule != "" || a.Config.ScheduleTimezone != "" || a.Config.ScheduleAlert {
		data.Schedule = &ScheduleModel{
			Format:       types.StringValue(a.Config.ScheduleFormat),
			Expression:   types.StringValue(a.Config.Schedule),
			Enabled:      types.BoolValue(a.Config.ScheduleEnabled),
			Timezone:     types.StringValue(a.Config.ScheduleTimezone),
			AlertEnabled: types.BoolValue(a.Config.ScheduleAlert),
		}
	} else {
		data.Schedule = nil
	}
	data.FailureAlert = types.BoolValue(a.Config.FailureAlert)
	webhookSecret := types.StringNull()
	if a.Config.WebhookSecret != "" {
		webhookSecret = types.StringValue(a.Config.WebhookSecret)
	}
	if a.Config.WebhookEnabled || a.Config.WebhookSecret != "" {
		data.Webhook = &WebhookModel{
			Enabled: types.BoolValue(a.Config.WebhookEnabled),
			Secret:  webhookSecret,
		}
	} else {
		data.Webhook = nil
	}
	data.ReloadDependenciesEnabled = types.BoolValue(a.Config.ReloadDenoDeps)
	data.FileContents = NewTrimmedStringValue(strings.TrimRight(a.Config.FileContents, "\n"))
	data.Arguments = parseActionArguments(a.Config.ArgumentsFormat, a.Config.Arguments)
}

// parseActionArguments converts a serialised arguments string back into a slice of ArgumentModel.
func parseActionArguments(format, raw string) []ArgumentModel {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	switch format {
	case "json", "":
		var m map[string]string
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil
		}
		result := make([]ArgumentModel, 0, len(m))
		for k, v := range m {
			result = append(result, ArgumentModel{
				Name:  types.StringValue(k),
				Value: types.StringValue(v),
			})
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name.ValueString() < result[j].Name.ValueString()
		})
		return result
	default:
		// key_value / toml / yaml: fall back to line-oriented KEY=VALUE parsing
		var result []ArgumentModel
		for _, line := range strings.Split(raw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			idx := strings.Index(line, "=")
			if idx < 0 {
				continue
			}
			result = append(result, ArgumentModel{
				Name:  types.StringValue(strings.TrimSpace(line[:idx])),
				Value: types.StringValue(strings.Trim(strings.TrimSpace(line[idx+1:]), `"`)),
			})
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name.ValueString() < result[j].Name.ValueString()
		})
		return result
	}
}
