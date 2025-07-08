package remove

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func Remove(cmd *cobra.Command, args []string) (idsToRemove []uuid.UUID, err error) {
	for _, arg := range args {
		// any pre-removal processing can be done here

		u, err := uuid.Parse(arg)
		if err != nil {
			return nil, err
		}
		idsToRemove = append(idsToRemove, u)
	}
	return idsToRemove, nil
}
