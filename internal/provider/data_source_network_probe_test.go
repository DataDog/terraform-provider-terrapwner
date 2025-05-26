// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerrapwnerNetworkProbeDataSource(t *testing.T) {
	t.Parallel() // Mark test as parallel to ensure isolation

	// Start a TCP listener for testing
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer tcpListener.Close()

	// Start a UDP listener for testing
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to start UDP listener: %v", err)
	}
	defer udpConn.Close()

	// Extract host and port from TCP listener
	tcpHost, tcpPortStr, err := net.SplitHostPort(tcpListener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to split TCP address: %v", err)
	}
	// Extract host and port from UDP listener
	udpHost, udpPortStr, err := net.SplitHostPort(udpConn.LocalAddr().String())
	if err != nil {
		t.Fatalf("Failed to split UDP address: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test successful TCP connection
			{
				Config: providerConfig + fmt.Sprintf(`
data "terrapwner_network_probe" "test" {
  type       = "tcp"
  host       = %q
  port       = %s
  timeout    = 1
}
`, tcpHost, tcpPortStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "true"),
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "fail_reason", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test successful UDP connection
			{
				Config: providerConfig + fmt.Sprintf(`
data "terrapwner_network_probe" "test" {
  type       = "udp"
  host       = %q
  port       = %s
  timeout    = 1
}
`, udpHost, udpPortStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "true"),
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "fail_reason", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test connection to non-existent port
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type       = "tcp"
  host       = "127.0.0.1"
  port       = 1
  timeout    = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "false"),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "fail_reason"),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test connection timeout
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type       = "tcp"
  host       = "10.255.255.1"
  port       = 80
  timeout    = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "fail_reason", "TCP connection failed: dial tcp 10.255.255.1:80: i/o timeout"),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test DNS resolution
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type       = "dns"
  host       = "localhost"
  timeout    = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "true"),
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "fail_reason", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test DNS resolution failure
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type           = "dns"
  host           = "this.domain.does.not.exist.example"
  timeout    	 = 1
  expect_success = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_network_probe.test", "success", "false"),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "fail_reason"),
					resource.TestCheckResourceAttrSet("data.terrapwner_network_probe.test", "duration_ms"),
				),
			},
			// Test invalid probe type
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type = "invalid"
  host = "127.0.0.1"
}
`,
				ExpectError: regexp.MustCompile(`unsupported probe type: invalid`),
			},
			// Test missing port for TCP
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type = "tcp"
  host = "127.0.0.1"
}
`,
				ExpectError: regexp.MustCompile("port is required for tcp/udp probes"),
			},
			// Test invalid port number
			{
				Config: providerConfig + `
data "terrapwner_network_probe" "test" {
  type = "tcp"
  host = "127.0.0.1"
  port = 70000
}
`,
				ExpectError: regexp.MustCompile("port must be between 1 and 65535"),
			},
		},
	})
}
