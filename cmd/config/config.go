/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config is the top-level config struct that is written/read to/from a file
type Config struct {
	Session *Session `yaml:"session"`
}

// InitConfig creates a default config file if one does not exist
func InitConfig(cfg string) (err error) {
	// Create the directory if it doesn't exist
	configDir := filepath.Dir(cfg)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating config directory: %s", err)
		}
	}
	tl := filepath.Join(configDir, taxonomy.LogFile)
	customDir := fmt.Sprintf("%s/%s", configDir, "hardware-types")
	// Write a default config file if it doesn't exist
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		log.Debug().Msg(fmt.Sprintf("%s does not exist, creating default config file", cfg))

		// Create a config with default values since one does not exist
		conf := &Config{
			Session: &Session{
				DomainOptions: &domain.DomainOpts{
					DatastorePath:          filepath.Join(configDir, taxonomy.DsFile),
					LogFilePath:            tl,
					CustomHardwareTypesDir: customDir,
				},
			},
		}

		// Create the config file
		WriteConfig(cfg, conf)
	}

	err = os.MkdirAll(customDir, 0755)
	if err != nil {
		return err
	}

	return nil
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (c *Config, err error) {
	// Create the directory if it doesn't exist
	cfgDir := filepath.Dir(path)
	os.MkdirAll(cfgDir, os.ModePerm)

	// Open the file, create it if it doesn't exist
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the file into a byte slice
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// unmarshal the byte slice into a struct
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// WriteConfig saves the configuration to a file
func WriteConfig(path string, cfg *Config) error {
	// convert the cfg struct to a YAML-formatted byte slice,
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// write the byte slice to a file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
