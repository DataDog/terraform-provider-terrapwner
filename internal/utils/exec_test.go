// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		command       string
		args          []string
		timeout       time.Duration
		expectedError bool
		checkResult   func(t *testing.T, result *ExecResult)
	}{
		{
			name:    "echo command",
			command: "echo",
			args:    []string{"hello"},
			timeout: 5 * time.Second,
			checkResult: func(t *testing.T, result *ExecResult) {
				assert.Equal(t, "hello\n", result.Stdout)
				assert.Equal(t, "", result.Stderr)
				assert.Equal(t, 0, result.ExitCode)
			},
		},
		{
			name:    "command with stderr",
			command: "sh",
			args:    []string{"-c", "echo error >&2"},
			timeout: 5 * time.Second,
			checkResult: func(t *testing.T, result *ExecResult) {
				assert.Equal(t, "", result.Stdout)
				assert.Equal(t, "error\n", result.Stderr)
				assert.Equal(t, 0, result.ExitCode)
			},
		},
		{
			name:    "command with non-zero exit code",
			command: "sh",
			args:    []string{"-c", "exit 1"},
			timeout: 5 * time.Second,
			checkResult: func(t *testing.T, result *ExecResult) {
				assert.Equal(t, "", result.Stdout)
				assert.Equal(t, "", result.Stderr)
				assert.Equal(t, 1, result.ExitCode)
			},
		},
		{
			name:          "invalid command",
			command:       "nonexistentcommand",
			args:          []string{},
			timeout:       5 * time.Second,
			expectedError: true,
		},
		{
			name:          "command timeout",
			command:       "sleep",
			args:          []string{"10"},
			timeout:       100 * time.Millisecond,
			expectedError: true,
		},
		{
			name:    "command with both stdout and stderr",
			command: "sh",
			args:    []string{"-c", "echo out; echo err >&2"},
			timeout: 5 * time.Second,
			checkResult: func(t *testing.T, result *ExecResult) {
				assert.Equal(t, "out\n", result.Stdout)
				assert.Equal(t, "err\n", result.Stderr)
				assert.Equal(t, 0, result.ExitCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := Execute(context.Background(), tt.command, tt.args, tt.timeout)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			tt.checkResult(t, result)
		})
	}
}

func TestExecute_CancelContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running command
	done := make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		defer close(done)
		_, err := Execute(ctx, "sleep", []string{"10"}, 20*time.Second)
		errCh <- err
	}()

	// Cancel the context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to finish
	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	case <-time.After(2 * time.Second):
		t.Fatal("command was not cancelled")
	}
}
