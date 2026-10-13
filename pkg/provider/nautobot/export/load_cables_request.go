/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// cableRequest builds the create request for a cable whose ends are resolved:
// termination types default to dcim.interface, the type comes from the
// cable's own fields or its slug, and length, colour and unit are sent only
// when the cable carries them.
func cableRequest(cable *devicetypes.CaniCableType, a, b cableEnd, statusID uuid.UUID) (nautobotapi.WritableCableRequest, error) {
	termAID, termBID := openapi_types.UUID(a.iface.ID), openapi_types.UUID(b.iface.ID)
	termAType, termBType := terminationType(cable.TerminationAType), terminationType(cable.TerminationBType)
	req := nautobotapi.WritableCableRequest{
		TerminationAId: &termAID, TerminationAType: &termAType,
		TerminationBId: &termBID, TerminationBType: &termBType,
		Label: &cable.Label,
		Type:  resolveCableType(cable),
	}
	if cable.Length != nil {
		length := int(*cable.Length)
		req.Length = &length
	}
	if err := setRefID(&req.Status, statusID); err != nil {
		return req, fmt.Errorf("set cable status reference: %w", err)
	}
	if cable.Color != "" {
		hex := colorNameToHex(cable.Color)
		req.Color = &hex
	}
	req.LengthUnit = lengthUnit(cable.LengthUnit)
	return req, nil
}

// terminationType is the explicit termination type when the cable has one,
// otherwise an interface.
func terminationType(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return "dcim.interface"
}

// lengthUnit maps a cable's unit to Nautobot's; empty or unknown stays unset.
func lengthUnit(unit string) *nautobotapi.LengthUnitEnum {
	var mapped nautobotapi.LengthUnitEnum
	switch unit {
	case "m":
		mapped = nautobotapi.LengthUnitEnumM
	case "cm":
		mapped = nautobotapi.LengthUnitEnumCm
	case "ft":
		mapped = nautobotapi.LengthUnitEnumFt
	case "in":
		mapped = nautobotapi.LengthUnitEnumIn
	default:
		return nil
	}
	return &mapped
}

// postCable creates the cable and marks both interfaces cabled in the cache,
// so a later cable on either port is reported as a conflict.
func (e *Exporter) postCable(ctx context.Context, cable *devicetypes.CaniCableType, a, b cableEnd, req nautobotapi.WritableCableRequest, result *LoadResult) error {
	resp, err := e.Client.DcimCablesCreateWithResponse(ctx, &nautobotapi.DcimCablesCreateParams{}, req)
	if err != nil {
		return fmt.Errorf(errFmtAPIError, err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s (cable: %s, ifaceA=%s, ifaceB=%s)",
			resp.StatusCode(), string(resp.Body), cableEndpoints(a, b), a.iface.ID, b.iface.ID)
	}
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		if created := uuid.UUID(*resp.JSON201.Id); created != uuid.Nil {
			a.iface.CableID, b.iface.CableID = created, created
		}
	}
	clog.Created("Created cable: %s (%s)", cable.Label, cableEndpoints(a, b))
	result.CablesCreated++
	return nil
}
