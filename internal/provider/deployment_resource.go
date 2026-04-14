// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}
var _ resource.ResourceWithConfigValidators = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client *client.Client
}

// DeploymentImageModel is the Terraform model for the image block.
type DeploymentImageModel struct {
	Type            types.String `tfsdk:"type"`
	Image           types.String `tfsdk:"name"`
	BuildID         types.String `tfsdk:"build_id"`
	Version         types.String `tfsdk:"version"`
	RegistryAccount types.String `tfsdk:"account_id"`
	RedeployEnabled types.Bool   `tfsdk:"redeploy_enabled"`
}

// DeploymentContainerModel is the Terraform model for the container block.
type DeploymentContainerModel struct {
	Network        types.String `tfsdk:"network"`
	Restart        types.String `tfsdk:"restart"`
	Command        types.String `tfsdk:"command"`
	Replicas       types.Int64  `tfsdk:"replicas"`
	ExtraArguments types.List   `tfsdk:"extra_arguments"`
	Ports          types.List   `tfsdk:"ports"`
	Volumes        types.List   `tfsdk:"volumes"`
	Environment    types.Map    `tfsdk:"environment"`
	Labels         types.List   `tfsdk:"labels"`
	Links          types.List   `tfsdk:"links"`
}

// DeploymentTerminationModel is the Terraform model for the termination block.
type DeploymentTerminationModel struct {
	Signal       types.String `tfsdk:"signal"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	SignalLabels types.String `tfsdk:"signal_labels"`
}

// DeploymentResourceModel is the Terraform resource model for komodo_deployment.
type DeploymentResourceModel struct {
	ID                             types.String                `tfsdk:"id"`
	Name                           types.String                `tfsdk:"name"`
	Tags                           types.List                  `tfsdk:"tags"`
	SwarmID                        types.String                `tfsdk:"swarm_id"`
	ServerID                       types.String                `tfsdk:"server_id"`
	Image                          *DeploymentImageModel       `tfsdk:"image"`
	SkipSecretInterpolationEnabled types.Bool                  `tfsdk:"secret_interpolation_enabled"`
	PollForUpdatesEnabled          types.Bool                  `tfsdk:"poll_updates_enabled"`
	AutoUpdateEnabled              types.Bool                  `tfsdk:"auto_update_enabled"`
	SendAlertsEnabled              types.Bool                  `tfsdk:"alerts_enabled"`
	Container                      *DeploymentContainerModel   `tfsdk:"container"`
	Termination                    *DeploymentTerminationModel `tfsdk:"termination"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Komodo deployment resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The deployment identifier (ObjectId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique name of the deployment.",
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
			"swarm_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The Swarm to deploy on (Swarm mode). Overrides `server_id` when set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The Server to deploy on (Container mode).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_interpolation_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to interpolate secrets into deployment environment variables.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"poll_updates_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to poll for image updates.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_update_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to automatically redeploy when a newer image is found. Implicitly enables `poll_updates_enabled`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"alerts_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send ContainerStateChange alerts for this deployment.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"container": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Docker container runtime configuration.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"network":         types.StringType,
						"restart":         types.StringType,
						"command":         types.StringType,
						"replicas":        types.Int64Type,
						"extra_arguments": types.ListType{ElemType: types.StringType},
						"ports":           types.ListType{ElemType: types.StringType},
						"volumes":         types.ListType{ElemType: types.StringType},
						"environment":     types.MapType{ElemType: types.StringType},
						"labels":          types.ListType{ElemType: types.StringType},
						"links":           types.ListType{ElemType: types.StringType},
					},
					map[string]attr.Value{
						"network":         types.StringValue("host"),
						"restart":         types.StringValue("no"),
						"command":         types.StringValue(""),
						"replicas":        types.Int64Value(1),
						"extra_arguments": types.ListValueMust(types.StringType, []attr.Value{}),
						"ports":           types.ListValueMust(types.StringType, []attr.Value{}),
						"volumes":         types.ListValueMust(types.StringType, []attr.Value{}),
						"environment":     types.MapValueMust(types.StringType, map[string]attr.Value{}),
						"labels":          types.ListValueMust(types.StringType, []attr.Value{}),
						"links":           types.ListValueMust(types.StringType, []attr.Value{}),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"network": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("host"),
						MarkdownDescription: "Network attached to the container. Defaults to `host`.",
					},
					"restart": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("no"),
						MarkdownDescription: "Restart mode for the container (e.g. `no`, `always`, `unless-stopped`, `on-failure`).",
					},
					"command": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Command appended to `docker run`. Passed to the container process or replaces CMD.",
					},
					"replicas": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(1),
						MarkdownDescription: "Number of replicas for the Service. Only used in Swarm mode.",
					},
					"extra_arguments": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Extra arguments interpolated into the `docker run` / `docker service create` command.",
					},
					"ports": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Container port mappings as a list of strings (e.g. `80:80`, `127.0.0.1:8080:8080`).",
					},
					"volumes": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Container volume mappings as a list of strings (e.g. `/host/path:/container/path`).",
					},
					"environment": schema.MapAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
						MarkdownDescription: "Environment variables for the container as a map of `KEY=value` pairs. Keys are automatically uppercased.",
					},
					"labels": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Docker labels for the container as a list of `key=value` strings.",
					}, "links": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
						MarkdownDescription: "Quick links displayed in the resource header.",
					}},
			},
			"termination": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Container termination behaviour.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"signal":        types.StringType,
						"timeout":       types.Int64Type,
						"signal_labels": types.StringType,
					},
					map[string]attr.Value{
						"signal":        types.StringValue("SIGTERM"),
						"timeout":       types.Int64Value(10),
						"signal_labels": types.StringValue(""),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"signal": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("SIGTERM"),
						MarkdownDescription: "Default termination signal (e.g. `SIGTERM`, `SIGKILL`). Defaults to `SIGTERM`.",
					},
					"timeout": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(10),
						MarkdownDescription: "Termination timeout in seconds.",
					},
					"signal_labels": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Labels for termination signal options (JSON/TOML string).",
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"image": schema.SingleNestedBlock{
				MarkdownDescription: "The image source for this deployment.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Image type: `Image` for an external Docker image, `Build` for a Komodo Build.",
						Validators: []validator.String{
							stringvalidator.OneOf("Image", "Build"),
						},
					},
					"name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Docker image to deploy. Used when `type` is `Image`.",
					},
					"build_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "ID of the Komodo Build to deploy. Used when `type` is `Build`.",
					},
					"version": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Build version to deploy, e.g. `1.0.0`. Used when `type` is `Build`. Defaults to latest (0.0.0).",
					},
					"account_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Account used to pull the image. Empty string uses the build/image default.",
					},
					"redeploy_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Whether to redeploy whenever the attached Build finishes.",
					},
				},
			},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating deployment", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := client.CreateDeploymentRequest{
		Name:   data.Name.ValueString(),
		Config: partialDeploymentConfigFromModel(ctx, r.client, &data),
	}
	d, err := r.client.CreateDeployment(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create deployment, got error: %s", err))
		return
	}
	if d.ID.OID == "" {
		resp.Diagnostics.AddError(
			"Deployment creation failed: missing ID",
			"The Komodo API did not return a deployment ID. Resource cannot be tracked in state.",
		)
		return
	}
	plannedTags := data.Tags
	deploymentToModel(ctx, r.client, d, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Deployment", ID: d.ID.OID},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on deployment, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	tflog.Trace(ctx, "Created deployment resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d, err := r.client.GetDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read deployment, got error: %s", err))
		return
	}
	if d == nil {
		tflog.Debug(ctx, "Deployment not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}
	deploymentToModel(ctx, r.client, d, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID
	if data.Name.ValueString() != state.Name.ValueString() {
		if err := r.client.RenameDeployment(ctx, client.RenameDeploymentRequest{
			ID:   state.ID.ValueString(),
			Name: data.Name.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename deployment, got error: %s", err))
			return
		}
	}

	updateReq := client.UpdateDeploymentRequest{
		ID:     data.ID.ValueString(),
		Config: partialDeploymentConfigFromModel(ctx, r.client, &data),
	}
	d, err := r.client.UpdateDeployment(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update deployment, got error: %s", err))
		return
	}
	plannedTags := data.Tags
	deploymentToModel(ctx, r.client, d, &data)
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plannedTags.ElementsAs(ctx, &tags, false)...)
		if !resp.Diagnostics.HasError() {
			if err := r.client.UpdateResourceMeta(ctx, client.UpdateResourceMetaRequest{
				Target: client.ResourceTarget{Type: "Deployment", ID: data.ID.ValueString()},
				Tags:   &tags,
			}); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set tags on deployment, got error: %s", err))
				return
			}
			data.Tags = plannedTags
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting deployment", map[string]interface{}{"id": data.ID.ValueString()})
	err := r.client.DeleteDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete deployment, got error: %s", err))
		return
	}
	tflog.Trace(ctx, "Deleted deployment resource")
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DeploymentResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		deploymentImageValidator{},
		deploymentAutoUpdateValidator{},
	}
}

type deploymentImageValidator struct{}

func (v deploymentImageValidator) Description(_ context.Context) string {
	return "`image.name` is required when `image.type` is `Image`"
}

func (v deploymentImageValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v deploymentImageValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Image == nil {
		return
	}
	if data.Image.Type.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("type"),
			"image.type is required",
			"`image.type` must be set when the `image` block is present.",
		)
		return
	}
	if data.Image.Type.IsUnknown() {
		return
	}
	if data.Image.Type.ValueString() == "Image" && !data.Image.Image.IsUnknown() && (data.Image.Image.IsNull() || data.Image.Image.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("name"),
			"image.name is required",
			"`image.name` must be set when `image.type` is `Image`.",
		)
	}
	if data.Image.Type.ValueString() == "Image" && !data.Image.Version.IsNull() && !data.Image.Version.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("version"),
			"image.version not allowed",
			"`image.version` can only be set when `image.type` is `Build`.",
		)
	}
	if data.Image.Type.ValueString() == "Image" && !data.Image.RedeployEnabled.IsNull() && !data.Image.RedeployEnabled.IsUnknown() && data.Image.RedeployEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("redeploy_enabled"),
			"image.redeploy_enabled not allowed",
			"`image.redeploy_enabled` can only be set when `image.type` is `Build`.",
		)
	}
	if data.Image.Type.ValueString() == "Build" && !data.Image.BuildID.IsUnknown() && (data.Image.BuildID.IsNull() || data.Image.BuildID.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("build_id"),
			"image.build_id is required",
			"`image.build_id` must be set when `image.type` is `Build`.",
		)
	}
	if data.Image.Type.ValueString() == "Build" && !data.Image.Image.IsNull() && !data.Image.Image.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("name"),
			"image.name not allowed",
			"`image.name` can only be set when `image.type` is `Image`.",
		)
	}
	if data.Image.Type.ValueString() == "Image" && !data.Image.BuildID.IsNull() && !data.Image.BuildID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("image").AtName("build_id"),
			"image.build_id not allowed",
			"`image.build_id` can only be set when `image.type` is `Build`.",
		)
	}
}

type deploymentAutoUpdateValidator struct{}

func (v deploymentAutoUpdateValidator) Description(_ context.Context) string {
	return "`poll_updates_enabled` must be true when `auto_update_enabled` is true"
}

func (v deploymentAutoUpdateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v deploymentAutoUpdateValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AutoUpdateEnabled.IsUnknown() || data.PollForUpdatesEnabled.IsUnknown() {
		return
	}
	if data.AutoUpdateEnabled.ValueBool() && !data.PollForUpdatesEnabled.IsNull() && !data.PollForUpdatesEnabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("poll_updates_enabled"),
			"poll_updates_enabled required",
			"`poll_updates_enabled` must be `true` when `auto_update_enabled` is `true`.",
		)
	}
}

// partialDeploymentConfigFromModel converts the Terraform model into a PartialDeploymentConfig.
func partialDeploymentConfigFromModel(ctx context.Context, c *client.Client, data *DeploymentResourceModel) client.PartialDeploymentConfig {
	cfg := client.PartialDeploymentConfig{}

	if !data.SwarmID.IsNull() && !data.SwarmID.IsUnknown() {
		v := data.SwarmID.ValueString()
		cfg.SwarmID = &v
	}
	if !data.ServerID.IsNull() && !data.ServerID.IsUnknown() {
		v := data.ServerID.ValueString()
		cfg.ServerID = &v
	}
	if data.Image != nil {
		img := imageModelToClient(data.Image)
		cfg.Image = &img
		// Always send account — empty string clears a previously-set value.
		username := ""
		if !data.Image.RegistryAccount.IsNull() && !data.Image.RegistryAccount.IsUnknown() {
			accountID := data.Image.RegistryAccount.ValueString()
			if accountID != "" {
				if acc, err := c.GetDockerRegistryAccount(ctx, accountID); err == nil && acc != nil {
					username = acc.Username
				} else {
					username = accountID
				}
			}
		}
		cfg.ImageRegistryAccount = &username
		if data.Image != nil && !data.Image.RedeployEnabled.IsNull() && !data.Image.RedeployEnabled.IsUnknown() {
			v := data.Image.RedeployEnabled.ValueBool()
			cfg.RedeployOnBuild = &v
		}
	}
	if !data.SkipSecretInterpolationEnabled.IsNull() && !data.SkipSecretInterpolationEnabled.IsUnknown() {
		v := !data.SkipSecretInterpolationEnabled.ValueBool()
		cfg.SkipSecretInterpolation = &v
	}
	if !data.PollForUpdatesEnabled.IsNull() && !data.PollForUpdatesEnabled.IsUnknown() {
		v := data.PollForUpdatesEnabled.ValueBool()
		cfg.PollForUpdates = &v
	}
	if !data.AutoUpdateEnabled.IsNull() && !data.AutoUpdateEnabled.IsUnknown() {
		v := data.AutoUpdateEnabled.ValueBool()
		cfg.AutoUpdate = &v
	}
	if !data.SendAlertsEnabled.IsNull() && !data.SendAlertsEnabled.IsUnknown() {
		v := data.SendAlertsEnabled.ValueBool()
		cfg.SendAlerts = &v
	}
	if data.Container != nil {
		c := data.Container
		if !c.Network.IsNull() && !c.Network.IsUnknown() {
			v := c.Network.ValueString()
			cfg.Network = &v
		}
		if !c.Restart.IsNull() && !c.Restart.IsUnknown() {
			v := c.Restart.ValueString()
			cfg.Restart = &v
		}
		if !c.Command.IsNull() && !c.Command.IsUnknown() {
			v := c.Command.ValueString()
			cfg.Command = &v
		}
		if !c.Replicas.IsNull() && !c.Replicas.IsUnknown() {
			v := int(c.Replicas.ValueInt64())
			cfg.Replicas = &v
		}
		if !c.ExtraArguments.IsNull() && !c.ExtraArguments.IsUnknown() {
			var args []string
			c.ExtraArguments.ElementsAs(ctx, &args, false)
			if args == nil {
				args = []string{}
			}
			cfg.ExtraArguments = &args
		}
		if !c.Ports.IsNull() && !c.Ports.IsUnknown() {
			var items []string
			c.Ports.ElementsAs(ctx, &items, false)
			if items == nil {
				items = []string{}
			}
			v := strings.Join(items, "\n")
			cfg.Ports = &v
		}
		if !c.Volumes.IsNull() && !c.Volumes.IsUnknown() {
			var items []string
			c.Volumes.ElementsAs(ctx, &items, false)
			if items == nil {
				items = []string{}
			}
			v := strings.Join(items, "\n")
			cfg.Volumes = &v
		}
		if !c.Environment.IsNull() && !c.Environment.IsUnknown() {
			v := envMapToString(ctx, c.Environment)
			cfg.Environment = &v
		}
		if !c.Labels.IsNull() && !c.Labels.IsUnknown() {
			var labelItems []string
			c.Labels.ElementsAs(ctx, &labelItems, false)
			if labelItems == nil {
				labelItems = []string{}
			}
			v := strings.Join(labelItems, "\n")
			cfg.Labels = &v
		}
		if !c.Links.IsNull() && !c.Links.IsUnknown() {
			var links []string
			c.Links.ElementsAs(ctx, &links, false)
			if links == nil {
				links = []string{}
			}
			cfg.Links = &links
		}
	}
	if data.Termination != nil {
		t := data.Termination
		if !t.Signal.IsNull() && !t.Signal.IsUnknown() {
			v := t.Signal.ValueString()
			cfg.TerminationSignal = &v
		}
		if !t.Timeout.IsNull() && !t.Timeout.IsUnknown() {
			v := int(t.Timeout.ValueInt64())
			cfg.TerminationTimeout = &v
		}
		if !t.SignalLabels.IsNull() && !t.SignalLabels.IsUnknown() {
			v := t.SignalLabels.ValueString()
			cfg.TerminationSignalLabels = &v
		}
	}
	return cfg
}

// imageModelToClient converts a DeploymentImageModel to a client.DeploymentImage.
func imageModelToClient(m *DeploymentImageModel) client.DeploymentImage {
	img := client.DeploymentImage{}
	imageType := m.Type.ValueString()
	switch imageType {
	case "Build":
		b := &client.DeploymentImageBuild{
			BuildID: m.BuildID.ValueString(),
		}
		if v := m.Version.ValueString(); v != "" {
			parts := strings.SplitN(v, ".", 3)
			for len(parts) < 3 {
				parts = append(parts, "0")
			}
			major, _ := strconv.Atoi(parts[0])
			minor, _ := strconv.Atoi(parts[1])
			patch, _ := strconv.Atoi(parts[2])
			b.Version = client.BuildVersion{Major: major, Minor: minor, Patch: patch}
		}
		img.Build = b
	default: // "Image"
		img.Image = &client.DeploymentImageExternal{
			Image: m.Image.ValueString(),
		}
	}
	return img
}

// deploymentToModel populates the Terraform model from an API Deployment response.
func deploymentToModel(ctx context.Context, c *client.Client, d *client.Deployment, data *DeploymentResourceModel) {
	data.ID = types.StringValue(d.ID.OID)
	data.Name = types.StringValue(d.Name)
	tagsSlice := d.Tags
	if tagsSlice == nil {
		tagsSlice = []string{}
	}
	tags, _ := types.ListValueFrom(ctx, types.StringType, tagsSlice)
	data.Tags = tags
	data.SwarmID = types.StringValue(d.Config.SwarmID)
	data.ServerID = types.StringValue(d.Config.ServerID)

	// image block: populate from the API response.
	// The API stores and returns the registry account as a username; resolve it back to an ID for state.
	registryAccountID := c.ResolveDockerRegistryAccountID(ctx, "", d.Config.ImageRegistryAccount)
	if registryAccountID == "" {
		registryAccountID = d.Config.ImageRegistryAccount
	}
	var registryAccountVal types.String
	if registryAccountID == "" {
		registryAccountVal = types.StringNull()
	} else {
		registryAccountVal = types.StringValue(registryAccountID)
	}
	if d.Config.Image.Build != nil {
		var verVal types.String
		if bv := d.Config.Image.Build.Version; bv.Major != 0 || bv.Minor != 0 || bv.Patch != 0 {
			verVal = types.StringValue(fmt.Sprintf("%d.%d.%d", bv.Major, bv.Minor, bv.Patch))
		} else if !data.Image.Version.IsNull() && !data.Image.Version.IsUnknown() && data.Image.Version.ValueString() != "" {
			// keep the user-supplied version string if it rounds to 0.0.0
			verVal = data.Image.Version
		} else {
			verVal = types.StringNull()
		}
		data.Image = &DeploymentImageModel{
			Type:            types.StringValue("Build"),
			Image:           types.StringNull(),
			BuildID:         types.StringValue(d.Config.Image.Build.BuildID),
			Version:         verVal,
			RegistryAccount: registryAccountVal,
			RedeployEnabled: types.BoolValue(d.Config.RedeployOnBuild),
		}
	} else if d.Config.Image.Image != nil && (d.Config.Image.Image.Image != "" || data.Image != nil) {
		// Only set image block when the API returned a real image value, or the user
		// already had an image block configured (empty string is the default/zero value).
		imageStr := types.StringNull()
		if d.Config.Image.Image.Image != "" {
			imageStr = types.StringValue(d.Config.Image.Image.Image)
		}
		data.Image = &DeploymentImageModel{
			Type:            types.StringValue("Image"),
			Image:           imageStr,
			BuildID:         types.StringNull(),
			Version:         types.StringNull(),
			RegistryAccount: registryAccountVal,
			RedeployEnabled: types.BoolValue(d.Config.RedeployOnBuild),
		}
	} else if data.Image != nil {
		// preserve existing model if API returned neither variant (empty/default)
		data.Image.RegistryAccount = registryAccountVal
		data.Image.RedeployEnabled = types.BoolValue(d.Config.RedeployOnBuild)
	} else {
		data.Image = nil
	}

	data.SkipSecretInterpolationEnabled = types.BoolValue(!d.Config.SkipSecretInterpolation)
	data.PollForUpdatesEnabled = types.BoolValue(d.Config.PollForUpdates)
	data.AutoUpdateEnabled = types.BoolValue(d.Config.AutoUpdate)
	data.SendAlertsEnabled = types.BoolValue(d.Config.SendAlerts)

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, d.Config.ExtraArguments)
	var labelList types.List
	if rawLabels := strings.TrimRight(d.Config.Labels, "\n"); rawLabels != "" {
		labelItems := strings.Split(rawLabels, "\n")
		labelList, _ = types.ListValueFrom(ctx, types.StringType, labelItems)
	} else {
		labelList = types.ListValueMust(types.StringType, []attr.Value{})
	}
	if data.Container != nil || d.Config.Network != "" || d.Config.Restart != "" || d.Config.Command != "" ||
		d.Config.Replicas != 0 || len(d.Config.ExtraArguments) > 0 || d.Config.Ports != "" ||
		d.Config.Volumes != "" || d.Config.Environment != "" || d.Config.Labels != "" {
		data.Container = &DeploymentContainerModel{
			Network:        types.StringValue(d.Config.Network),
			Restart:        types.StringValue(d.Config.Restart),
			Command:        types.StringValue(d.Config.Command),
			Replicas:       types.Int64Value(int64(d.Config.Replicas)),
			ExtraArguments: extraArgs,
			Ports: func() types.List {
				if raw := strings.TrimRight(d.Config.Ports, "\n"); raw != "" {
					items := strings.Split(raw, "\n")
					v, _ := types.ListValueFrom(ctx, types.StringType, items)
					return v
				}
				return types.ListValueMust(types.StringType, []attr.Value{})
			}(),
			Volumes: func() types.List {
				if raw := strings.TrimRight(d.Config.Volumes, "\n"); raw != "" {
					items := strings.Split(raw, "\n")
					v, _ := types.ListValueFrom(ctx, types.StringType, items)
					return v
				}
				return types.ListValueMust(types.StringType, []attr.Value{})
			}(),
			Environment: func() types.Map {
				m := envStringToMap(strings.TrimRight(d.Config.Environment, "\n"))
				if m.IsNull() {
					return types.MapValueMust(types.StringType, map[string]attr.Value{})
				}
				return m
			}(),
			Labels: labelList,
			Links: func() types.List {
				v, _ := types.ListValueFrom(ctx, types.StringType, d.Config.Links)
				return v
			}(),
		}
	}

	if data.Termination != nil || d.Config.TerminationSignal != "" || d.Config.TerminationTimeout != 0 || d.Config.TerminationSignalLabels != "" {
		data.Termination = &DeploymentTerminationModel{
			Signal:       types.StringValue(d.Config.TerminationSignal),
			Timeout:      types.Int64Value(int64(d.Config.TerminationTimeout)),
			SignalLabels: types.StringValue(d.Config.TerminationSignalLabels),
		}
	}
}
