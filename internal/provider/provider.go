package provider

import (
	"context"
	"os"
	"strings"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultEndpoint = "https://api.polylane.com/v1"

type polylaneProvider struct {
	version string
}

type providerModel struct {
	APIKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
}

var _ provider.Provider = &polylaneProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &polylaneProvider{version: version}
	}
}

func (p *polylaneProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "polylane"
	resp.Version = p.version
}

func (p *polylaneProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage an existing Polylane workspace, members, teams, and workspace-level settings.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Polylane workspace API key. May also be set with the POLYLANE_API_KEY environment variable.",
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Polylane API base endpoint, including the API version path. Defaults to https://api.polylane.com/v1. May also be set with the POLYLANE_ENDPOINT environment variable.",
			},
		},
	}
}

func (p *polylaneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Polylane API Key",
			"The provider cannot create an API client because api_key is unknown. Use a known value or the POLYLANE_API_KEY environment variable.",
		)
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("POLYLANE_API_KEY"))
	if !config.APIKey.IsNull() {
		apiKey = strings.TrimSpace(config.APIKey.ValueString())
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Polylane API Key",
			"Set api_key in the provider configuration or set the POLYLANE_API_KEY environment variable.",
		)
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Polylane API Endpoint",
			"The provider cannot create an API client because endpoint is unknown. Use a known value or the POLYLANE_ENDPOINT environment variable.",
		)
		return
	}

	endpoint := strings.TrimSpace(os.Getenv("POLYLANE_ENDPOINT"))
	if !config.Endpoint.IsNull() {
		endpoint = strings.TrimSpace(config.Endpoint.ValueString())
	}
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	apiClient, err := client.New(apiKey, endpoint, p.version, nil)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Invalid Polylane API Endpoint",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *polylaneProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkspaceResource,
		NewWorkspaceMemberResource,
		NewTeamResource,
		NewTeamMemberResource,
		NewWorkspaceAutofixSettingsResource,
		NewWorkspaceDigestSettingsResource,
		NewWorkspaceInvestigationsSettingsResource,
		NewWorkspaceInvestigationLimitsSettingsResource,
		NewWorkspaceModelTrainingSettingsResource,
		NewWorkspaceObservabilitySettingsResource,
		NewWorkspacePRReviewSettingsResource,
	}
}

func (p *polylaneProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
