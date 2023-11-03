/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package csm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
)

func (csm *CSM) ExportJson(ctx context.Context, datastore inventory.Datastore, skipValidation bool) ([]byte, error) {
	currentSLSState, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to get the current SLS state"),
			err,
		)
	}

	modifiedState, _, _, err := csm.reconcileSlsChanges(currentSLSState, datastore)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to reconcile requested SLS changes with current SLS state"),
			err)
	}

	if !skipValidation {
		_, err = csm.TBV.Validate(modifiedState)
		if err != nil {
			return nil, fmt.Errorf("validation failed %v", err)
		}
	}

	j, err := json.MarshalIndent(modifiedState, "", "  ")
	if err != nil {
		return nil, err
	}
	return j, nil
}
