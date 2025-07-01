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
package devicetypes

import (
	"embed"
	stdlog "log"
	"os"
	"path"
)

//go:embed device-types/HPE/*.y*ml
var deviceTypes embed.FS

var (
	allDeviceTypes = make(map[string]DeviceType, 0)
)

var log = stdlog.New(os.Stdout, "[devicetypes] ", stdlog.LstdFlags)

func init() {
	log.Printf("Looking for built-in device-types")
	defaultFiles, err := deviceTypes.ReadDir("device-types/HPE")
	if err != nil {
		log.Printf("Failed to read device-types directory: %v", err)
		os.Exit(1)
	}

	// load all the device types from device-types dir
	// Parse device type files
	for _, file := range defaultFiles {
		filePath := path.Join("device-types/HPE", file.Name())
		// log.Printf("Parsing built-in device-type: %s", filePath)

		fileRaw, err := deviceTypes.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read device-type file %s: %v", filePath, err)
			continue
		}

		var fileDeviceTypes []DeviceType
		if err := unmarshalMultiple(fileRaw, &fileDeviceTypes); err != nil {
			log.Printf("Failed to parse device-type file %s: %v", filePath, err)
			continue
		}

		for _, deviceType := range fileDeviceTypes {
			allDeviceTypes[deviceType.Slug] = deviceType
		}
	}
}
