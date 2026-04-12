// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client *client.Client
}

// DeploymentImageModel is the Terraform model for the image block.
type DeploymentImageModel struct {
	Type    types.String       `tfsdk:"type"`
	Image   types.String       `tfsdk:"image"`
	BuildID types.String       `tfsdk:"build_id"`
	Version *BuildVersionModel `tfsdk:"version"`
}

// DeploymentResourceModel is the Terraform resource model for komodo_deployment.
type DeploymentResourceModel struct {
	ID                   types.String          `tfsdk:"id"`
	Name                 types.String          `tfsdk:"name"`
	SwarmID              types.String          `tfsdk:"swarm_id"`
	ServerID             types.String          `tfsdk:"server_id"`
	Image                *DeploymentImageModel `tfsdk:"image"`
	ImageRegistryAccount types.String          `tfsdk:"image_registry_account"`
	SkipSecretInterp     types.Bool            `tfsdk:"skip_secret_interp"`
	RedeployOnBuild      types.Bool            `tfsdk:"redeploy_on_build"`
	PollForUpdates       types.Bool            `tfsdk:"poll_for_updates"`
	AutoUpdate           types.Bool            `tfsdk:"auto_update"`
	SendAlerts           types.Bool            `tfsdk:"send_alerts"`
	Links                types.List            `tfsdk:"links"`
	Network              types.String          `tfsdk:"network"`
	Restart              types.String          `tfsdk:"restart"`
	Command              types.String          `tfsdk:"command"`
	Replicas             types.Int64           `tfsdk:"replicas"`
	TerminationSignal    types.String          `tfsdk:"termination_signal"`
	TerminationTimeout   types.Int64           `tfsdk:"termination_timeout"`
	ExtraArgs            types.List            `tfsdk:"extra_args"`
	TermSignalLabels     types.String          `tfsdk:"term_signal_labels"`
	Ports                types.String          `tfsdk:"ports"`
	Volumes              types.String          `tfsdk:"volumes"`
	Environment          types.String          `tfsdk:"environment"`
	Labels               types.String          `tfsdk:"labels"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	buildVersionAttrs := map[string]schema.Attribute{
		"major": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Major version component.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"minor": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Minor version component.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"patch": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Patch version component.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	}

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
			"image": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The image source for this deployment.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Image type: `Image` for an external Docker image, `Build` for a Komodo Build.",
					},
					"image": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Docker image to deploy. Used when `type` is `Image`.",
					},
					"build_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "ID of the Komodo Build to deploy. Used when `type` is `Build`.",
					},
					"version": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Build version to deploy. Used when `type` is `Build`. Defaults to latest (0.0.0).",
						Attributes:          buildVersionAttrs,
					},
				},
			},
			"image_registry_account": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Account used to pull the image. Empty string uses the build/image default.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_secret_interp": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to skip secret interpolation into deployment environment variables.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"redeploy_on_build": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to redeploy whenever the attached Build finishes.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"poll_for_updates": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to poll for image updates.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_update": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to automatically redeploy when a newer image is found. Implicitly enables `poll_for_updates`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to send ContainerStateChange alerts for this deployment.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"links": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Quick links displayed in the resource header.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"network": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network attached to the container. Defaults to `host`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"restart": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Restart mode for the container (e.g. `no`, `always`, `unless-stopped`, `on-failure`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Command appended to `docker run`. Passed to the container process or replaces CMD.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Number of replicas for the Service. Only used in Swarm mode.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"termination_signal": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default termination signal (e.g. `SigTerm`, `SigKill`). Defaults to `SigTerm`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"termination_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Termination timeout in seconds.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"extra_args": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra arguments interpolated into the `docker run` / `docker service create` command.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"term_signal_labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Labels for termination signal options (JSON/TOML string).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ports": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Container port mapping. Irrelevant when network is `host`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"volumes": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Container volume mapping (host path → container path).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Environment variables passed to the container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Docker labels for the container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
		Config: partialDeploymentConfigFromModel(ctx, &data),
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
	deploymentToModel(ctx, d, &data)
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
	deploymentToModel(ctx, d, &data)
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
		Config: partialDeploymentConfigFromModel(ctx, &data),
	}
	d, err := r.client.UpdateDeployment(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update deployment, got error: %s", err))
		return
	}
	deploymentToModel(ctx, d, &data)
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

// partialDeploymentConfigFromModel converts the Terraform model into a PartialDeploymentConfig.
func partialDeploymentConfigFromModel(ctx context.Context, data *DeploymentResourceModel) client.PartialDeploymentConfig {
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
	}
	if !data.ImageRegistryAccount.IsNull() && !data.ImageRegistryAccount.IsUnknown() {
		v := data.ImageRegistryAccount.ValueString()
		cfg.ImageRegistryAccount = &v
	}
	if !data.SkipSecretInterp.IsNull() && !data.SkipSecretInterp.IsUnknown() {
		v := data.SkipSecretInterp.ValueBool()
		cfg.SkipSecretInterp = &v
	}
	if !data.RedeployOnBuild.IsNull() && !data.RedeployOnBuild.IsUnknown() {
		v := data.RedeployOnBuild.ValueBool()
		cfg.RedeployOnBuild = &v
	}
	if !data.PollForUpdates.IsNull() && !data.PollForUpdates.IsUnknown() {
		v := data.PollForUpdates.ValueBool()
		cfg.PollForUpdates = &v
	}
	if !data.AutoUpdate.IsNull() && !data.AutoUpdate.IsUnknown() {
		v := data.AutoUpdate.ValueBool()
		cfg.AutoUpdate = &v
	}
	if !data.SendAlerts.IsNull() && !data.SendAlerts.IsUnknown() {
		v := data.SendAlerts.ValueBool()
		cfg.SendAlerts = &v
	}
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var links []string
		data.Links.ElementsAs(ctx, &links, false)
		if links == nil {
			links = []string{}
		}
		cfg.Links = &links
	}
	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		v := data.Network.ValueString()
		cfg.Network = &v
	}
	if !data.Restart.IsNull() && !data.Restart.IsUnknown() {
		v := data.Restart.ValueString()
		cfg.Restart = &v
	}
	if !data.Command.IsNull() && !data.Command.IsUnknown() {
		v := data.Command.ValueString()
		cfg.Command = &v
	}
	if !data.Replicas.IsNull() && !data.Replicas.IsUnknown() {
		v := int(data.Replicas.ValueInt64())
		cfg.Replicas = &v
	}
	if !data.TerminationSignal.IsNull() && !data.TerminationSignal.IsUnknown() {
		v := data.TerminationSignal.ValueString()
		cfg.TerminationSignal = &v
	}
	if !data.TerminationTimeout.IsNull() && !data.TerminationTimeout.IsUnknown() {
		v := int(data.TerminationTimeout.ValueInt64())
		cfg.TerminationTimeout = &v
	}
	if !data.ExtraArgs.IsNull() && !data.ExtraArgs.IsUnknown() {
		var args []string
		data.ExtraArgs.ElementsAs(ctx, &args, false)
		if args == nil {
			args = []string{}
		}
		cfg.ExtraArgs = &args
	}
	if !data.TermSignalLabels.IsNull() && !data.TermSignalLabels.IsUnknown() {
		v := data.TermSignalLabels.ValueString()
		cfg.TermSignalLabels = &v
	}
	if !data.Ports.IsNull() && !data.Ports.IsUnknown() {
		v := data.Ports.ValueString()
		cfg.Ports = &v
	}
	if !data.Volumes.IsNull() && !data.Volumes.IsUnknown() {
		v := data.Volumes.ValueString()
		cfg.Volumes = &v
	}
	if !data.Environment.IsNull() && !data.Environment.IsUnknown() {
		v := data.Environment.ValueString()
		cfg.Environment = &v
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		v := data.Labels.ValueString()
		cfg.Labels = &v
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
		if m.Version != nil {
			b.Version = client.BuildVersion{
				Major: int(m.Version.Major.ValueInt64()),
				Minor: int(m.Version.Minor.ValueInt64()),
				Patch: int(m.Version.Patch.ValueInt64()),
			}
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
func deploymentToModel(ctx context.Context, d *client.Deployment, data *DeploymentResourceModel) {
	data.ID = types.StringValue(d.ID.OID)
	data.Name = types.StringValue(d.Name)
	data.SwarmID = types.StringValue(d.Config.SwarmID)
	data.ServerID = types.StringValue(d.Config.ServerID)

	// image block: populate from the API response.
	if d.Config.Image.Build != nil {
		ver := &BuildVersionModel{
			Major: types.Int64Value(int64(d.Config.Image.Build.Version.Major)),
			Minor: types.Int64Value(int64(d.Config.Image.Build.Version.Minor)),
			Patch: types.Int64Value(int64(d.Config.Image.Build.Version.Patch)),
		}
		// preserve version block only if it was set in model or any component non-zero
		if data.Image != nil && data.Image.Version != nil {
			// keep it
		} else if d.Config.Image.Build.Version.Major != 0 || d.Config.Image.Build.Version.Minor != 0 || d.Config.Image.Build.Version.Patch != 0 {
			// keep it
		} else {
			ver = nil
		}
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Build"),
			Image:   types.StringNull(),
			BuildID: types.StringValue(d.Config.Image.Build.BuildID),
			Version: ver,
		}
	} else if d.Config.Image.Image != nil && (d.Config.Image.Image.Image != "" || data.Image != nil) {
		// Only set image block when the API returned a real image value, or the user
		// already had an image block configured (empty string is the default/zero value).
		data.Image = &DeploymentImageModel{
			Type:    types.StringValue("Image"),
			Image:   types.StringValue(d.Config.Image.Image.Image),
			BuildID: types.StringNull(),
			Version: nil,
		}
	} else if data.Image != nil {
		// preserve existing model if API returned neither variant (empty/default)
		// do nothing
	} else {
		data.Image = nil
	}

	data.ImageRegistryAccount = types.StringValue(d.Config.ImageRegistryAccount)
	data.SkipSecretInterp = types.BoolValue(d.Config.SkipSecretInterp)
	data.RedeployOnBuild = types.BoolValue(d.Config.RedeployOnBuild)
	data.PollForUpdates = types.BoolValue(d.Config.PollForUpdates)
	data.AutoUpdate = types.BoolValue(d.Config.AutoUpdate)
	data.SendAlerts = types.BoolValue(d.Config.SendAlerts)

	links, _ := types.ListValueFrom(ctx, types.StringType, d.Config.Links)
	data.Links = links

	data.Network = types.StringValue(d.Config.Network)
	data.Restart = types.StringValue(d.Config.Restart)
	data.Command = types.StringValue(d.Config.Command)
	data.Replicas = types.Int64Value(int64(d.Config.Replicas))
	data.TerminationSignal = types.StringValue(d.Config.TerminationSignal)
	data.TerminationTimeout = types.Int64Value(int64(d.Config.TerminationTimeout))

	extraArgs, _ := types.ListValueFrom(ctx, types.StringType, d.Config.ExtraArgs)
	data.ExtraArgs = extraArgs

	data.TermSignalLabels = types.StringValue(d.Config.TermSignalLabels)
	data.Ports = types.StringValue(d.Config.Ports)
	data.Volumes = types.StringValue(d.Config.Volumes)
	data.Environment = types.StringValue(d.Config.Environment)
	data.Labels = types.StringValue(d.Config.Labels)
}
