/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"log"
	"os"

	"github.com/Cray-HPE/cani/internal/core"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ActiveProvider string                    `yaml:"active_provider"`
	Providers      map[string]map[string]any `yaml:"providers"`
	Path           string                    `yaml:"-"`         // Path to the config file, set when loaded
	Datastore      string                    `yaml:"datastore"` // Path to the datastore, set when loaded
}

var Cfg *Config // the singleton

// Load reads or creates the config at path.
func Load(path string) error {
	Cfg = &Config{Providers: map[string]map[string]any{}}
	Cfg.Path = path
	Cfg.Datastore = core.DsPath
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true) // Fail on unknown fields

	err = decoder.Decode(Cfg)
	if err != nil {
		return err
	}

	log.Printf("Loaded config: %s", Cfg.Path)
	return nil
}

// Save writes Cfg back to path.
func Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(Cfg)
}
