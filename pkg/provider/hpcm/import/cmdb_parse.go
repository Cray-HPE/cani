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
package import_

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchCmdb GETs every resource collection from the HPCM CMDB REST API and
// returns the populated Cmdb struct. baseURL should look like
// "https://host:8080/cmu/v1" (no trailing slash). If httpClient is nil,
// http.DefaultClient is used.
func FetchCmdb(ctx context.Context, baseURL string, httpClient *http.Client) (*Cmdb, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	var db Cmdb

	type endpoint struct {
		path   string
		target interface{}
	}

	endpoints := []endpoint{
		{"/controllers", &db.Controllers},
		{"/customgroups", &db.CustomGroups},
		{"/imagegroups", &db.ImageGroups},
		{"/managementcards", &db.ManagementCards},
		{"/networkgroups", &db.NetworkGroups},
		{"/networks", &db.Networks},
		{"/nics", &db.Nics},
		{"/nodes/templates", &db.NodeTemplates},
		{"/nodes", &db.Nodes},
		{"/systemgroups", &db.SystemGroups},
	}

	for _, ep := range endpoints {
		if err := fetchResource(ctx, httpClient, baseURL+ep.path, ep.target); err != nil {
			return nil, fmt.Errorf("fetching %s: %w", ep.path, err)
		}
	}

	return &db, nil
}

// fetchResource performs an HTTP GET and JSON-decodes the response body into target.
func fetchResource(ctx context.Context, client *http.Client, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// ParseCmdbJSON unmarshals a pre-dumped CMDB JSON blob into a Cmdb struct.
// The expected format matches the Cmdb struct's JSON tags:
//
//	{
//	  "controllers": [...],
//	  "customGroups": [...],
//	  "imageGroups": [...],
//	  "managementCards": [...],
//	  "networkGroups": [...],
//	  "networks": [...],
//	  "nics": [...],
//	  "nodeTemplates": [...],
//	  "nodes": [...],
//	  "systemGroups": [...]
//	}
func ParseCmdbJSON(data []byte) (*Cmdb, error) {
	var db Cmdb
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("parsing CMDB JSON: %w", err)
	}
	return &db, nil
}
