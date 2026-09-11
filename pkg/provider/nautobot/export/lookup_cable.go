/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"net/http"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// GetCableByTerminations checks if a cable exists between two interfaces
// It checks both directions (A-B and B-A) since cables are bidirectional
// Returns the cable CachedItem if found, nil if not found
func (c *LookupCache) GetCableByTerminations(interfaceAID, interfaceBID uuid.UUID) (*CachedItem, error) {
	if c.ctx == nil {
		return nil, fmt.Errorf("lookup cache context not set, call SetContext first")
	}

	// Search for cables with termination_a matching interfaceA
	aID := openapi_types.UUID(interfaceAID)
	bID := openapi_types.UUID(interfaceBID)

	// Try A->B direction
	params := &nautobotapi.DcimCablesListParams{
		TerminationAId: &[]openapi_types.UUID{aID},
	}
	resp, err := c.client.DcimCablesListWithResponse(c.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query cables: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status %d querying cables", resp.StatusCode())
	}

	// Check if any cable has termination_b matching interfaceB
	for _, cable := range resp.JSON200.Results {
		if cable.TerminationBId != nil && *cable.TerminationBId == bID {
			cableID := uuid.UUID(*cable.Id)
			label := ""
			if cable.Label != nil {
				label = *cable.Label
			}
			return &CachedItem{ID: cableID, Name: label}, nil
		}
	}

	// Try B->A direction (cable may be stored in reverse)
	params2 := &nautobotapi.DcimCablesListParams{
		TerminationAId: &[]openapi_types.UUID{bID},
	}
	resp2, err := c.client.DcimCablesListWithResponse(c.ctx, params2)
	if err != nil {
		return nil, fmt.Errorf("failed to query cables: %w", err)
	}
	if resp2.StatusCode() != http.StatusOK || resp2.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status %d querying cables", resp2.StatusCode())
	}

	// Check if any cable has termination_b matching interfaceA
	for _, cable := range resp2.JSON200.Results {
		if cable.TerminationBId != nil && *cable.TerminationBId == aID {
			cableID := uuid.UUID(*cable.Id)
			label := ""
			if cable.Label != nil {
				label = *cable.Label
			}
			return &CachedItem{ID: cableID, Name: label}, nil
		}
	}

	return nil, nil
}
