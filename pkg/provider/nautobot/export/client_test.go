/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package export

import (
	"strings"
	"testing"
)

// TestNewNautobotClient verifies the Nautobot API client constructor accepts a
// valid URL/token pair and rejects calls that omit either credential.
//
// Why it matters: every export to Nautobot flows through this client, so a
// missing endpoint or token must fail fast at construction rather than surface
// as a confusing error partway through pushing devices.
// Inputs: a base URL and an API token. Outputs: a *NautobotClient and an error.
// Data choice: a localhost API URL stands in for a real Nautobot instance; the
// two empty-string cases isolate each required credential so the test can assert
// the specific validation message instead of a generic failure.
func TestNewNautobotClient(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		token     string
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "valid url and token creates client",
			url:       "http://localhost:8080/api",
			token:     "abc123",
			expectErr: false,
		},
		{
			name:      "empty url returns error",
			url:       "",
			token:     "abc123",
			expectErr: true,
			errorMsg:  "nautobot URL is required",
		},
		{
			name:      "empty token returns error",
			url:       "http://localhost:8080/api",
			token:     "",
			expectErr: true,
			errorMsg:  "nautobot API token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewNautobotClient(tt.url, tt.token)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if client == nil {
				t.Error("expected client to be non-nil")
			}
		})
	}
}
