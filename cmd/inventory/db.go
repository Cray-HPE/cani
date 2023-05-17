package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/rs/zerolog/log"
)

// Database is a simple key-value store for Hardware objects
type Database struct {
	mu        sync.RWMutex
	Inventory *Inventory
	dataFile  string
}

var (
	DbPath   = taxonomy.DsFile
	instance *Database
	once     sync.Once
)

func GetInstance(filename string) *Database {
	once.Do(func() {
		instance = &Database{
			dataFile:  filename,
			Inventory: NewInventory(),
		}
		instance.loadDb()
	})
	return instance
}

// InitDb creates a default database file if one does not exist
func InitDb(path string) (db *Database, err error) {
	// Write a default config file if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info().Msg(fmt.Sprintf("%s does not exist, creating default database", path))

		// Create the directory if it doesn't exist
		dbDir := filepath.Dir(path)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			err = os.Mkdir(dbDir, 0755)
			if err != nil {
				return &Database{}, errors.New(fmt.Sprintf("Error creating database directory: %s", err))
			}
		}

		// Create a config with default values since one does not exist
		db := &Database{
			Inventory: NewInventory(),
			dataFile:  path,
		}

		// Create the config file
		WriteDb(path, db)
	}

	tlogdir := filepath.Dir(path)
	if _, err := os.Stat(tlogdir); os.IsNotExist(err) {
		err = os.Mkdir(tlogdir, 0755)
		if err != nil {
			return &Database{}, errors.New(fmt.Sprintf("Error creating transaction log directory: %s", err))
		}
	}

	// create the transaction log file
	fname := filepath.Base(path)
	fname = strings.TrimSuffix(fname, filepath.Ext(fname))
	fname = fmt.Sprintf("%s.log", fname)
	tlogfile := filepath.Join(tlogdir, fname)
	transactionLogFile, err = os.OpenFile(tlogfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &Database{}, errors.New(fmt.Sprintf("Error creating transaction log file: %s", err))
	}

	return db, nil
}

// LoadDb loads the database from a JSON file
func LoadDb(path string, db *Database) (*Database, error) {
	// Create the directory if it doesn't exist
	cfgDir := filepath.Dir(path)
	os.MkdirAll(cfgDir, os.ModePerm)

	file, err := os.Open(path)
	if err != nil {
		return db, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return db, err
	}

	err = json.Unmarshal(data, &db)
	if err != nil {
		return db, err
	}
	return db, nil
}

// WriteDb saves the database to a file
func WriteDb(path string, db *Database) error {
	// convert the cfg struct to a YAML-formatted byte slice,
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}

	// write the byte slice to a file
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// loadDb loads the database from a JSON file
func (db *Database) loadDb() (*Database, error) {
	file, err := os.Open(db.dataFile)
	if err != nil {
		return db, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return db, err
	}

	err = json.Unmarshal(bytes, &db.Inventory)
	if err != nil {
		return db, err
	}

	return db, nil
}

// writeDb saves the database to a file
func (db *Database) writeDb() error {
	bytes, err := json.Marshal(db.Inventory)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(db.dataFile, bytes, 0644)
}
