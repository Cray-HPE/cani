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
	"context"
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// NautobotClient wraps the generated Nautobot API client with convenience methods
type NautobotClient struct {
	*nautobotapi.ClientWithResponses
	baseURL string
	token   string
}

// NewNautobotClient creates a new Nautobot API client with token authentication
func NewNautobotClient(url, token string) (*NautobotClient, error) {
	if url == "" {
		return nil, fmt.Errorf("nautobot URL is required")
	}
	if token == "" {
		return nil, fmt.Errorf("nautobot API token is required")
	}

	// Create auth header injector
	authEditor := nautobotapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Token "+token)
		return nil
	})

	client, err := nautobotapi.NewClientWithResponses(url, authEditor)
	if err != nil {
		return nil, fmt.Errorf("failed to create nautobot client: %w", err)
	}

	return &NautobotClient{
		ClientWithResponses: client,
		baseURL:             url,
		token:               token,
	}, nil
}

// TestConnection verifies that the client can connect to Nautobot
func (c *NautobotClient) TestConnection(ctx context.Context) error {
	// Try to list statuses as a simple connectivity test
	resp, err := c.ExtrasStatusesListWithResponse(ctx, &nautobotapi.ExtrasStatusesListParams{})
	if err != nil {
		return fmt.Errorf("failed to connect to nautobot: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("nautobot returned status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	return nil
}
