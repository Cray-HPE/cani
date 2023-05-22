package csm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/rs/zerolog/log"
)

// Reconcile CANI's inventory state with the external inventory state and apply required changes
func (csm *CSM) Reconcile(inv inventory.Inventory, slsHw []sls_client.Hardware) ([]sls_client.Hardware, error) {
	// compare and reconcile the two inventories
	var reconciled []sls_client.Hardware

	// logic to reconcile the two inventories
	for _, hw := range slsHw {
		log.Info().Msgf("Reconciling %s", hw.Xname)
		reconciled = append(reconciled, hw)
	}

	// return the reconciled inventory
	return reconciled, nil
}
