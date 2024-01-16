package ngsm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (ngsm *Ngsm) ListCabinetMetadataColumns() (columns []string) {
	log.Warn().Msgf("ListCabinetMetadataColumns not yet implemented")
	return columns
}

func (ngsm *Ngsm) ListCabinetMetadataRow(inventory.Hardware) (values []string, err error) {
	log.Warn().Msgf("ListCabinetMetadataRow not yet implemented")
	return values, err
}

func (ngsm *Ngsm) NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) error {
	log.Warn().Msgf("NewHardwareMetadata not yet implemented")
	return nil
}
