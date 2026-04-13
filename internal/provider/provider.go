// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// Ensure KomodoProvider satisfies various provider interfaces.
var _ provider.Provider = &KomodoProvider{}
var _ provider.ProviderWithFunctions = &KomodoProvider{}
var _ provider.ProviderWithEphemeralResources = &KomodoProvider{}
var _ provider.ProviderWithActions = &KomodoProvider{}

// KomodoProvider defines the provider implementation.
type KomodoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// KomodoProviderModel describes the provider data model.
type KomodoProviderModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
}

func (p *KomodoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "komodo"
	resp.Version = p.version
}

func (p *KomodoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing Komodo resources via the Komodo API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The Komodo API endpoint URL. Can also be set via the KOMODO_ENDPOINT environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The Komodo username. Can also be set via the KOMODO_USERNAME environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The Komodo password. Can also be set via the KOMODO_PASSWORD environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Komodo API key. Can also be set via the KOMODO_API_KEY environment variable. Use with api_secret as an alternative to username/password.",
				Optional:            true,
			},
			"api_secret": schema.StringAttribute{
				MarkdownDescription: "The Komodo API secret. Can also be set via the KOMODO_API_SECRET environment variable. Use with api_key as an alternative to username/password.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *KomodoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KomodoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check for environment variables if not set in config
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("KOMODO_ENDPOINT")
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("KOMODO_API_KEY")
	}

	apiSecret := data.APISecret.ValueString()
	if apiSecret == "" {
		apiSecret = os.Getenv("KOMODO_API_SECRET")
	}

	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("KOMODO_USERNAME")
	}

	password := data.Password.ValueString()
	if password == "" {
		password = os.Getenv("KOMODO_PASSWORD")
	}

	// Validate required configuration
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Komodo Endpoint",
			"The provider cannot create the Komodo API client as there is a missing or empty value for the Komodo endpoint. "+
				"Set the endpoint value in the configuration or use the KOMODO_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Require either api_key+api_secret OR username+password
	useAPIKey := apiKey != "" || apiSecret != ""
	usePassword := username != "" || password != ""

	if useAPIKey {
		if apiKey == "" {
			resp.Diagnostics.AddError(
				"Missing Komodo API Key",
				"api_secret is set but api_key is missing. Set the api_key value in the configuration or use the KOMODO_API_KEY environment variable.",
			)
		}
		if apiSecret == "" {
			resp.Diagnostics.AddError(
				"Missing Komodo API Secret",
				"api_key is set but api_secret is missing. Set the api_secret value in the configuration or use the KOMODO_API_SECRET environment variable.",
			)
		}
	} else if usePassword {
		if username == "" {
			resp.Diagnostics.AddError(
				"Missing Komodo Username",
				"The provider cannot create the Komodo API client as there is a missing or empty value for the Komodo username. "+
					"Set the username value in the configuration or use the KOMODO_USERNAME environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
		if password == "" {
			resp.Diagnostics.AddError(
				"Missing Komodo Password",
				"The provider cannot create the Komodo API client as there is a missing or empty value for the Komodo password. "+
					"Set the password value in the configuration or use the KOMODO_PASSWORD environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Komodo Credentials",
			"The provider requires either api_key+api_secret (KOMODO_API_KEY+KOMODO_API_SECRET) "+
				"or username+password (KOMODO_USERNAME+KOMODO_PASSWORD) to authenticate.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create Komodo API client
	var komodoClient *client.Client
	if useAPIKey {
		komodoClient = client.NewClientWithApiKey(endpoint, apiKey, apiSecret)
	} else {
		komodoClient = client.NewClient(endpoint, username, password)
	}
	resp.DataSourceData = komodoClient
	resp.ResourceData = komodoClient
	resp.EphemeralResourceData = komodoClient
	resp.ActionData = komodoClient

	// Enforce minimum Komodo Core server version.
	versionResp, err := komodoClient.GetVersion(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Komodo Version Check Failed",
			"Unable to retrieve the Komodo Core server version: "+err.Error()+
				". Verify the endpoint is reachable and the credentials are correct.",
		)
		return
	}
	if !versionAtLeast(versionResp.Version, 2, 0, 0) {
		resp.Diagnostics.AddError(
			"Unsupported Komodo Server Version",
			"This provider requires Komodo Core v2.0.0 or later, but the connected server is running v"+
				versionResp.Version+". Please upgrade your Komodo Core installation.",
		)
		return
	}
}

func (p *KomodoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApiKeyResource,
		NewVariableResource,
		NewTagResource,
		NewUserGroupResource,
		NewUserGroupMembershipResource,
		NewUserResource,
		NewServiceUserResource,
		NewProviderAccountResource,
		NewRegistryAccountResource,
		NewBuilderResource,
		NewAlerterResource,
		NewRepoResource,
		NewStackResource,
		NewServerResource,
		NewNetworkResource,
		NewTerminalResource,
		NewActionResource,
		NewBuildResource,
		NewDeploymentResource,
		NewProcedureResource,
		NewResourceSyncResource,
		NewOnboardingKeyResource,
	}
}

func (p *KomodoProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		NewStackDeployAction,
		NewStackStartAction,
		NewStackStopAction,
		NewStackPauseAction,
		NewStackDestroyAction,
		NewRepoBuildAction,
		NewRepoCloneAction,
		NewRepoPullAction,
		NewServerPruneBuildxAction,
		NewServerPruneContainersAction,
		NewServerPruneBuildersAction,
		NewServerPruneImagesAction,
		NewServerPruneNetworksAction,
		NewServerPruneSystemAction,
		NewServerPruneVolumesAction,
		NewStartDeploymentAction,
		NewPullDeploymentAction,
		NewRunActionAction,
		NewRunBuildAction,
		NewRunProcedureAction,
		NewRunSyncAction,
		NewTestAlerterAction,
		NewDeployDeploymentAction,
		NewStopDeploymentAction,
		NewDestroyDeploymentAction,
		NewRestartDeploymentAction,
		NewPauseDeploymentAction,
		NewUnpauseDeploymentAction,
		NewStackRestartAction,
		NewStackUnpauseAction,
		NewStackPullAction,
		NewStackDeployIfChangedAction,
		NewStackRunServiceAction,
	}
}

func (p *KomodoProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *KomodoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVariableDataSource,
		NewVariablesDataSource,
		NewTagDataSource,
		NewTagsDataSource,
		NewUserGroupDataSource,
		NewUserGroupsDataSource,
		NewUserDataSource,
		NewUsersDataSource,
		NewServiceUserDataSource,
		NewServiceUsersDataSource,
		NewProviderAccountDataSource,
		NewProviderAccountsDataSource,
		NewRegistryAccountDataSource,
		NewRegistryAccountsDataSource,
		NewBuilderDataSource,
		NewBuildersDataSource,
		NewAlerterDataSource,
		NewAlertersDataSource,
		NewRepoDataSource,
		NewReposDataSource,
		NewStackDataSource,
		NewStacksDataSource,
		NewServerDataSource,
		NewServersDataSource,
		NewNetworkDataSource,
		NewNetworksDataSource,
		NewTerminalDataSource,
		NewTerminalsDataSource,
		NewActionDataSource,
		NewActionsDataSource,
		NewBuildDataSource,
		NewBuildsDataSource,
		NewDeploymentDataSource,
		NewDeploymentsDataSource,
		NewProcedureDataSource,
		NewProceduresDataSource,
		NewResourceSyncDataSource,
		NewResourceSyncsDataSource,
		NewOnboardingKeyDataSource,
		NewApiKeyDataSource,
		NewVersionDataSource,
	}
}

func (p *KomodoProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KomodoProvider{
			version: version,
		}
	}
}
