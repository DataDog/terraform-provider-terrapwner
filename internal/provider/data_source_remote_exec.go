// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/datadog/terraform-provider-terrapwner/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TerrapwnerRemoteExecDataSource{}
	_ datasource.DataSourceWithConfigure = &TerrapwnerRemoteExecDataSource{}
)

// NewTerrapwnerRemoteExecDataSource is a helper function to simplify the provider implementation.
func NewTerrapwnerRemoteExecDataSource() datasource.DataSource {
	return &TerrapwnerRemoteExecDataSource{}
}

// TerrapwnerRemoteExecDataSource is the data source implementation.
type TerrapwnerRemoteExecDataSource struct{}

// TerrapwnerRemoteExecDataSourceModel describes the data source data model.
type TerrapwnerRemoteExecDataSourceModel struct {
	URL           types.String `tfsdk:"url"`
	Interpreter   types.String `tfsdk:"interpreter"`
	Args          types.List   `tfsdk:"args"`
	ExpectSuccess types.Bool   `tfsdk:"expect_success"`
	FailOnError   types.Bool   `tfsdk:"fail_on_error"`
	Success       types.Bool   `tfsdk:"success"`
	Stdout        types.String `tfsdk:"stdout"`
	Stderr        types.String `tfsdk:"stderr"`
	ExitCode      types.Int64  `tfsdk:"exit_code"`
}

// Configure adds the provider configured client to the data source.
func (d *TerrapwnerRemoteExecDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// No configuration needed
}

// Metadata returns the data source type name.
func (d *TerrapwnerRemoteExecDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_exec"
}

// Schema defines the schema for the data source.
func (d *TerrapwnerRemoteExecDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Downloads and executes a script from a URL.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "URL of the script to download and execute.",
				Required:    true,
			},
			"interpreter": schema.StringAttribute{
				Description: "Interpreter to use for executing the script (e.g., bash, python, powershell).",
				Required:    true,
			},
			"args": schema.ListAttribute{
				Description: "Arguments to pass to the script.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"expect_success": schema.BoolAttribute{
				Description: "Whether the script is expected to exit with code 0. If true, a non-zero exit code will result in an error.",
				Optional:    true,
			},
			"fail_on_error": schema.BoolAttribute{
				Description: "Whether to fail on any error (download or execution). If false, the data source will continue with default values.",
				Optional:    true,
			},
			"success": schema.BoolAttribute{
				Description: "Whether the script executed successfully.",
				Computed:    true,
			},
			"stdout": schema.StringAttribute{
				Description: "Standard output of the script.",
				Computed:    true,
			},
			"stderr": schema.StringAttribute{
				Description: "Standard error of the script.",
				Computed:    true,
			},
			"exit_code": schema.Int64Attribute{
				Description: "Exit code of the script.",
				Computed:    true,
			},
		},
	}
}

// downloadScript downloads a script from the given URL, makes it executable, and returns the path.
func downloadScript(ctx context.Context, url string) (string, error) {
	// Download the script using the generic download function
	scriptPath, err := utils.DownloadFile(ctx, url)
	if err != nil {
		return "", err
	}

	// Make the script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}

	return scriptPath, nil
}

// executeScript executes a script with the given interpreter and arguments.
func executeScript(ctx context.Context, scriptPath string, interpreter string, args []string) (*utils.ExecResult, error) {
	// Execute the script with the interpreter using utils package
	result, err := utils.Execute(ctx, interpreter, append([]string{scriptPath}, args...), 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	return result, nil
}

// Read executes the script and updates the state.
func (d *TerrapwnerRemoteExecDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerRemoteExecDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default value for fail_on_error to false if not provided
	if data.FailOnError.IsNull() {
		data.FailOnError = types.BoolValue(false)
	}

	// Convert args to []string
	var args []string
	if !data.Args.IsNull() {
		resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &args, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Download the script
	scriptPath, err := downloadScript(ctx, data.URL.ValueString())
	if err != nil {
		if !data.FailOnError.IsNull() && data.FailOnError.ValueBool() {
			resp.Diagnostics.AddError(
				"Failed to download script",
				err.Error(),
			)
			return
		}
		// Instead of failing, we'll set default values and add a warning
		resp.Diagnostics.AddWarning(
			"Failed to download script",
			err.Error(),
		)
		// Set default values
		data.Success = types.BoolValue(false)
		data.Stdout = types.StringValue("")
		data.Stderr = types.StringValue(err.Error())
		data.ExitCode = types.Int64Value(-1)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}
	defer os.Remove(scriptPath)

	// Execute the script
	result, err := executeScript(ctx, scriptPath, data.Interpreter.ValueString(), args)
	if err != nil {
		if !data.FailOnError.IsNull() && data.FailOnError.ValueBool() {
			resp.Diagnostics.AddError(
				"Failed to execute script",
				err.Error(),
			)
			return
		}
		// Instead of failing, we'll set default values and add a warning
		resp.Diagnostics.AddWarning(
			"Failed to execute script",
			err.Error(),
		)
		// Set default values
		data.Success = types.BoolValue(false)
		data.Stdout = types.StringValue("")
		data.Stderr = types.StringValue(err.Error())
		data.ExitCode = types.Int64Value(-1)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Update the model with the result
	data.Success = types.BoolValue(result.ExitCode == 0)
	data.Stdout = types.StringValue(result.Stdout)
	data.Stderr = types.StringValue(result.Stderr)
	data.ExitCode = types.Int64Value(int64(result.ExitCode))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
