package inventory

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Set adds a new Hardware object to the database
func (db *Database) Set(key uuid.UUID, value Hardware) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := (*db.Inventory)[key]; exists {
		return errors.New(fmt.Sprintf("%s already exists.", key.String()))
	}

	(*db.Inventory)[key] = value
	logTransaction("ADD", key.String(), value, nil)
	return db.writeDb()
}
