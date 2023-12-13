package hpengi

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpengi *Hpengi) ListCabinetMetadataColumns() (columns []string) {
	log.Warn().Msgf("ListCabinetMetadataColumns not yet implemented")
	return columns
}

func (hpengi *Hpengi) ListCabinetMetadataRow(inventory.Hardware) (values []string, err error) {
	log.Warn().Msgf("ListCabinetMetadataRow not yet implemented")
	return values, err
}

func (hpengi *Hpengi) NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) error {
	log.Warn().Msgf("NewHardwareMetadata not yet implemented")
	return nil
}
