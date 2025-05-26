// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/datadog/terraform-provider-terrapwner/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// defaultCommandTimeout is the default timeout for command execution.
	// This is set to 30 seconds to allow for longer-running commands while still
	// maintaining reasonable plan performance.
	defaultCommandTimeout = 30 * time.Second
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TerrapwnerLocalExecDataSource{}
	_ datasource.DataSourceWithConfigure = &TerrapwnerLocalExecDataSource{}
)

// TerrapwnerLocalExecDataSourceModel describes the data source data model.
type TerrapwnerLocalExecDataSourceModel struct {
	Command       types.List   `tfsdk:"command"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	ExpectSuccess types.Bool   `tfsdk:"expect_success"`
	FailOnError   types.Bool   `tfsdk:"fail_on_error"`
	Success       types.Bool   `tfsdk:"success"`
	Stdout        types.String `tfsdk:"stdout"`
	Stderr        types.String `tfsdk:"stderr"`
	ExitCode      types.Int64  `tfsdk:"exit_code"`
	FailReason    types.String `tfsdk:"fail_reason"`
	DurationMs    types.Int64  `tfsdk:"duration_ms"`
}

// NewTerrapwnerLocalExecDataSource is a helper function to simplify the provider implementation.
func NewTerrapwnerLocalExecDataSource() datasource.DataSource {
	return &TerrapwnerLocalExecDataSource{}
}

// TerrapwnerLocalExecDataSource is the data source implementation.
type TerrapwnerLocalExecDataSource struct{}

// Metadata returns the data source type name.
func (d *TerrapwnerLocalExecDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_exec"
}

// Schema defines the schema for the data source.
func (d *TerrapwnerLocalExecDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes a local command and captures its output, exit code, and runtime details. " +
			"This data source is used in CI/CD pipeline assessments to determine what can be executed inside the Terraform runtime environment. " +
			"Commands are executed with a configurable timeout (default: 30 seconds).",
		Attributes: map[string]schema.Attribute{
			"command": schema.ListAttribute{
				Description: "The command to execute as a list of strings. The first element is the executable, and the rest are arguments.",
				ElementType: types.StringType,
				Required:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in seconds for command execution (default: 30).",
				Optional:    true,
			},
			"expect_success": schema.BoolAttribute{
				Description: "Whether an exit code of 0 is expected (default: true).",
				Optional:    true,
			},
			"fail_on_error": schema.BoolAttribute{
				Description: "Whether to fail the Terraform operation if the command fails (default: false).",
				Optional:    true,
			},
			"success": schema.BoolAttribute{
				Description: "True if the command exited with code 0.",
				Computed:    true,
			},
			"stdout": schema.StringAttribute{
				Description: "Captured standard output.",
				Computed:    true,
			},
			"stderr": schema.StringAttribute{
				Description: "Captured standard error.",
				Computed:    true,
			},
			"exit_code": schema.Int64Attribute{
				Description: "Exit code of the process.",
				Computed:    true,
			},
			"fail_reason": schema.StringAttribute{
				Description: "If execution fails or times out, this contains the error.",
				Computed:    true,
			},
			"duration_ms": schema.Int64Attribute{
				Description: "Total execution time in milliseconds.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TerrapwnerLocalExecDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// No configuration needed
}

// Read executes the command and updates the state.
func (d *TerrapwnerLocalExecDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerLocalExecDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values
	if data.Timeout.IsNull() {
		data.Timeout = types.Int64Value(int64(defaultCommandTimeout.Seconds()))
	}
	if data.ExpectSuccess.IsNull() {
		data.ExpectSuccess = types.BoolValue(true)
	}
	if data.FailOnError.IsNull() {
		data.FailOnError = types.BoolValue(false)
	}

	// Start timing
	startTime := time.Now()

	// Convert command from List to []string
	var command []string
	resp.Diagnostics.Append(data.Command.ElementsAs(ctx, &command, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(command) == 0 {
		resp.Diagnostics.AddError(
			"Invalid command",
			"Command list cannot be empty",
		)
		return
	}

	// Execute the command with the configured timeout
	result, err := utils.Execute(
		ctx,
		command[0],
		command[1:],
		time.Duration(data.Timeout.ValueInt64())*time.Second,
	)
	if err != nil {
		data.Success = types.BoolValue(false)
		data.FailReason = types.StringValue(fmt.Sprintf("Failed to execute command: %v", err))
		data.ExitCode = types.Int64Value(-1)
		data.DurationMs = types.Int64Value(time.Since(startTime).Milliseconds())
		if data.FailOnError.ValueBool() {
			resp.Diagnostics.AddError(
				"Command execution failed",
				data.FailReason.ValueString(),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Set the results
	data.Success = types.BoolValue(result.ExitCode == 0)
	data.Stdout = types.StringValue(result.Stdout)
	data.Stderr = types.StringValue(result.Stderr)
	data.ExitCode = types.Int64Value(int64(result.ExitCode))
	data.FailReason = types.StringValue("")
	data.DurationMs = types.Int64Value(time.Since(startTime).Milliseconds())

	// Check if we should fail on non-zero exit code
	if !data.Success.ValueBool() && data.FailOnError.ValueBool() {
		resp.Diagnostics.AddError(
			"Command failed",
			fmt.Sprintf("Command exited with code %d: %s", result.ExitCode, result.Stderr),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
