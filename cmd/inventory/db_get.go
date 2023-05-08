package inventory

import (
	"errors"

	"github.com/google/uuid"
)

// Get returns an Inventory object from the database
func (db *Database) Get(keys []uuid.UUID) (Inventory, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if len(keys) == 0 {
		logTransaction("GET", "", "", nil)
		return *db.Inventory, nil
	}

	results := make(Inventory)

	for _, u := range keys {
		if hardware, exists := (*db.Inventory)[u]; exists {
			results[u] = hardware
			logTransaction("GET", u.String(), hardware, nil)
		} else {
			return Inventory{}, errors.New("key not found")
		}
	}

	return results, nil
}
