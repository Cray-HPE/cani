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
package add

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// resolveSpareLocation resolves --location for a --spare module.
func resolveSpareLocation(cmd *cli.Command, inventory *devicetypes.Inventory) (uuid.UUID, error) {
	locationArg, _ := cmd.Flags().GetString(flagLocation)
	if locationArg == "" {
		return uuid.Nil, fmt.Errorf("--%s requires --%s to name where the module is stored", flagSpare, flagLocation)
	}
	loc := inventory.FindLocationByNameOrID(locationArg)
	if loc == nil {
		return uuid.Nil, fmt.Errorf("location %q not found", locationArg)
	}
	return loc.ID, nil
}
