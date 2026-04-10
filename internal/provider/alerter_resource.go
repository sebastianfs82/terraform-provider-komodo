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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &AlerterResource{}
var _ resource.ResourceWithImportState = &AlerterResource{}

func NewAlerterResource() resource.Resource {
	return &AlerterResource{}
}

type AlerterResource struct {
	client *client.Client
}

type AlerterResourceModel struct {
	ID               types.String          `tfsdk:"id"`
	Name             types.String          `tfsdk:"name"`
	Enabled          types.Bool            `tfsdk:"enabled"`
	EndpointType     types.String          `tfsdk:"endpoint_type"`
	AlertTypes       types.List            `tfsdk:"alert_types"`
	CustomEndpoint   *AlerterCustomModel   `tfsdk:"custom_endpoint"`
	SlackEndpoint    *AlerterSlackModel    `tfsdk:"slack_endpoint"`
	DiscordEndpoint  *AlerterDiscordModel  `tfsdk:"discord_endpoint"`
	NtfyEndpoint     *AlerterNtfyModel     `tfsdk:"ntfy_endpoint"`
	PushoverEndpoint *AlerterPushoverModel `tfsdk:"pushover_endpoint"`
}

type AlerterCustomModel struct {
	URL types.String `tfsdk:"url"`
}

type AlerterSlackModel struct {
	URL types.String `tfsdk:"url"`
}

type AlerterDiscordModel struct {
	URL types.String `tfsdk:"url"`
}

type AlerterNtfyModel struct {
	URL   types.String `tfsdk:"url"`
	Email types.String `tfsdk:"email"`
}

type AlerterPushoverModel struct {
	URL types.String `tfsdk:"url"`
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
				MarkdownDescription: "The unique name of the alerter. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
			"endpoint_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The alerter endpoint type. One of `Custom`, `Slack`, `Discord`, `Ntfy`, `Pushover`.",
			},
			"alert_types": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Only send specific alert types. If empty, all alert types are sent.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a Custom HTTP endpoint alerter. Required when `endpoint_type` is `Custom`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The HTTP/S endpoint URL to send the POST to.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"slack_endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a Slack alerter. Required when `endpoint_type` is `Slack`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Slack app webhook URL.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"discord_endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a Discord alerter. Required when `endpoint_type` is `Discord`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Discord webhook URL.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"ntfy_endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a Ntfy alerter. Required when `endpoint_type` is `Ntfy`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Ntfy topic URL.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"email": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Optional email address for Ntfy email notifications. Requires SMTP configured on the Ntfy server.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"pushover_endpoint": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Configuration for a Pushover alerter. Required when `endpoint_type` is `Pushover`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Pushover URL including application and user tokens in parameters.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
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
		"name":          data.Name.ValueString(),
		"endpoint_type": data.EndpointType.ValueString(),
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
	resp.Diagnostics.Append(alerterToModel(ctx, a, &data)...)
	if resp.Diagnostics.HasError() {
		return
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
	resp.Diagnostics.Append(alerterToModel(ctx, a, &data)...)
	if resp.Diagnostics.HasError() {
		return
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

	// alert_types
	if !data.AlertTypes.IsNull() && !data.AlertTypes.IsUnknown() {
		var alertTypes []string
		diags.Append(data.AlertTypes.ElementsAs(ctx, &alertTypes, false)...)
		if diags.HasError() {
			return cfg, diags
		}
		cfg.AlertTypes = alertTypes
	}

	// endpoint
	switch data.EndpointType.ValueString() {
	case "Custom":
		params := client.CustomAlerterEndpoint{URL: "http://localhost:7000"}
		if data.CustomEndpoint != nil {
			params.URL = data.CustomEndpoint.URL.ValueString()
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Custom", Params: params}
	case "Slack":
		params := client.SlackAlerterEndpoint{}
		if data.SlackEndpoint != nil {
			params.URL = data.SlackEndpoint.URL.ValueString()
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Slack", Params: params}
	case "Discord":
		params := client.DiscordAlerterEndpoint{}
		if data.DiscordEndpoint != nil {
			params.URL = data.DiscordEndpoint.URL.ValueString()
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Discord", Params: params}
	case "Ntfy":
		params := client.NtfyAlerterEndpoint{}
		if data.NtfyEndpoint != nil {
			params.URL = data.NtfyEndpoint.URL.ValueString()
			if !data.NtfyEndpoint.Email.IsNull() && !data.NtfyEndpoint.Email.IsUnknown() {
				email := data.NtfyEndpoint.Email.ValueString()
				params.Email = &email
			}
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Ntfy", Params: params}
	case "Pushover":
		params := client.PushoverAlerterEndpoint{}
		if data.PushoverEndpoint != nil {
			params.URL = data.PushoverEndpoint.URL.ValueString()
		}
		cfg.Endpoint = &client.AlerterEndpointInput{Type: "Pushover", Params: params}
	default:
		diags.AddError("Config Error", fmt.Sprintf("Unknown endpoint_type: %q; must be one of Custom, Slack, Discord, Ntfy, Pushover", data.EndpointType.ValueString()))
	}

	return cfg, diags
}

// alerterToModel reads an Alerter API response into the Terraform resource model.
func alerterToModel(ctx context.Context, a *client.Alerter, data *AlerterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(a.ID.OID)
	data.Name = types.StringValue(a.Name)
	data.Enabled = types.BoolValue(a.Config.Enabled)
	data.EndpointType = types.StringValue(a.Config.Endpoint.Type)

	// alert_types
	alertTypes, atDiags := types.ListValueFrom(ctx, types.StringType, a.Config.AlertTypes)
	diags.Append(atDiags...)
	if diags.HasError() {
		return diags
	}
	data.AlertTypes = alertTypes

	// endpoint nested blocks
	switch a.Config.Endpoint.Type {
	case "Custom":
		p, err := a.Config.Endpoint.GetCustomParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Custom alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			data.CustomEndpoint = &AlerterCustomModel{URL: types.StringValue(p.URL)}
		}
		data.SlackEndpoint = nil
		data.DiscordEndpoint = nil
		data.NtfyEndpoint = nil
		data.PushoverEndpoint = nil
	case "Slack":
		p, err := a.Config.Endpoint.GetSlackParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Slack alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			data.SlackEndpoint = &AlerterSlackModel{URL: types.StringValue(p.URL)}
		}
		data.CustomEndpoint = nil
		data.DiscordEndpoint = nil
		data.NtfyEndpoint = nil
		data.PushoverEndpoint = nil
	case "Discord":
		p, err := a.Config.Endpoint.GetDiscordParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Discord alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			data.DiscordEndpoint = &AlerterDiscordModel{URL: types.StringValue(p.URL)}
		}
		data.CustomEndpoint = nil
		data.SlackEndpoint = nil
		data.NtfyEndpoint = nil
		data.PushoverEndpoint = nil
	case "Ntfy":
		p, err := a.Config.Endpoint.GetNtfyParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Ntfy alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			ntfy := &AlerterNtfyModel{URL: types.StringValue(p.URL)}
			if p.Email != nil {
				ntfy.Email = types.StringValue(*p.Email)
			} else {
				ntfy.Email = types.StringNull()
			}
			data.NtfyEndpoint = ntfy
		}
		data.CustomEndpoint = nil
		data.SlackEndpoint = nil
		data.DiscordEndpoint = nil
		data.PushoverEndpoint = nil
	case "Pushover":
		p, err := a.Config.Endpoint.GetPushoverParams()
		if err != nil {
			diags.AddError("Parse Error", fmt.Sprintf("Unable to decode Pushover alerter endpoint: %s", err))
			return diags
		}
		if p != nil {
			data.PushoverEndpoint = &AlerterPushoverModel{URL: types.StringValue(p.URL)}
		}
		data.CustomEndpoint = nil
		data.SlackEndpoint = nil
		data.DiscordEndpoint = nil
		data.NtfyEndpoint = nil
	default:
		diags.AddError("Unknown Endpoint Type", fmt.Sprintf("Unknown alerter endpoint type from API: %q", a.Config.Endpoint.Type))
	}

	return diags
}
