// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	downloadTimeout = 10 * time.Second
)

// GetUserAgent returns a consistent User-Agent string for all HTTP requests.
func GetUserAgent() string {
	return fmt.Sprintf("terrapwner (%s; %s; go%s)", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// DownloadFile downloads a file from the given URL and returns the path to the downloaded file.
func DownloadFile(ctx context.Context, url string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "terrapwner-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", GetUserAgent())

	// Send the request with timeout
	client := &http.Client{
		Timeout: downloadTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	// Copy the response body to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Get the path of the temporary file
	filePath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return filePath, nil
}
