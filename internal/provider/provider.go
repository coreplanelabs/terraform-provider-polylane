package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultEndpoint = "https://api.polylane.io/api/v1"

// providerConfig carries provider-level configuration to resources and data
// sources via ConfigureResponse.ResourceData / DataSourceData. The Polylane
// API client will live here once it exists.
type providerConfig struct {
	APIKey   string
	Endpoint string
}

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
		Description: "Terraform provider for Polylane.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Polylane API key. May also be set with POLYLANE_API_KEY.",
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Polylane API endpoint. May also be set with POLYLANE_ENDPOINT.",
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

	apiKey := os.Getenv("POLYLANE_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	endpoint := os.Getenv("POLYLANE_ENDPOINT")
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	// No resources or data sources exist yet, so an API key is not required.
	// Once the Polylane API client lands, construct it here and fail with an
	// attribute error when the key is missing.
	cfg := &providerConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
	}

	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

func (p *polylaneProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *polylaneProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
