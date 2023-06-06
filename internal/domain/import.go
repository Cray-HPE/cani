package domain

import (
	"context"
)

func (d *Domain) Import(ctx context.Context) error {
	return d.externalInventoryProvider.Import(ctx, d.datastore)
}
