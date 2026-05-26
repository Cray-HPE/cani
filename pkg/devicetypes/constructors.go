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
	"math/rand"

	"github.com/google/uuid"
)

// NewLocation creates a new empty CaniLocationType.
func NewLocation() *CaniLocationType {
	return &CaniLocationType{}
}

// NewDefaultLocation creates a location with a generated UUID and
// sensible defaults, suitable for use as the root location.
func NewDefaultLocation() *CaniLocationType {
	return &CaniLocationType{
		ID:           uuid.New(),
		Name:         "default-cani",
		LocationType: "site",
		ObjectMeta:   ObjectMeta{Status: string(StatusActive)},
	}
}

// NewLocationFromSlug creates a CaniLocationType from a registered LocationTypeDefinition.
func NewLocationFromSlug(slug string) (*CaniLocationType, error) {
	lt, ok := GetLocationTypeBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("location type slug %q not found in library", slug)
	}
	loc := &CaniLocationType{
		ID:           uuid.New(),
		Name:         lt.Name,
		Slug:         lt.Slug,
		LocationType: lt.Slug,
		ObjectMeta:   ObjectMeta{Status: string(StatusActive)},
		Nestable:     lt.Nestable,
		ContentTypes: lt.ContentTypes,
		Description:  lt.Description,
		Source:       lt.Source,
	}
	return loc, nil
}

// generateCaniName returns a random placeholder name for new devices.
func generateCaniName() string {
	return fmt.Sprintf("cani-device-%d", rand.Intn(1000000))
}

// NewDeviceFromSlug creates a CaniDeviceType inventory instance from a registry slug.
func NewDeviceFromSlug(slug string) (*CaniDeviceType, error) {
	dt, ok := GetBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("device type slug %q not found in library", slug)
	}
	device := dt // shallow copy
	device.ID = uuid.New()
	device.Name = generateCaniName()
	device.Status = string(StatusStaged)
	return &device, nil
}

// NewDeviceFromPartNumber creates a CaniDeviceType from a part number.
func NewDeviceFromPartNumber(partNumber string) (*CaniDeviceType, error) {
	dt, ok := GetByPartNumber(partNumber)
	if !ok {
		return nil, fmt.Errorf("device type part number %q not found in library", partNumber)
	}
	device := dt
	device.ID = uuid.New()
	device.Name = generateCaniName()
	device.Status = string(StatusStaged)
	return &device, nil
}

// NewRackFromSlug creates a CaniRackType inventory instance from a registry slug.
func NewRackFromSlug(slug string) (*CaniRackType, error) {
	rt, ok := GetRackTypeBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("rack type slug %q not found in library", slug)
	}
	rack := rt
	rack.ID = uuid.New()
	rack.Status = string(StatusActive)
	rack.OccupiedSlots = make(map[int]map[string]uuid.UUID)
	return &rack, nil
}

// NewRackFromPartNumber creates a CaniRackType from a part number.
func NewRackFromPartNumber(partNumber string) (*CaniRackType, error) {
	rt, ok := GetRackTypeByPartNumber(partNumber)
	if !ok {
		return nil, fmt.Errorf("rack type part number %q not found in library", partNumber)
	}
	rack := rt
	rack.ID = uuid.New()
	rack.Status = string(StatusActive)
	rack.OccupiedSlots = make(map[int]map[string]uuid.UUID)
	return &rack, nil
}

// NewModuleFromSlug creates a CaniModuleType inventory instance from a registry slug.
func NewModuleFromSlug(slug string) (*CaniModuleType, error) {
	mt, ok := GetModuleBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("module type slug %q not found in library", slug)
	}
	mod := mt
	mod.ID = uuid.New()
	mod.Name = mt.Model
	mod.Status = string(StatusActive)
	return &mod, nil
}

// NewModuleFromPartNumber creates a CaniModuleType from a part number.
func NewModuleFromPartNumber(partNumber string) (*CaniModuleType, error) {
	mt, ok := GetModuleTypeByPartNumber(partNumber)
	if !ok {
		return nil, fmt.Errorf("module type part number %q not found in library", partNumber)
	}
	mod := mt
	mod.ID = uuid.New()
	mod.Name = mt.Model
	mod.Status = string(StatusActive)
	return &mod, nil
}

// NewCableFromSlug creates a CaniCableType instance from a registry slug.
func NewCableFromSlug(slug string) (*CaniCableType, error) {
	ct, ok := GetCableTypeBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("cable type slug %q not found in library", slug)
	}
	cable := ct
	cable.ID = uuid.New()
	cable.Status = string(StatusConnected)
	return &cable, nil
}

// NewCableFromPartNumber creates a CaniCableType from a part number.
func NewCableFromPartNumber(partNumber string) (*CaniCableType, error) {
	ct, ok := GetCableTypeByPartNumber(partNumber)
	if !ok {
		return nil, fmt.Errorf("cable type part number %q not found in library", partNumber)
	}
	cable := ct
	cable.ID = uuid.New()
	cable.Status = string(StatusConnected)
	return &cable, nil
}

// NewFruFromSlug creates a CaniFruType instance from a registry slug.
func NewFruFromSlug(slug string) (*CaniFruType, error) {
	ft, ok := GetFruTypeBySlug(slug)
	if !ok {
		return nil, fmt.Errorf("FRU type slug %q not found in library", slug)
	}
	fru := ft
	fru.ID = uuid.New()
	fru.Status = string(StatusActive)
	return &fru, nil
}

// NewFruFromPartNumber creates a CaniFruType from a part number.
func NewFruFromPartNumber(partNumber string) (*CaniFruType, error) {
	ft, ok := GetFruTypeByPartNumber(partNumber)
	if !ok {
		return nil, fmt.Errorf("FRU type part number %q not found in library", partNumber)
	}
	fru := ft
	fru.ID = uuid.New()
	fru.Status = string(StatusActive)
	return &fru, nil
}
