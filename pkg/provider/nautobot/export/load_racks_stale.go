/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// pruneStaleRackID drops a rack's stored Nautobot ID when the destination
// definitively has no rack with it, so the rack is resolved by name and
// created when missing. An inventory imported from one Nautobot carries that
// Nautobot's IDs; trusted on a fresh destination, every rack would be
// skipped as already exported and every device placed in one would fail. A
// transient failure keeps the ID, as pruneStaleDeviceID does.
func (e *Exporter) pruneStaleRackID(ctx context.Context, rack *devicetypes.CaniRackType) {
	nid, ok := rack.ExternalIDs[externalIDKeyNautobot]
	if !ok || nid == uuid.Nil {
		return
	}
	if exists, err := e.rackExistsByID(ctx, nid); err == nil && !exists {
		clog.Warn("Stored Nautobot ID %s for rack %q is not in this Nautobot; resolving it by name", nid, rack.Name)
		delete(rack.ExternalIDs, externalIDKeyNautobot)
	}
}

// rackExistsByID reports whether a rack with the given ID exists in Nautobot.
// A 404 returns (false, nil); any other failure returns an error so it is
// never read as absence.
func (e *Exporter) rackExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	resp, err := e.Client.DcimRacksRetrieveWithResponse(ctx, id, &nautobotapi.DcimRacksRetrieveParams{})
	if err != nil {
		return false, fmt.Errorf(errFmtAPIError, err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status %d checking rack %s", resp.StatusCode(), id)
	}
}
