package inventory

import (
	"errors"

	"github.com/google/uuid"
)

// Delete removes a Hardware object from the database
func (db *Database) Delete(keys []uuid.UUID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, u := range keys {
		if _, exists := (*db.Inventory)[u]; !exists {
			return errors.New("key not found")
		}
		delete(*db.Inventory, u)
		logTransaction("DELETE", u.String(), "", nil)
	}

	return db.writeDb()
}
