package ngsm

import (
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/remove"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Remove is called when the user runs `cani remove <device> <device-type-slug> <args>`
func (p *Ngsm) Remove(cmd *cobra.Command, args []string) (idsToRemove []uuid.UUID, err error) {
	return remove.Remove(cmd, args)
}
