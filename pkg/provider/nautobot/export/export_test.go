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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestValidateInventory(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name      string
		inv       *devicetypes.Inventory
		expectErr bool
		errorMsg  string
	}{
		{
			name: "valid inventory with exportable device",
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {Name: "server-01", Type: devicetypes.Type("node")},
				},
			},
			expectErr: false,
		},
		{
			name:      "nil inventory returns error",
			inv:       nil,
			expectErr: true,
			errorMsg:  "inventory is nil",
		},
		{
			name: "empty devices returns error",
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			},
			expectErr: true,
			errorMsg:  "inventory is empty",
		},
		{
			name: "only system devices returns error",
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {Name: "sys-root", Type: devicetypes.Type("system")},
				},
			},
			expectErr: true,
			errorMsg:  "no exportable devices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInventory(tt.inv)

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
		})
	}
}
