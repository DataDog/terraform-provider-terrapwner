// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerrapwnerEnvDumpDataSource(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR1", "test_value1")
	os.Setenv("TEST_VAR2", "test_value2")
	defer os.Unsetenv("TEST_VAR1")
	defer os.Unsetenv("TEST_VAR2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "terrapwner_env_dump" "test" {mask_values = false}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "id", "env_dump"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR1", "test_value1"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR2", "test_value2"),
				),
			},
		},
	})
}

func TestAccTerrapwnerEnvDumpDataSource_EmptyEnv(t *testing.T) {
	// Ensure no test variables are set
	os.Unsetenv("TEST_VAR1")
	os.Unsetenv("TEST_VAR2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing with empty environment
			{
				Config: providerConfig + `
data "terrapwner_env_dump" "test" {mask_values = false}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "id", "env_dump"),
					resource.TestCheckNoResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR1"),
					resource.TestCheckNoResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR2"),
				),
			},
		},
	})
}

func TestAccTerrapwnerEnvDumpDataSource_WithExistingEnv(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR1", "test_value1")
	os.Setenv("TEST_VAR2", "test_value2")
	defer os.Unsetenv("TEST_VAR1")
	defer os.Unsetenv("TEST_VAR2")

	// Set some existing environment variables that should be present
	os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin")
	os.Setenv("HOME", "/home/testuser")
	defer os.Unsetenv("PATH")
	defer os.Unsetenv("HOME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing with existing environment variables
			{
				Config: providerConfig + `
data "terrapwner_env_dump" "test" {mask_values = false}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "id", "env_dump"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR1", "test_value1"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR2", "test_value2"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.PATH", "/usr/local/bin:/usr/bin:/bin"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.HOME", "/home/testuser"),
				),
			},
		},
	})
}

func TestAccTerrapwnerEnvDumpDataSource_MaskedValues(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR1", "test_value1")
	os.Setenv("TEST_VAR2", "test_value2")
	defer os.Unsetenv("TEST_VAR1")
	defer os.Unsetenv("TEST_VAR2")

	// Set some existing environment variables that should be present
	os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin")
	os.Setenv("HOME", "/home/testuser")
	defer os.Unsetenv("PATH")
	defer os.Unsetenv("HOME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing with existing environment variables
			{
				Config: providerConfig + `
data "terrapwner_env_dump" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "id", "env_dump"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR1", "<REDACTED>"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.TEST_VAR2", "<REDACTED>"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.PATH", "<REDACTED>"),
					resource.TestCheckResourceAttr("data.terrapwner_env_dump.test", "vars.HOME", "<REDACTED>"),
				),
			},
		},
	})
}
