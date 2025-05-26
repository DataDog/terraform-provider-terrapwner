// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerrapwnerLocalExecDataSource(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp(t.TempDir(), "terrapwner-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if err := os.WriteFile(tempFile.Name(), []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Set PATH to include standard Unix command locations
	t.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test successful command execution
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["ls", "-l"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.terrapwner_local_exec.test", "stdout"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "stderr", ""),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "true"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "0"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", ""),
				),
			},
			// Test command with stderr output
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["ls", "nonexistent_file"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "stdout", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_local_exec.test", "stderr"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "1"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", ""),
				),
			},
			// Test command with both stdout and stderr
			{
				Config: providerConfig + fmt.Sprintf(`
data "terrapwner_local_exec" "test" {
  command = ["ls", "-l", "nonexistent_file", %q]
}
`, tempFile.Name()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.terrapwner_local_exec.test", "stdout"),
					resource.TestCheckResourceAttrSet("data.terrapwner_local_exec.test", "stderr"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "1"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", ""),
				),
			},
		},
	})
}

func TestAccTerrapwnerLocalExecDataSource_Failures(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non-existent command
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["this_command_does_not_exist"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "-1"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", "Failed to execute command: failed to start command: exec: \"this_command_does_not_exist\": executable file not found in $PATH"),
				),
			},
			// Test command that returns non-zero exit code
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["false"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "1"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", ""),
				),
			},
		},
	})
}

func TestAccTerrapwnerLocalExecDataSource_Timeout(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test command timeout
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["sleep", "10"]
  timeout = 1
  expect_success = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "-1"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", "Failed to execute command: context deadline exceeded"),
				),
			},
		},
	})
}

func TestAccTerrapwnerLocalExecDataSource_FailOnError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test fail_on_error with failing command
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["ls", "nonexistent_file"]
  fail_on_error = true
}
`,
				ExpectError: regexp.MustCompile("Command failed"),
			},
		},
	})
}

func TestAccTerrapwnerLocalExecDataSource_CurrentDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file in the temporary directory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to the temporary directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test command execution in current directory
			{
				Config: providerConfig + `
data "terrapwner_local_exec" "test" {
  command = ["ls", "-l", "test.txt"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.terrapwner_local_exec.test", "stdout"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "stderr", ""),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "success", "true"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "exit_code", "0"),
					resource.TestCheckResourceAttr("data.terrapwner_local_exec.test", "fail_reason", ""),
				),
			},
		},
	})
}
