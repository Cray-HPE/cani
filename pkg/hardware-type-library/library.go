package hardware_type_library

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"path"

	"gopkg.in/yaml.v3"
)

//go:embed hardware-types/*.yaml
var defaultHardwareTypesFS embed.FS

//go:embed hardware-types/schema/*.json
var hardwareTypeSchemas embed.FS

var ErrDeviceTypeAlreadyExists = fmt.Errorf("device type already exists")

type Library struct {
	DeviceTypes map[string]DeviceType
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
	library := &Library{
		DeviceTypes: map[string]DeviceType{},
	}

	// Load the embedded hardware type embedded files
	basePath := "hardware-types"
	files, err := defaultHardwareTypesFS.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	// Parse hardware type files
	for _, file := range files {
		filePath := path.Join(basePath, file.Name())
		fmt.Println("Parsing file:", filePath)

		fileRaw, err := defaultHardwareTypesFS.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var fileDeviceTypes []DeviceType
		if err := unmarshalMultiple(fileRaw, &fileDeviceTypes); err != nil {
			return nil, err
		}

		for _, deviceType := range fileDeviceTypes {
			fmt.Println("  Registering device type:", deviceType.Slug)
			if err := library.RegisterDeviceType(deviceType); err != nil {
				return nil, errors.Join(
					fmt.Errorf("failed to register device type '%s'", deviceType.Slug),
					err,
				)
			}
		}
	}

	return library, nil
}

func NewLibraryFromPath(path string) (*Library, error) {
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

	return result
}

// func GetDeviceType(name string) DeviceType {

// }

// func GetDeviceTypeBuildOut(name string) []DeviceBay {

// }
