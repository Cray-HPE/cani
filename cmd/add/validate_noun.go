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
package add

import (
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// Noun identifies which inventory noun a subcommand operates on.
type Noun string

const (
	NounLocation Noun = "location"
	NounRack     Noun = "rack"
	NounDevice   Noun = "device"
	NounModule   Noun = "module"
	NounCable    Noun = "cable"
)

// nounTypeMap maps noun names to the hardware types they accept.
var nounTypeMap = map[Noun][]devicetypes.Type{
	NounRack: {
		devicetypes.TypeRack,
		devicetypes.TypeCabinet,
	},
	NounDevice: {
		devicetypes.TypeBlade,
		devicetypes.TypeNode,
		devicetypes.TypeNodeCard,
		devicetypes.TypeChassis,
		devicetypes.TypeSwitch,
		devicetypes.TypeMgmtSwitch,
		devicetypes.TypeHsnSwitch,
		devicetypes.TypeCabinetPDU,
		devicetypes.TypeCDU,
	},
	NounModule: {
		devicetypes.TypeModule,
		devicetypes.TypeNIC,
		devicetypes.TypeGPU,
		devicetypes.TypeCPU,
		devicetypes.TypeMemory,
		devicetypes.TypePowerSupply,
	},
	NounCable: {
		devicetypes.TypeCable,
	},
}

// validSlugOrPartNumber validates the positional arg for an add subcommand.
// It checks device, rack, module, and cable registries as appropriate.
func validSlugOrPartNumber(noun Noun) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(os.Stderr)

		if cmd.Flags().Changed("list-supported-types") {
			return listTypesForNoun(cmd, noun)
		}

		if len(args) == 0 {
			return fmt.Errorf("slug or part number required")
		}

		slug := args[0]
		if _, err := lookupBySlugOrPart(noun, slug); err != nil {
			return err
		}
		return nil
	}
}

// lookupResult is a discriminated union returned by lookupBySlugOrPart.
type lookupResult struct {
	Noun     Noun
	Device   *devicetypes.CaniDeviceType
	Rack     *devicetypes.CaniRackType
	Module   *devicetypes.CaniModuleType
	Cable    *devicetypes.CaniCableType
	Location *devicetypes.CaniLocationType
}

// lookupBySlugOrPart finds a hardware type by slug, falling back to part number.
func lookupBySlugOrPart(noun Noun, key string) (*lookupResult, error) {
	switch noun {
	case NounDevice:
		if d, err := devicetypes.NewDeviceFromSlug(key); err == nil {
			return &lookupResult{Noun: noun, Device: d}, nil
		}
		if d, err := devicetypes.NewDeviceFromPartNumber(key); err == nil {
			return &lookupResult{Noun: noun, Device: d}, nil
		}
		return nil, fmt.Errorf("unknown device slug or part number: %s", key)

	case NounRack:
		if r, err := devicetypes.NewRackFromSlug(key); err == nil {
			return &lookupResult{Noun: noun, Rack: r}, nil
		}
		if r, err := devicetypes.NewRackFromPartNumber(key); err == nil {
			return &lookupResult{Noun: noun, Rack: r}, nil
		}
		return nil, fmt.Errorf("unknown rack slug or part number: %s", key)

	case NounModule:
		if m, err := devicetypes.NewModuleFromSlug(key); err == nil {
			return &lookupResult{Noun: noun, Module: m}, nil
		}
		if m, err := devicetypes.NewModuleFromPartNumber(key); err == nil {
			return &lookupResult{Noun: noun, Module: m}, nil
		}
		return nil, fmt.Errorf("unknown module slug or part number: %s", key)

	case NounCable:
		if c, err := devicetypes.NewCableFromSlug(key); err == nil {
			return &lookupResult{Noun: noun, Cable: c}, nil
		}
		if c, err := devicetypes.NewCableFromPartNumber(key); err == nil {
			return &lookupResult{Noun: noun, Cable: c}, nil
		}
		return nil, fmt.Errorf("unknown cable slug or part number: %s", key)

	case NounLocation:
		if l, err := devicetypes.NewLocationFromSlug(key); err == nil {
			return &lookupResult{Noun: noun, Location: l}, nil
		}
		// Fall back to untyped location (created via flags).
		return &lookupResult{Noun: noun}, nil

	default:
		return nil, fmt.Errorf("unknown noun: %s", noun)
	}
}

// listTypesForNoun prints supported slugs for a noun and exits.
func listTypesForNoun(cmd *cobra.Command, noun Noun) error {
	var entries []devicetypes.TypeEntry

	switch noun {
	case NounDevice:
		types := nounTypeMap[NounDevice]
		devices := devicetypes.ListCaniDeviceTypes(types...)
		for _, d := range devices {
			entries = append(entries, devicetypes.TypeEntry{
				Name: d.Model, Slug: d.Slug,
				PartNumber: d.PartNumber, Category: string(d.Type),
				Source: d.Source,
			})
		}
	case NounRack:
		for _, r := range devicetypes.AllRackTypes() {
			entries = append(entries, devicetypes.TypeEntry{
				Name: r.Model, Slug: r.Slug,
				PartNumber: r.PartNumber, Category: string(r.Type),
				Source: r.Source,
			})
		}
	case NounModule:
		for _, m := range devicetypes.AllModules() {
			entries = append(entries, devicetypes.TypeEntry{
				Name: m.Model, Slug: m.Slug,
				PartNumber: m.PartNumber, Category: string(m.Type),
				Source: m.Source,
			})
		}
	case NounCable:
		for _, c := range devicetypes.AllCables() {
			entries = append(entries, devicetypes.TypeEntry{
				Name: c.Model, Slug: c.Slug,
				PartNumber: c.PartNumber, Category: string(c.Type),
				Source: c.Source,
			})
		}
	case NounLocation:
		return listLocationTypes(cmd)
	}

	cmd.SetOut(os.Stderr)
	printTypeTable(cmd, entries)
	os.Exit(0)
	return nil
}
