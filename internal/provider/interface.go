package provider

import (
	"context"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
)

var ErrDataValidationFailure = fmt.Errorf("data validation failure")

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal(ctx context.Context) error

	// Validate the representation of the inventory data into the destination inventory system
	// is consistent. The default set of checks will verify all currently provided data is valid.
	// If enableRequiredDataChecks is set to true, additional checks focusing on missing data will be ran.
	ValidateInternal(ctx context.Context, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]HardwareValidationResult, error)

	// Import external inventory data into CANI's inventory format
	Import(ctx context.Context, datastore inventory.Datastore) error

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile(ctx context.Context, datastore inventory.Datastore) error

	// Build metadata, and add ito the hardware object
	// This function could return the data to put into object
	BuildHardwareMetadata(hw *inventory.Hardware, rawProperties map[string]interface{}) error
}

type HardwareValidationResult struct {
	Hardware inventory.Hardware
	Errors   []string
}
