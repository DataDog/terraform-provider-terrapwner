// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"terraform-provider-terrapwner/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerrapwnerRemoteExecDataSource_InvalidURL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "terrapwner_remote_exec" "test" {
  url         = "https://invalid-url-that-does-not-exist.com/script.sh"
  interpreter = "bash"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "exit_code", "-1"),
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "stdout", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_remote_exec.test", "stderr"),
				),
			},
		},
	})
}

func TestAccTerrapwnerRemoteExecDataSource_FailOnError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "terrapwner_remote_exec" "test" {
  url          = "https://invalid-url-that-does-not-exist.com/script.sh"
  interpreter  = "bash"
  fail_on_error = true
}
`,
				ExpectError: regexp.MustCompile("failed to download file"),
			},
		},
	})
}

func TestAccTerrapwnerRemoteExecDataSource_InvalidInterpreter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "terrapwner_remote_exec" "test" {
  url         = "https://invalid-url-that-does-not-exist.com/script.sh"
  interpreter = "nonexistent-interpreter"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "success", "false"),
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "exit_code", "-1"),
					resource.TestCheckResourceAttr("data.terrapwner_remote_exec.test", "stdout", ""),
					resource.TestCheckResourceAttrSet("data.terrapwner_remote_exec.test", "stderr"),
				),
			},
		},
	})
}

func TestExecuteScript(t *testing.T) {
	// Create a temporary directory for test scripts
	tempDir, err := os.MkdirTemp("", "terrapwner-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		script      string
		interpreter string
		args        []string
		wantErr     bool
		checkResult func(t *testing.T, result *utils.ExecResult)
	}{
		{
			name: "successful script execution",
			script: `#!/bin/sh
echo "Hello from test script"
exit 0
`,
			interpreter: "/bin/sh",
			args:        []string{},
			wantErr:     false,
			checkResult: func(t *testing.T, result *utils.ExecResult) {
				if result.ExitCode != 0 {
					t.Errorf("expected exit code 0, got %d", result.ExitCode)
				}
				if result.Stdout != "Hello from test script\n" {
					t.Errorf("expected stdout 'Hello from test script\n', got '%s'", result.Stdout)
				}
				if result.Stderr != "" {
					t.Errorf("expected empty stderr, got '%s'", result.Stderr)
				}
			},
		},
		{
			name: "script with arguments",
			script: `#!/bin/sh
echo "Hello, $1!"
exit 0
`,
			interpreter: "/bin/sh",
			args:        []string{"world"},
			wantErr:     false,
			checkResult: func(t *testing.T, result *utils.ExecResult) {
				if result.ExitCode != 0 {
					t.Errorf("expected exit code 0, got %d", result.ExitCode)
				}
				if result.Stdout != "Hello, world!\n" {
					t.Errorf("expected stdout 'Hello, world!\n', got '%s'", result.Stdout)
				}
			},
		},
		{
			name: "script with error",
			script: `#!/bin/sh
echo "This script will fail" >&2
exit 1
`,
			interpreter: "/bin/sh",
			args:        []string{},
			wantErr:     false, // We don't want an error from executeScript, just a non-zero exit code
			checkResult: func(t *testing.T, result *utils.ExecResult) {
				if result.ExitCode != 1 {
					t.Errorf("expected exit code 1, got %d", result.ExitCode)
				}
				if result.Stderr != "This script will fail\n" {
					t.Errorf("expected stderr 'This script will fail\n', got '%s'", result.Stderr)
				}
			},
		},
		{
			name: "invalid interpreter",
			script: `#!/bin/sh
echo "This should not run"
`,
			interpreter: "nonexistent-interpreter",
			args:        []string{},
			wantErr:     true,
			checkResult: nil,
		},
		{
			name: "script timeout",
			script: `#!/bin/sh
sleep 2
echo "This should timeout"
`,
			interpreter: "/bin/sh",
			args:        []string{},
			wantErr:     true,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the script file
			scriptPath := filepath.Join(tempDir, "test.sh")
			if err := os.WriteFile(scriptPath, []byte(tt.script), 0755); err != nil {
				t.Fatalf("Failed to write test script: %v", err)
			}

			// Create a context with a short timeout for the timeout test
			ctx := context.Background()
			if tt.name == "script timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}

			// Execute the script
			result, err := executeScript(ctx, scriptPath, tt.interpreter, tt.args)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check result
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
