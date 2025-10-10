// Copyright (c) Igor Voronin
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/divStar/terraform-provider-zitactl/internal/provider/application_oidc"
	"github.com/divStar/terraform-provider-zitactl/internal/provider/client"
	"github.com/divStar/terraform-provider-zitactl/internal/provider/org"
	"github.com/divStar/terraform-provider-zitactl/internal/provider/project"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &ZitactlProvider{}

// ZitactlProvider is the provider implementation.
type ZitactlProvider struct {
	version       string
	clientFactory client.ClientFactory
}

// New creates a provider with the default client factory.
func New(version string) func() provider.Provider {
	return NewWithClientFactory(version, nil)
}

// NewWithClientFactory creates a provider with a custom client factory.
// This is primarily used for testing to inject mock client factories.
// If factory is nil, the default client factory will be used.
func NewWithClientFactory(version string, factory client.ClientFactory) func() provider.Provider {
	return func() provider.Provider {
		return &ZitactlProvider{
			version:       version,
			clientFactory: factory,
		}
	}
}

// Metadata sets the provider name and version.
func (p *ZitactlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zitactl"
	resp.Version = p.version
}

// Schema defines the provider-level configuration schema.
func (p *ZitactlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Zitadel instance domain (e.g., 'zitadel.example.com'). Can also be set via ZITACTL_DOMAIN environment variable.",
				Optional:            true,
			},
			"skip_tls_verification": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS verification (e.g. when using self-signed certificates). Can also be set via ZITACTL_SKIP_TLS_VERIFICATION environment variable. Note: this ensures, that invalid HTTPS certificates are tolerated.",
				Optional:            true,
			},
			"service_account_key": schema.StringAttribute{
				MarkdownDescription: "Service account key as a **__decoded__ JSON string**. Can also be set via ZITACTL_SERVICE_ACCOUNT_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// Resources returns the list of resources provided by this provider.
func (p *ZitactlProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		project.NewProjectResource,
		application_oidc.NewApplicationOIDCResource,
	}
}

// DataSources returns the list of data sources provided by this provider.
func (p *ZitactlProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		org.NewOrgsDataSource,
	}
}

// Configure prepares the Zitadel client for data sources and resources.
// It handles configuration from both explicit provider configuration and environment variables.
func (p *ZitactlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data client.ZitactlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Don't check for unknown values - just store the config
	// Client will be created later when actually needed
	clientInfo := &client.ClientInfo{
		Config:        &data,
		ClientFactory: p.clientFactory,
	}

	resp.DataSourceData = clientInfo
	resp.ResourceData = clientInfo
}
