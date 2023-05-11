package hardware_type_library

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"path"

	"gopkg.in/yaml.v3"
)

//go:embed hardware-types/*.yaml
var defaultHardwareTypesFS embed.FS

//go:embed hardware-types/schema/*.json
var hardwareTypeSchemas embed.FS

type Library struct {
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

func NewEmbeddedLibrary() (*Library, error) {
	// Load the embedded hardware type embedded files
	basePath := "hardware-types"
	files, err := defaultHardwareTypesFS.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	// Parse hardware type files
	for _, file := range files {
		filePath := path.Join(basePath, file.Name())
		fmt.Println(filePath)

		fileRaw, err := defaultHardwareTypesFS.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var fileDeviceTypes []DeviceType
		if err := unmarshalMultiple(fileRaw, &fileDeviceTypes); err != nil {
			return nil, err
		}
	}

	return &Library{}, nil
}

func NewLibraryFromPath(path string) (*Library, error) {
	return &Library{}, nil
}

// func GetDeviceTypesByHardwareType(hardwareClass HardwareType) []DeviceType {

// }

// func GetDeviceType(name string) DeviceType {

// }

// func GetDeviceTypeBuildOut(name string) []DeviceBay {

// }
