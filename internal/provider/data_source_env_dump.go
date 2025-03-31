// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TerrapwnerEnvDumpDataSource{}

func NewTerrapwnerEnvDumpDataSource() datasource.DataSource {
	return &TerrapwnerEnvDumpDataSource{}
}

// TerrapwnerEnvDumpDataSource defines the data source implementation.
type TerrapwnerEnvDumpDataSource struct{}

// TerrapwnerEnvDumpDataSourceModel describes the data source data model.
type TerrapwnerEnvDumpDataSourceModel struct {
	Vars       types.Map    `tfsdk:"vars"`
	Id         types.String `tfsdk:"id"`
	MaskValues types.Bool   `tfsdk:"mask_values"`
}

func (d *TerrapwnerEnvDumpDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_dump"
}

func (d *TerrapwnerEnvDumpDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all environment variables and makes them available as a map",

		Attributes: map[string]schema.Attribute{
			"vars": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Map of all environment variables",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "Identifier for this data source",
				Computed:    true,
			},
			"mask_values": schema.BoolAttribute{
				Description: "If true, all environment variable values are replaced with '<REDACTED>' (default: true)",
				Optional:    true,
			},
		},
	}
}

func (d *TerrapwnerEnvDumpDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// No configuration needed
}

func (d *TerrapwnerEnvDumpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerEnvDumpDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default value for mask_values if not set
	if data.MaskValues.IsNull() {
		data.MaskValues = types.BoolValue(true)
	}

	// Read all environment variables
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		envVars[key] = value
	}

	// If mask_values is true, mask the values
	if data.MaskValues.ValueBool() {
		for k := range envVars {
			envVars[k] = "<REDACTED>"
		}
	}

	// Convert the map to types.Map
	envVarsMap, diags := types.MapValueFrom(ctx, types.StringType, envVars)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the environment variables in the model
	data.Vars = envVarsMap
	data.Id = types.StringValue("env_dump")

	// Write logs using the tflog package
	tflog.Trace(ctx, "read environment variables data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
