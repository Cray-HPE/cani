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
package devicetypes

import (
	"fmt"
	"strings"
)

// NautobotStatus represents a valid Nautobot status value.
type NautobotStatus string

const (
	StatusActive          NautobotStatus = "Active"
	StatusAvailable       NautobotStatus = "Available"
	StatusConnected       NautobotStatus = "Connected"
	StatusDecommissioned  NautobotStatus = "Decommissioned"
	StatusDecommissioning NautobotStatus = "Decommissioning"
	StatusDeprecated      NautobotStatus = "Deprecated"
	StatusDeprovisioning  NautobotStatus = "Deprovisioning"
	StatusDown            NautobotStatus = "Down"
	StatusEndOfLife       NautobotStatus = "End-of-Life"
	StatusExtendedSupport NautobotStatus = "Extended Support"
	StatusFailed          NautobotStatus = "Failed"
	StatusInventory       NautobotStatus = "Inventory"
	StatusMaintenance     NautobotStatus = "Maintenance"
	StatusOffline         NautobotStatus = "Offline"
	StatusPlanned         NautobotStatus = "Planned"
	StatusPrimary         NautobotStatus = "Primary"
	StatusProvisioning    NautobotStatus = "Provisioning"
	StatusReserved        NautobotStatus = "Reserved"
	StatusRetired         NautobotStatus = "Retired"
	StatusSecondary       NautobotStatus = "Secondary"
	StatusStaged          NautobotStatus = "Staged"
	StatusStaging         NautobotStatus = "Staging"
)

// AllStatuses contains every known Nautobot status, including Active
// and Staged which are used internally as constructor defaults.
var AllStatuses = []NautobotStatus{
	StatusActive,
	StatusAvailable,
	StatusConnected,
	StatusDecommissioned,
	StatusDecommissioning,
	StatusDeprecated,
	StatusDeprovisioning,
	StatusDown,
	StatusEndOfLife,
	StatusExtendedSupport,
	StatusFailed,
	StatusInventory,
	StatusMaintenance,
	StatusOffline,
	StatusPlanned,
	StatusPrimary,
	StatusProvisioning,
	StatusReserved,
	StatusRetired,
	StatusSecondary,
	StatusStaged,
	StatusStaging,
}

// AllUserStatuses lists statuses selectable by users at the CLI.
// Staged is excluded because it is assigned automatically by constructors.
var AllUserStatuses = []NautobotStatus{
	StatusActive,
	StatusAvailable,
	StatusConnected,
	StatusDecommissioned,
	StatusDecommissioning,
	StatusDeprecated,
	StatusDeprovisioning,
	StatusDown,
	StatusEndOfLife,
	StatusExtendedSupport,
	StatusFailed,
	StatusInventory,
	StatusMaintenance,
	StatusOffline,
	StatusPlanned,
	StatusPrimary,
	StatusProvisioning,
	StatusReserved,
	StatusRetired,
	StatusSecondary,
	StatusStaging,
}

// statusLookup is a case-insensitive index of all valid statuses.
var statusLookup map[string]NautobotStatus

func init() {
	statusLookup = make(map[string]NautobotStatus, len(AllStatuses))
	for _, s := range AllStatuses {
		statusLookup[strings.ToLower(string(s))] = s
	}
}

// ValidateUserStatus checks that s matches one of AllUserStatuses
// (case-insensitive) and returns the canonical Title-case form.
func ValidateUserStatus(s string) (NautobotStatus, error) {
	key := strings.ToLower(s)
	ns, ok := statusLookup[key]
	if !ok {
		return "", statusError(s)
	}
	// Reject Staged from user input (assigned automatically by constructors).
	if ns == StatusStaged {
		return "", statusError(s)
	}
	return ns, nil
}

// IsValidStatus returns true when s matches any known Nautobot status
// (including Active and Staged). The comparison is case-insensitive.
func IsValidStatus(s string) bool {
	_, ok := statusLookup[strings.ToLower(s)]
	return ok
}

// NormalizeStatus returns the canonical Title-case form of s if it is a
// known Nautobot status, or s unchanged when unrecognised.
func NormalizeStatus(s string) string {
	if ns, ok := statusLookup[strings.ToLower(s)]; ok {
		return string(ns)
	}
	return s
}

// UserStatusNames returns the display names of AllUserStatuses as a
// comma-separated string, suitable for error messages and help text.
func UserStatusNames() string {
	names := make([]string, len(AllUserStatuses))
	for i, s := range AllUserStatuses {
		names[i] = string(s)
	}
	return strings.Join(names, ", ")
}

func statusError(s string) error {
	return fmt.Errorf("invalid status %q: must be one of [%s]", s, UserStatusNames())
}
