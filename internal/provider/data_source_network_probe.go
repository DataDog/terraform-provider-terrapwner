// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TerrapwnerNetworkProbeDataSource{}
	_ datasource.DataSourceWithConfigure = &TerrapwnerNetworkProbeDataSource{}
)

// NewTerrapwnerNetworkProbeDataSource is a helper function to simplify the provider implementation.
func NewTerrapwnerNetworkProbeDataSource() datasource.DataSource {
	return &TerrapwnerNetworkProbeDataSource{}
}

// TerrapwnerNetworkProbeDataSource is the data source implementation.
type TerrapwnerNetworkProbeDataSource struct{}

// TerrapwnerNetworkProbeDataSourceModel describes the data source data model.
type TerrapwnerNetworkProbeDataSourceModel struct {
	Type          types.String `tfsdk:"type"`
	Host          types.String `tfsdk:"host"`
	Port          types.Int64  `tfsdk:"port"`
	ExpectSuccess types.Bool   `tfsdk:"expect_success"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	Success       types.Bool   `tfsdk:"success"`
	FailReason    types.String `tfsdk:"fail_reason"`
	DurationMs    types.Int64  `tfsdk:"duration_ms"`
}

// Configure adds the provider configured client to the data source.
func (d *TerrapwnerNetworkProbeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// No configuration needed
}

// Metadata returns the data source type name.
func (d *TerrapwnerNetworkProbeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_probe"
}

// Schema defines the schema for the data source.
func (d *TerrapwnerNetworkProbeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Probes network connectivity to a host using DNS resolution, TCP connection, UDP connection, or ICMP ping.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Type of probe to perform. Must be one of: dns, tcp, udp, icmp",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Host to probe (domain name or IP address)",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port to probe (required for tcp/udp probes, ignored for dns/icmp)",
				Optional:    true,
			},
			"expect_success": schema.BoolAttribute{
				Description: "Whether the probe is expected to succeed (default: true)",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in seconds (default: 5)",
				Optional:    true,
			},
			"success": schema.BoolAttribute{
				Description: "Whether the probe succeeded",
				Computed:    true,
			},
			"fail_reason": schema.StringAttribute{
				Description: "Reason for failure if probe failed",
				Computed:    true,
			},
			"duration_ms": schema.Int64Attribute{
				Description: "Duration of the probe in milliseconds",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *TerrapwnerNetworkProbeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TerrapwnerNetworkProbeDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	if state.ExpectSuccess.IsNull() {
		state.ExpectSuccess = types.BoolValue(true)
	}
	if state.Timeout.IsNull() {
		state.Timeout = types.Int64Value(5)
	}

	// Validate probe type
	if state.Type.IsNull() || state.Type.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid probe type", "type must be specified")
		return
	}

	// Validate host
	if state.Host.IsNull() || state.Host.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid host", "host must be specified")
		return
	}

	// Validate port for TCP/UDP probes
	if state.Type.ValueString() == "tcp" || state.Type.ValueString() == "udp" {
		if state.Port.IsNull() {
			resp.Diagnostics.AddError("Missing port", "port is required for tcp/udp probes")
			return
		}
		if state.Port.ValueInt64() < 1 || state.Port.ValueInt64() > 65535 {
			resp.Diagnostics.AddError("Invalid port", "port must be between 1 and 65535")
			return
		}
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(state.Timeout.ValueInt64())*time.Second)
	defer cancel()

	// Start timing
	start := time.Now()

	// Perform the appropriate probe
	var success bool
	var failReason string
	var err error

	switch state.Type.ValueString() {
	case "dns":
		success, failReason, err = probeDNS(ctx, state.Host.ValueString())
	case "tcp":
		success, failReason, err = probeTCP(ctx, state.Host.ValueString(), int(state.Port.ValueInt64()))
	case "udp":
		success, failReason, err = probeUDP(ctx, state.Host.ValueString(), int(state.Port.ValueInt64()))
	case "icmp":
		success, failReason, err = probeICMP(ctx, state.Host.ValueString())
	default:
		resp.Diagnostics.AddError("Invalid probe type", fmt.Sprintf("unsupported probe type: %s", state.Type.ValueString()))
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Probe error", err.Error())
		return
	}

	// Calculate duration
	duration := time.Since(start)

	// Set the state
	state.Success = types.BoolValue(success)
	state.FailReason = types.StringValue(failReason)
	state.DurationMs = types.Int64Value(duration.Milliseconds())

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// probeDNS performs a DNS resolution probe.
func probeDNS(ctx context.Context, host string) (bool, string, error) {
	_, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return false, fmt.Sprintf("DNS resolution failed: %v", err), nil
	}
	return true, "", nil
}

// probeTCP performs a TCP connection probe.
func probeTCP(ctx context.Context, host string, port int) (bool, string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return false, fmt.Sprintf("TCP connection failed: %v", err), nil
	}
	conn.Close()
	return true, "", nil
}

// probeUDP performs a UDP connection probe.
func probeUDP(ctx context.Context, host string, port int) (bool, string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	if err != nil {
		return false, fmt.Sprintf("UDP connection failed: %v", err), nil
	}
	conn.Close()
	return true, "", nil
}

// probeICMP performs an ICMP ping probe.
func probeICMP(ctx context.Context, host string) (bool, string, error) {
	// Resolve the host to get IP address
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return false, fmt.Sprintf("Failed to resolve host: %v", err), nil
	}
	if len(ips) == 0 {
		return false, "No IP addresses found", nil
	}

	// Try to ping each IP address
	for _, ip := range ips {
		conn, err := net.DialTimeout("ip4:icmp", ip.String(), 5*time.Second)
		if err != nil {
			continue // Try next IP if this one fails
		}
		conn.Close()
		return true, "", nil
	}

	return false, "ICMP ping failed for all IP addresses", nil
}
