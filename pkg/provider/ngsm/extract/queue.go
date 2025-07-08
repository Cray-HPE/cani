/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
package extract

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Queues is a map of Queues for each bom file
type Queues map[string]*Queue

// Queue represents a Queue of racks and devices to be created in netbox
// gathered from an expert bom
type Queue struct {
	ExistingDevices map[string]devicetypes.DeviceType
	RacksToCreate   map[string]*Row
	DevicesToCreate map[string]*Row
	Bom             string
}

// newQueues creates a new Queue for a bom file
// it needs the existing device types, racks, and devices from netbox so it can
// run idempotently
func newQueues(bom string) (q *Queue, err error) {
	q = &Queue{
		ExistingDevices: make(map[string]devicetypes.DeviceType),
		RacksToCreate:   make(map[string]*Row),
		DevicesToCreate: make(map[string]*Row),
		Bom:             bom,
	}

	return q, nil
}

// func (q *Queue) addSystem(hw *devicetypes.DeviceType) (err error) {
// 	q.systems[hw.Name] = *hw
// 	return nil
// }

// // addRack adds a rack to the Queue if it is not already an existing rack
// func (q *Queue) addRack(row *Row) (err error) {
// 	// row.HardwareType = devicetypes.Rack //TODO: make a function to determeine the hardware type
// 	_, ok := q.existingRacks[row.NetboxName]
// 	if !ok {
// 		q.racksToCreate[row.NetboxName] = *row
// 	}

// 	return nil
// }

// addDerive adds a device to the Queue if it is not already an existing device
func (q *Queue) addDevice(row *Row) (err error) {
	_, ok := q.ExistingDevices[row.NetboxName]
	if !ok {
		if row.IsRack() {
			q.RacksToCreate[row.NetboxName] = row
		} else {
			q.DevicesToCreate[row.NetboxName] = row
		}
	}
	return nil
}

// AddRow adds the row to the Queue if it is a valid row for netbox
// If the row Quantity cell is more than 1, it will create as many duplicate
// rows as needed for the quantity, each with their own unique name
func (q *Queue) QueuesRow(row *Row) (err error) {
	if debug {
		log.Printf("queuing %+v:%v", filepath.Base(q.Bom), row.Row)
	}
	switch {
	case row.Quantity == 1:
		row.SetNetboxName(fmt.Sprintf("%s-row-%d-%d/%d", filepath.Base(q.Bom), row.Row, 1, row.Quantity))
		err = q.QueueRow(row)
		if err != nil {
			return err
		}
	case row.Quantity > 1:
		for i := 0; i < row.Quantity; i++ {
			// make a duplicate row for each quantity
			newRow := row.NewRowFromRow()
			newRow.SetSource(filepath.Base(q.Bom))
			// set the name to be unique with the row number and quantity
			newRow.SetNetboxName(fmt.Sprintf("%s-row-%d-%d/%d", filepath.Base(q.Bom), row.Row, i+1, row.Quantity))
			// add each row to the Queue
			err = q.QueueRow(&newRow)
			if err != nil {
				return err
			}
		}

	default:
	}

	return nil
}

// QueueRow adds a row to the Queue if it is a valid row for netbox and will
// either add it as a rack or a device
func (q *Queue) QueueRow(row *Row) (err error) {
	slug := core.Slugify(row.ProductDescription)
	row.SetSource(filepath.Base(q.Bom))

	switch {

	// case row.IsRack():
	// 	if debug {
	// 		log.Printf("Validating %+v:%v -> Queuing a rack: %v", filepath.Base(q.bom), row.row, row.ProductNumber)
	// 	}
	// 	row.SetNetboxDeviceTypeSlug(slug)
	// 	err = q.addDevice(row)

	case row.IsDevice():
		if debug {
			log.Printf("queuing %v:%v -> %v", filepath.Base(q.Bom), row.Row, row.ProductNumber)
		}
		row.SetNetboxDeviceTypeSlug(slug)
		err = q.addDevice(row)

	default:
		// log.Printf("%+v:%v Will not Queue an unknown device type: '%v'", filepath.Base(q.bom), row.row, row.ProductDescription)
	}

	if err != nil {
		return err
	}
	return nil
}

// HasRack checks if the Queue contains any rack hardware
// Returns the first rack found and true if found, otherwise empty row and false
func (q *Queue) HasRack() (map[string]*Row, bool) {
	racks := make(map[string]*Row, 0)
	for _, row := range q.DevicesToCreate {
		re1 := regexp.MustCompile(`\d+U`)
		re2 := regexp.MustCompile(`[Rr]ack`)
		re3 := regexp.MustCompile(`(?i)[Kk]it|[Ss]ervice`)
		// if a row has both 'rack', a U count, and does not contain 'kit' or 'service', it is likely a rack
		if re1.MatchString(row.ProductDescription) && re2.MatchString(row.ProductDescription) && !re3.MatchString(row.ProductDescription) {
			racks[row.NetboxName] = row
		}
	}
	return racks, len(racks) > 0
}
