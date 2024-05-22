package ngsm

import (
	"fmt"
	"path/filepath"

	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// Queue is a map of queues for each bom file
type Queue map[string]*queue

// queue represents a queue of racks and devices to be created in netbox
// gathered from an expert bom
type queue struct {
	existingRacks       map[string]netbox.Rack
	existingDevices     map[string]netbox.DeviceWithConfigContext
	existingDeviceTypes map[string]netbox.DeviceType
	deviceTypeIds       map[string]int32
	racksDetected       map[string]Row
	racksToCreate       map[string]Row
	rackSummary         map[int32]int
	devicesDetected     map[string]Row
	devicesToCreate     map[string]Row
	deviceSummary       map[int32]int
	bom                 string
}

// newQueue creates a new queue for a bom file
// it needs the existing device types, racks, and devices from netbox so it can
// run idempotently
func newQueue(bom string, existingDeviceTypes map[string]netbox.DeviceType, existingRacks map[string]netbox.Rack, existingDevices map[string]netbox.DeviceWithConfigContext) (q *queue, err error) {
	q = &queue{
		existingRacks:       existingRacks,
		existingDevices:     existingDevices,
		existingDeviceTypes: existingDeviceTypes,
		deviceTypeIds:       make(map[string]int32),
		racksDetected:       make(map[string]Row),
		racksToCreate:       make(map[string]Row),
		rackSummary:         make(map[int32]int),
		devicesDetected:     make(map[string]Row),
		devicesToCreate:     make(map[string]Row),
		deviceSummary:       make(map[int32]int),
		bom:                 bom,
	}

	// get the device type IDs
	for _, deviceType := range existingDeviceTypes {
		if deviceType.GetPartNumber() != "" {
			q.deviceTypeIds[deviceType.GetPartNumber()] = deviceType.GetId()
		}
	}

	return q, nil
}

// addRack adds a rack to the queue if it is not already an existing rack
func (q *queue) addRack(row *Row) (err error) {
	_, ok := q.existingRacks[row.netboxName]
	if !ok {
		log.Debug().Msgf("Netbox rack will be created: '%s'", row.netboxName)
		q.racksDetected[row.netboxName] = *row
		q.racksToCreate[row.netboxName] = *row
		q.rackSummary[row.netboxDeviceTypeID]++
	}
	return nil
}

// addDerive adds a device to the queue if it is not already an existing device
func (q *queue) addDevice(row *Row) (err error) {
	_, ok := q.existingDevices[row.netboxName]
	if !ok {
		log.Debug().Msgf("Netbox device will be created: '%s' (%d)", row.netboxName, row.netboxDeviceTypeID)
		q.devicesDetected[row.netboxName] = *row
		q.devicesToCreate[row.netboxName] = *row
		q.deviceSummary[row.netboxDeviceTypeID]++
	}
	return nil
}

// AddRow adds the row to the queue if it is a valid row for netbox
// If the row Quantity cell is more than 1, it will create as many duplicate
// rows as needed for the quantity, each with their own unique name
func (q *queue) AddRow(row *Row) (err error) {
	switch {
	case row.Quantity == 1:
		row.SetNetboxName(fmt.Sprintf("%s-row-%d-%d/%d", filepath.Base(q.bom), row.row, 1, row.Quantity))
		err = q.addRow(row)
		if err != nil {
			return err
		}
	case row.Quantity > 1:
		for i := 0; i < row.Quantity; i++ {
			// make a duplicate row for each quantity
			newRow := row.NewRowFromRow()
			newRow.SetSource(filepath.Base(q.bom))
			newRow.SetNetboxDeviceTypeID(q.deviceTypeIds[row.ProductNumber])
			// set the name to be unique with the row number and quantity
			newRow.SetNetboxName(fmt.Sprintf("%s-row-%d-%d/%d", filepath.Base(q.bom), row.row, i+1, row.Quantity))
			// add each row to the queue
			err = q.addRow(&newRow)
			if err != nil {
				return err
			}
		}

	default:
	}

	return nil
}

// addRow adds a row to the queue if it is a valid row for netbox and will
// either add it as a rack or a device
func (q *queue) addRow(row *Row) (err error) {
	row.SetSource(filepath.Base(q.bom))
	switch {

	case row.IsDevice(q.deviceTypeIds):
		log.Debug().Msgf("Detected a device: '%v' (%+v:%v)", row.ProductDescription, filepath.Base(q.bom), row.row)
		row.SetNetboxDeviceTypeID(q.deviceTypeIds[row.ProductNumber])
		err = q.addDevice(row)

	case row.IsRack():
		log.Debug().Msgf("Detected a rack: '%v' (%+v:%v)", row.ProductDescription, filepath.Base(q.bom), row.row)
		err = q.addRack(row)

	default:
		log.Trace().Msgf("Not a rack or device: '%v' (%+v:%v)", row.ProductDescription, filepath.Base(q.bom), row.row)
	}

	if err != nil {
		return err
	}
	return nil
}

// Sanitize attepts to sanitize the row's content so valid data is returned
// this will do things like normalize part numbers, clean up blankspace, etc.
func (row *Row) Sanitize() (parsed *Row, err error) {
	// sanitize the row's content
	err = row.sanitize()
	if err != nil {
		return nil, err
	}

	return row, nil
}
