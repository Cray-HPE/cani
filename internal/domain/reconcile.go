package domain

import (
	"errors"
	"fmt"
)

func (d *Domain) Reconcile() error {
	data, err := d.datastore.List()
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to retrieve inventory data"),
			err,
		)
	}

	return d.externalInventoryProvider.Reconcile(data)
}
