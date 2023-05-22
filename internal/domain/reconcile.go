package domain

func (d *Domain) Reconcile() error {
	return d.externalInventoryProvider.Reconcile(d.datastore)
}
