// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/datadog/terraform-provider-terrapwner/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TerrapwnerTfstateDataSource{}
	_ datasource.DataSourceWithConfigure = &TerrapwnerTfstateDataSource{}
)

// NewTerrapwnerTfstateDataSource is a helper function to simplify the provider implementation.
func NewTerrapwnerTfstateDataSource() datasource.DataSource {
	return &TerrapwnerTfstateDataSource{}
}

// TerrapwnerTfstateDataSource is the data source implementation.
type TerrapwnerTfstateDataSource struct{}

// TerrapwnerTfstateDataSourceModel describes the data source data model.
type TerrapwnerTfstateDataSourceModel struct {
	Success          types.Bool   `tfsdk:"success"`
	RawJSON          types.String `tfsdk:"raw_json"`
	ResourceTypes    types.List   `tfsdk:"resource_types"`
	ResourceCount    types.Int64  `tfsdk:"resource_count"`
	Providers        types.List   `tfsdk:"providers"`
	Modules          types.List   `tfsdk:"modules"`
	SensitiveOutputs types.Map    `tfsdk:"sensitive_outputs"`
}

// state represents the structure of the Terraform state JSON.
type state struct {
	Values struct {
		RootModule struct {
			Resources []struct {
				Type string `json:"type"`
			} `json:"resources"`
			ChildModules []struct {
				Address string `json:"address"`
			} `json:"child_modules"`
		} `json:"root_module"`
		Outputs map[string]struct {
			Sensitive bool `json:"sensitive"`
		} `json:"outputs"`
	} `json:"values"`
}

// Configure adds the provider configured client to the data source.
func (d *TerrapwnerTfstateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// No configuration needed
}

// Metadata returns the data source type name.
func (d *TerrapwnerTfstateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tfstate"
}

// Schema defines the schema for the data source.
func (d *TerrapwnerTfstateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads and leaks the Terraform state using 'terraform show -json'.",
		Attributes: map[string]schema.Attribute{
			"success": schema.BoolAttribute{
				Description: "Whether the state was read successfully.",
				Computed:    true,
			},
			"raw_json": schema.StringAttribute{
				Description: "Raw JSON output from 'terraform show -json'.",
				Computed:    true,
			},
			"resource_types": schema.ListAttribute{
				Description: "List of unique resource types in the Terraform state.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"resource_count": schema.Int64Attribute{
				Description: "Total number of managed resources in the Terraform state.",
				Computed:    true,
			},
			"providers": schema.ListAttribute{
				Description: "List of unique provider names used in the Terraform state.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"modules": schema.ListAttribute{
				Description: "List of unique module names used in the Terraform state.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"sensitive_outputs": schema.MapAttribute{
				Description: "Map of output names to true for all outputs marked as sensitive.",
				ElementType: types.BoolType,
				Computed:    true,
			},
		},
	}
}

// extractResourceInfo extracts resource types and provider names from the state.
func extractResourceInfo(resources []struct {
	Type string `json:"type"`
}) (resourceTypes, providers map[string]struct{}) {
	resourceTypes = make(map[string]struct{})
	providers = make(map[string]struct{})

	for _, resource := range resources {
		resourceTypes[resource.Type] = struct{}{}
		if parts := strings.SplitN(resource.Type, "_", 2); len(parts) > 0 {
			providers[parts[0]] = struct{}{}
		}
	}

	return resourceTypes, providers
}

// extractModuleNames extracts unique module names from the state.
func extractModuleNames(rootModule struct {
	Resources []struct {
		Type string `json:"type"`
	} `json:"resources"`
	ChildModules []struct {
		Address string `json:"address"`
	} `json:"child_modules"`
}) map[string]struct{} {
	modules := make(map[string]struct{})
	// Add root module
	modules[""] = struct{}{}
	// Add child modules
	for _, module := range rootModule.ChildModules {
		modules[module.Address] = struct{}{}
	}
	return modules
}

// extractSensitiveOutputs extracts sensitive output names from the state.
func extractSensitiveOutputs(outputs map[string]struct {
	Sensitive bool `json:"sensitive"`
}) map[string]bool {
	sensitiveOutputs := make(map[string]bool)
	for name, output := range outputs {
		if output.Sensitive {
			sensitiveOutputs[name] = true
		}
	}
	return sensitiveOutputs
}

// mapToSlice converts a map to a slice of its keys.
func mapToSlice[T comparable](m map[T]struct{}) []T {
	result := make([]T, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Read executes the data source and updates the state.
func (d *TerrapwnerTfstateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerTfstateDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute terraform show -json
	result, err := utils.Execute(ctx, "terraform", []string{"show", "-json"}, 30*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read state",
			err.Error(),
		)
		return
	}

	// Parse the JSON state
	var state state
	if err := json.Unmarshal([]byte(result.Stdout), &state); err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse state JSON",
			err.Error(),
		)
		return
	}

	// Extract information from the state
	resourceTypes, providers := extractResourceInfo(state.Values.RootModule.Resources)
	modules := extractModuleNames(state.Values.RootModule)
	sensitiveOutputs := extractSensitiveOutputs(state.Values.Outputs)

	// Convert maps to slices
	uniqueTypes := mapToSlice(resourceTypes)
	uniqueProviders := mapToSlice(providers)
	uniqueModules := mapToSlice(modules)

	// Update the model with the results
	data.Success = types.BoolValue(true)
	data.RawJSON = types.StringValue(result.Stdout)
	data.ResourceCount = types.Int64Value(int64(len(state.Values.RootModule.Resources)))

	// Convert to Terraform types
	typesList, diags := types.ListValueFrom(ctx, types.StringType, uniqueTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ResourceTypes = typesList

	providersList, diags := types.ListValueFrom(ctx, types.StringType, uniqueProviders)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Providers = providersList

	modulesList, diags := types.ListValueFrom(ctx, types.StringType, uniqueModules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Modules = modulesList

	outputsMap, diags := types.MapValueFrom(ctx, types.BoolType, sensitiveOutputs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.SensitiveOutputs = outputsMap

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
