// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// ExecResult represents the result of a command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Execute executes a command with a timeout and returns the result.
func Execute(ctx context.Context, command string, args []string, timeout time.Duration) (*ExecResult, error) {
	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(ctx, command, args...)

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for the command to complete
	waitErr := cmd.Wait()

	// Check if context was cancelled or timed out
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Create the result
	result := &ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	// Handle command completion
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return nil, fmt.Errorf("command failed: %w", waitErr)
	}

	return result, nil
}
