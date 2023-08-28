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
package hardwaretypes

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

//go:embed hardware-types/*.yaml
var defaultHardwareTypesFS embed.FS

//go:embed hardware-types/schema/*.json
var hardwareTypeSchemas embed.FS

var ErrDeviceTypeAlreadyExists = fmt.Errorf("device type already exists")

type Library struct {
	DeviceTypes map[string]DeviceType // TODO make private?
}

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func unmarshalMultiple(in []byte, out *[]DeviceType) error {
	r := bytes.NewReader(in)
	decoder := yaml.NewDecoder(r)

	var documentNumber int
	for {
		var deviceType DeviceType
		if err := decoder.Decode(&deviceType); err != nil {
			// Break out of loop when more yaml documents to process
			if err != io.EOF {
				return fmt.Errorf("failed to parse document %d, error %w", documentNumber, err)
			}

			break
		}

		*out = append(*out, deviceType)
		documentNumber++
	}

	return nil
}

func NewEmbeddedLibrary(customDir string) (*Library, error) {
	library := &Library{
		DeviceTypes: map[string]DeviceType{},
	}

	// Load the embedded hardware type embedded files
	basePath := "hardware-types"
	log.Debug().Msgf("Looking for built-in hardware-types")
	defaultFiles, err := defaultHardwareTypesFS.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Looking for custom hardware-types in %s", customDir)
	// append user-defined hardware-type files to the default embedded ones
	customFiles, err := os.ReadDir(customDir)
	if err != nil {
		// it is ok if no custom files exist
		log.Debug().Msgf("No custom hardware-types defined in %s", customDir)
	}

	// Parse hardware type files
	for _, file := range defaultFiles {
		filePath := path.Join(basePath, file.Name())
		log.Debug().Msgf("Parsing built-in hardware-type: %s", filePath)

		fileRaw, err := defaultHardwareTypesFS.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var fileDeviceTypes []DeviceType
		if err := unmarshalMultiple(fileRaw, &fileDeviceTypes); err != nil {
			return nil, err
		}

		for _, deviceType := range fileDeviceTypes {
			log.Debug().Msgf("Registering device type: %s", deviceType.Slug)
			if err := library.RegisterDeviceType(deviceType); err != nil {
				return nil, errors.Join(
					fmt.Errorf("failed to register device type '%s'", deviceType.Slug),
					err,
				)
			}
		}
	}

	// if there are user-defined files, read them
	if len(customFiles) != 0 {
		for _, file := range customFiles {
			filePath := filepath.Join(customDir, file.Name())
			log.Debug().Msgf("Parsing custom hardware-type: %v", filePath)

			fileRaw, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}

			var fileDeviceTypes []DeviceType
			if err := unmarshalMultiple(fileRaw, &fileDeviceTypes); err != nil {
				return nil, err
			}

			for _, deviceType := range fileDeviceTypes {
				log.Debug().Msgf("Registering device type: %s", deviceType.Slug)
				if err := library.RegisterDeviceType(deviceType); err != nil {
					return nil, errors.Join(
						fmt.Errorf("failed to register device type '%s'", deviceType.Slug),
						err,
					)
				}
			}
		}
	}

	return library, nil
}

func NewLibraryFromPath(path string) (*Library, error) {
	panic("TODO")
	return &Library{}, nil
}

func (l *Library) RegisterDeviceType(deviceType DeviceType) error {
	if _, exists := l.DeviceTypes[deviceType.Slug]; exists {
		return ErrDeviceTypeAlreadyExists
	}

	l.DeviceTypes[deviceType.Slug] = deviceType
	return nil
}

func (l *Library) GetDeviceTypesByHardwareType(hardwareType HardwareType) []DeviceType {
	var result []DeviceType
	for _, deviceType := range l.DeviceTypes {
		if deviceType.HardwareType == hardwareType {
			result = append(result, deviceType)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Slug) < strings.ToLower(result[j].Slug)
	})

	return result
}

func (l *Library) GetDeviceType(slug string) (DeviceType, error) {
	deviceType, ok := l.DeviceTypes[slug]
	if !ok {
		return DeviceType{}, fmt.Errorf("no device type exists with slug (%s)", slug)
	}

	return deviceType, nil
}

// func GetDeviceTypeBuildOut(name string) []DeviceBay {

// }
