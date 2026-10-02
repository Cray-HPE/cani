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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// postPrefixWithFallback POSTs the prefix, retrying once without a location when
// Nautobot rejects the association (e.g. the location type does not permit
// prefixes).
func (e *Exporter) postPrefixWithFallback(
	ctx context.Context,
	prefix *devicetypes.CaniPrefix,
	req *nautobotapi.WritablePrefixRequest,
) (uuid.UUID, error) {
	id, code, body, err := e.postPrefix(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if code == http.StatusCreated {
		return id, nil
	}
	// An unsupported location type makes Nautobot reject the prefix — cleanly
	// with 400, or with a 500 when its location validation crashes. Either way,
	// retry once without the location so the prefix still exports (unscoped).
	if req.Location != nil && (code == http.StatusBadRequest || code == http.StatusInternalServerError) {
		clog.Warn("  Prefix %s: location not accepted, retrying without location", prefix.Prefix)
		req.Location = nil
		id, code, body, err = e.postPrefix(ctx, req)
		if err != nil {
			return uuid.Nil, fmt.Errorf("API error: %w", err)
		}
		if code == http.StatusCreated {
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("unexpected status %d: %s", code, body)
}

// postPrefix performs a single prefix create using a raw API call (the generated
// response parser mishandles Nautobot's UUID format). It returns the new ID (on
// 201), the HTTP status code, and the response body.
func (e *Exporter) postPrefix(ctx context.Context, req *nautobotapi.WritablePrefixRequest) (uuid.UUID, int, string, error) {
	httpResp, err := e.Client.IpamPrefixesCreate(ctx, &nautobotapi.IpamPrefixesCreateParams{}, *req)
	if err != nil {
		return uuid.Nil, 0, "", err
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusCreated {
		return uuid.Nil, httpResp.StatusCode, string(body), nil
	}

	var respObj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		return uuid.Nil, httpResp.StatusCode, string(body), fmt.Errorf("failed to parse response: %w", err)
	}
	id, err := uuid.Parse(respObj.ID)
	if err != nil {
		return uuid.Nil, httpResp.StatusCode, string(body), fmt.Errorf("failed to parse UUID from response: %w", err)
	}
	return id, httpResp.StatusCode, string(body), nil
}
