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
package inventory

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
)

// Inventory is the top level object that represents the entire inventory
// This is what cani uses to represent the inventory
type Inventory struct {
	SchemaVersion SchemaVersion
	Provider      Provider
	Hardware      map[uuid.UUID]Hardware
}

func (i *Inventory) FilterHardware(filter func(Hardware) (bool, error)) (map[uuid.UUID]Hardware, error) {
	result := map[uuid.UUID]Hardware{}

	for id, hardware := range i.Hardware {
		ok, err := filter(hardware)
		if err != nil {
			return nil, err
		}

		if ok {
			result[id] = hardware
		}
	}

	return result, nil
}

func (i *Inventory) FilterHardwareByType(types ...hardwaretypes.HardwareType) map[uuid.UUID]Hardware {
	result, _ := i.FilterHardware(func(hardware Hardware) (bool, error) {
		for _, hadwareType := range types {
			if hardware.Type == hadwareType {
				return true, nil
			}
		}
		return false, nil
	})

	return result
}

func (i *Inventory) FilterHardwareByTypeStatus(status HardwareStatus, types ...hardwaretypes.HardwareType) map[uuid.UUID]Hardware {
	result, _ := i.FilterHardware(func(hardware Hardware) (bool, error) {
		for _, hardwareType := range types {
			if hardware.Status == status && hardware.Type == hardwareType {
				return true, nil
			}
		}
		return false, nil
	})

	return result
}

// Hardware is the smallest unit of inventory
// It has all the potential fields that hardware can have
type Hardware struct {
	ID               uuid.UUID
	Name             string                           `json:"Name,omitempty" yaml:"Name,omitempty" default:"" usage:"Friendly name"`
	Type             hardwaretypes.HardwareType       `json:"Type,omitempty" yaml:"Type,omitempty" default:"" usage:"Type"`
	DeviceTypeSlug   string                           `json:"DeviceTypeSlug,omitempty" yaml:"DeviceTypeSlug,omitempty" default:"" usage:"Hardware Type Library Device slug"`
	Vendor           string                           `json:"Vendor,omitempty" yaml:"Vendor,omitempty" default:"" usage:"Vendor"`
	Architecture     string                           `json:"Architecture,omitempty" yaml:"Architecture,omitempty" default:"" usage:"Architecture"`
	Model            string                           `json:"Model,omitempty" yaml:"Model,omitempty" default:"" usage:"Model"`
	Status           HardwareStatus                   `json:"Status,omitempty" yaml:"Status,omitempty" default:"Staged" usage:"Hardware can be [staged, provisioned, decomissioned]"`
	Properties       map[string]interface{}           `json:"Properties,omitempty" yaml:"Properties,omitempty" default:"" usage:"Properties"`
	ProviderMetadata map[Provider]ProviderMetadataRaw `json:"ProviderMetadata,omitempty" yaml:"ProviderMetadata,omitempty" default:"" usage:"ProviderMetadata"`

	Parent uuid.UUID `json:"Parent,omitempty" yaml:"Parent,omitempty" default:"00000000-0000-0000-0000-000000000000" usage:"Parent hardware"`
	// The following are derived from Parent
	Children     []uuid.UUID  `json:"Children,omitempty" yaml:"Children,omitempty"`
	LocationPath LocationPath `json:"LocationPath,omitempty" yaml:"LocationPath,omitempty"`

	LocationOrdinal *int
}

func (hardware *Hardware) SetProviderMetadata(provider Provider, metadata map[string]interface{}) {
	// Initialize ProviderMetadata map if nil
	if hardware.ProviderMetadata == nil {
		hardware.ProviderMetadata = map[Provider]ProviderMetadataRaw{}
	}

	// Set provider metadata
	hardware.ProviderMetadata[provider] = metadata
}

func NewHardwareFromBuildOut(hardwareBuildOut HardwareBuildOut, status HardwareStatus) Hardware {
	locationOrdinal := hardwareBuildOut.OrdinalPath[len(hardwareBuildOut.OrdinalPath)-1]

	return Hardware{
		ID:             hardwareBuildOut.ID,
		Parent:         hardwareBuildOut.ParentID,
		Type:           hardwareBuildOut.DeviceType.HardwareType,
		DeviceTypeSlug: hardwareBuildOut.DeviceType.Slug,
		Vendor:         hardwareBuildOut.DeviceType.Manufacturer,
		Model:          hardwareBuildOut.DeviceType.Model,

		LocationOrdinal: &locationOrdinal,

		Status: status,
	}
}

// HardwareStatus is the current state of the hardware
// Using a status allows for the hardware to be tracked through its lifecycle
// and allows for historical tracking of the hardware even if it is replaced or removed
type HardwareStatus string

// SchemaVersion is the version of the inventory schema
type SchemaVersion string

// Provider is the name of the external inventory provider
type Provider string

const (
	// Define constants for lifecyle states
	HardwareStatusEmpty          = HardwareStatus("empty")
	HardwareStatusStaged         = HardwareStatus("staged")
	HardwareStatusProvisioned    = HardwareStatus("provisioned")
	HardwareStatusDecommissioned = HardwareStatus("decommissioned")
	// Schema and proivider names are constant
	SchemaVersionV1Alpha1 = SchemaVersion("v1alpha1")
	CSMProvider           = Provider("csm")
)

// ProviderMetadataRaw stores the metadata from a provider in a generic map.
type ProviderMetadataRaw map[string]interface{}

type LocationToken struct {
	HardwareType hardwaretypes.HardwareType
	Ordinal      int
}

func (lt *LocationToken) String() string {
	return fmt.Sprintf("%s:%d", lt.HardwareType, lt.Ordinal)
}

type LocationPath []LocationToken

// String returns a string representation of the location path
func (lp LocationPath) String() string {
	tokens := []string{}

	for _, token := range lp {
		tokens = append(tokens, token.String())
	}

	return strings.Join(tokens, "->")
}

// GetHardwareTypePath returns the hardware type path of the location path
func (lp LocationPath) GetHardwareTypePath() hardwaretypes.HardwareTypePath {
	result := hardwaretypes.HardwareTypePath{}
	for _, token := range lp {
		result = append(result, token.HardwareType)
	}

	return result
}

// GetUUID returns the UUID of the location path
func (lp LocationPath) GetUUID(ds Datastore) (uuid.UUID, error) {
	hw, err := ds.GetAtLocation(lp)
	if err == nil {
		// Hardware found
		return hw.ID, nil
	} else if errors.Is(err, ErrHardwareNotFound) {
		// Hardware not found
		return uuid.Nil, ErrHardwareNotFound
	} else {
		// Oops something happened
		return uuid.Nil, err
	}
}

// GetOrdinalPath returns the ordinal of the location path
func (lp LocationPath) GetOrdinalPath() []int {
	result := []int{}
	for _, token := range lp {
		result = append(result, token.Ordinal)
	}

	return result
}

// Exists returns true if the hardware exists in the datastore
func (lp LocationPath) Exists(ds Datastore) (bool, error) {
	_, err := ds.GetAtLocation(lp)
	if err == nil {
		// Hardware found
		return true, nil
	} else if errors.Is(err, ErrHardwareNotFound) {
		// Hardware not found
		return false, nil
	} else {
		// Oops something happened
		return false, err
	}
}
