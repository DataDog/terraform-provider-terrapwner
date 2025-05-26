// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserAgent(t *testing.T) {
	t.Parallel()

	ua := GetUserAgent()
	expected := fmt.Sprintf("terrapwner (%s; %s; go%s)", runtime.GOOS, runtime.GOARCH, runtime.Version())
	assert.Equal(t, expected, ua)
}

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		handler       func(w http.ResponseWriter, r *http.Request)
		expectedError string
		checkFile     func(t *testing.T, path string)
	}{
		{
			name: "successful download",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, GetUserAgent(), r.Header.Get("User-Agent"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test content")) //nolint:errcheck
			},
			checkFile: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, "test content", string(content))
				assert.True(t, strings.HasPrefix(filepath.Base(path), "terrapwner-"))
			},
		},
		{
			name: "404 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedError: "failed to download file: status code 404",
		},
		{
			name: "500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "failed to download file: status code 500",
		},
		{
			name: "slow response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			expectedError: "context deadline exceeded",
		},
		{
			name: "large file",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Write 1MB of data
				data := make([]byte, 1024)
				for i := 0; i < 1024; i++ {
					w.Write(data) //nolint:errcheck
				}
			},
			checkFile: func(t *testing.T, path string) {
				info, err := os.Stat(path)
				require.NoError(t, err)
				assert.Equal(t, int64(1024*1024), info.Size())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// Download file
			path, err := DownloadFile(ctx, server.URL)

			// Clean up downloaded file if it exists
			if path != "" {
				defer os.Remove(path)
			}

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, path)

			if tt.checkFile != nil {
				tt.checkFile(t, path)
			}
		})
	}
}

func TestDownloadFile_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a server that sleeps before responding
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content")) //nolint:errcheck
	}))
	defer server.Close()

	// Create context and cancel it after a short delay
	ctx, cancel := context.WithCancel(context.Background())

	// Start download in a goroutine
	done := make(chan struct{})
	var downloadErr error
	var path string

	go func() {
		defer close(done)
		path, downloadErr = DownloadFile(ctx, server.URL)
	}()

	// Cancel context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for download to finish
	select {
	case <-done:
		assert.Error(t, downloadErr)
		assert.Contains(t, downloadErr.Error(), "context canceled")
		if path != "" {
			os.Remove(path)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("download did not cancel in time")
	}
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "empty url",
			url:  "",
		},
		{
			name: "invalid scheme",
			url:  "invalid://example.com",
		},
		{
			name: "malformed url",
			url:  "http://[::1]:namedport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path, err := DownloadFile(context.Background(), tt.url)
			assert.Error(t, err)
			if path != "" {
				os.Remove(path)
			}
		})
	}
}
