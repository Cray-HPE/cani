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
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// rackAttrs carries the per-rack attributes applied to each new rack.
type rackAttrs struct {
	statusArg  string
	serialArg  string
	locationID uuid.UUID
	tags       []string
	provMeta   map[string]string
}

// addAnyRack adds rack(s) using the resolved rack type.
func addAnyRack(cmd *cli.Command, args []string, rack *devicetypes.CaniRackType, qty int) error {
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	locationArg, _ := cmd.Flags().GetString("location")

	names, err := resolveNamesFromFlags(cmd, qty)
	if err != nil {
		return err
	}

	inventory, err := loadInventoryForAdd(cmd, args)
	if err != nil {
		return err
	}

	statusArg, err = normalizeStatus(statusArg, inventory)
	if err != nil {
		return err
	}

	tags, _ := cmd.Flags().GetStringArray("tag")
	attrs := rackAttrs{
		statusArg:  statusArg,
		serialArg:  serialArg,
		locationID: resolveLocation(inventory, locationArg),
		tags:       tags,
		provMeta:   collectProviderMetadata(cmd),
	}

	if err := addRacks(inventory, rack, names, attrs, qty); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d rack(s) added", qty)
	return nil
}

// addRacks builds and adds qty racks to the inventory.
func addRacks(inventory *devicetypes.Inventory, rack *devicetypes.CaniRackType, names []string, attrs rackAttrs, qty int) error {
	for i := range qty {
		r := *rack
		r.ID = uuid.New()
		r.Location = attrs.locationID
		setRackName(&r, names, i)
		if attrs.statusArg != "" {
			r.Status = attrs.statusArg
		}
		if attrs.serialArg != "" {
			r.Serial = attrs.serialArg
		}
		applyTagsToRack(&r, attrs.tags)
		applyProviderMetadataToRack(&r, attrs.provMeta)

		// Let registered providers apply post-add logic.
		if err := runRackPostAddHooks(&r, inventory); err != nil {
			return fmt.Errorf("provider hook failed: %w", err)
		}

		if err := inventory.AddRack(&r); err != nil {
			return fmt.Errorf("failed to add rack: %w", err)
		}
		inventory.AssignRacksToLocation(attrs.locationID)

		log.Printf("Added rack %s (%s)", r.ID, r.Name)
	}
	return nil
}

// setRackName assigns the rack name from the resolved names, falling back to
// the model when no name was supplied.
func setRackName(r *devicetypes.CaniRackType, names []string, i int) {
	if names != nil {
		r.Name = names[i]
	} else if r.Name == "" && r.Model != "" {
		r.Name = r.Model
	}
}
