package inventory

import (
	"errors"

	"github.com/google/uuid"
)

type Inventory struct {
	Hardware map[uuid.UUID]*Hardware
}

func (i *Inventory) Add(hardware *Hardware) error {
	// TODO Check to see if the UUID exists

	// Check to see if the UUID is empty
	if hardware.ID == uuid.Nil {
		return errors.New("hardware id is nil")
	}

	i.Hardware[hardware.ID] = hardware

	return nil
}

func (i *Inventory) Link(parent, child *Hardware) error {
	return nil
}
