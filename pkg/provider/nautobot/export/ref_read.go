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
	"encoding/json"
	"fmt"
	"sync"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// refID extracts a UUID from any generated per-field reference Id union,
// returning uuid.Nil for absent or integer-valued references. Malformed union
// output is reported once per generated type and reads as uuid.Nil.
func refID(id json.Marshaler) uuid.UUID {
	got, err := nautobotapi.RefUUIDChecked(id)
	if err != nil {
		warnReferenceShape(id, err)
	}
	return got
}

// tenantRefID extracts a UUID from any generated per-field reference Id union.
func tenantRefID(id json.Marshaler) uuid.UUID {
	return refID(id)
}

var referenceShapeWarnings sync.Map

func warnReferenceShape(ref any, err error) {
	if _, loaded := referenceShapeWarnings.LoadOrStore(fmt.Sprintf("%T", ref), struct{}{}); loaded {
		return
	}
	clog.Warn("generated reference shape mismatch: %v", err)
}
