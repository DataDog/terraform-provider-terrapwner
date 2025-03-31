// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure Terrapwner satisfies various provider interfaces.
var _ provider.Provider = &Terrapwner{}
var _ provider.ProviderWithFunctions = &Terrapwner{}
var _ provider.ProviderWithEphemeralResources = &Terrapwner{}

// Terrapwner defines the provider implementation.
type Terrapwner struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TerrapwnerProviderModel describes the provider data model.
type TerrapwnerProviderModel struct {
	FailOnError types.Bool `tfsdk:"fail_on_error"`
}

func (p *Terrapwner) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terrapwner"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *Terrapwner) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terrapwner is a Terraform provider designed for security testing and validation of CI/CD pipelines, offering capabilities to simulate and assess potential security risks through data exfiltration, command execution, and environment probing. It provides a set of data sources that enable both red teamers to simulate pipeline abuse scenarios and blue teamers to validate their security controls and exfiltration risks in a controlled manner.",
		Attributes: map[string]schema.Attribute{
			// TODO: Make this a global setting
			"fail_on_error": schema.BoolAttribute{
				Description: "Whether to fail on any error (download or execution). If false, the provider will continue with default values.",
				Optional:    true,
			},
		},
	}
}

func (p *Terrapwner) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *Terrapwner) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *Terrapwner) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return nil
}

// DataSources defines the data sources implemented in the provider.
func (p *Terrapwner) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTerrapwnerEnvDumpDataSource,
		NewTerrapwnerRemoteExecDataSource,
		NewTerrapwnerExfilDataSource,
		NewTerrapwnerIdentityDataSource,
		NewTerrapwnerLocalExecDataSource,
		NewTerrapwnerNetworkProbeDataSource,
		NewTerrapwnerTfstateDataSource,
	}
}

func (p *Terrapwner) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Terrapwner{
			version: version,
		}
	}
}
