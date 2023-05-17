package domain

func (d *Domain) Commit() error {
	return d.externalInventoryProvider.Reconcile()
}
