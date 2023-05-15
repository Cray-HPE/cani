package hardware_type_library

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"

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

// TODO needs a different name
type HardwareBuildOut struct {
	DeviceTypeString string
	DeviceType       DeviceType
	Path             []string
	Ordinal          int
	OrdinalPath      []int
	HardwareTypePath []HardwareType
}

// TODO make this should work the inventory data structure
func (l *Library) GetDefaultChildHardwareBuildOut(deviceTypeString string) (results []HardwareBuildOut, err error) {
	queue := []HardwareBuildOut{
		{
			DeviceTypeString: deviceTypeString,
			Path:             []string{}, // This is the root of the path
			Ordinal:          -1,
		},
	}

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]

		fmt.Println("Visiting: ", current.DeviceTypeString)
		currentDeviceType, ok := l.DeviceTypes[current.DeviceTypeString]
		if !ok {
			panic(fmt.Sprint("Device type does not exist", current.DeviceType))
		}

		// Retrieve the hardware type at this point in time, so we only lookup in the map once
		current.DeviceType = currentDeviceType
		current.HardwareTypePath = append(current.HardwareTypePath, current.DeviceType.HardwareType)

		for _, deviceBay := range currentDeviceType.DeviceBays {
			fmt.Println("  Device bay:", deviceBay.Name)
			if deviceBay.Default != nil {
				fmt.Println("    Default:", deviceBay.Default.Slug)

				// Extract the ordinal
				// This is one way of going about, but it assumes that each name has a number
				// There are two other ways to consider:
				// - Embed an actual ordinal number in the yaml files
				// - Get all of the device base with that type, and then sort them lexicographically. This is how HSM does it, but assumes the names can be sorted in a predictable order
				r := regexp.MustCompile(`\d+`)
				match := r.FindString(deviceBay.Name)
				fmt.Printf("%s|%s\n", deviceBay.Name, match)

				var ordinal int
				if match != "" {
					ordinal, err = strconv.Atoi(match)
					if err != nil {
						return nil, errors.Join(
							fmt.Errorf("unable extract ordinal from device bay name (%s) from device type (%s)", deviceBay.Name, current.DeviceTypeString),
							err,
						)
					}
				}

				queue = append(queue, HardwareBuildOut{
					// Hardware type is deferred until when it is processed
					DeviceTypeString: deviceBay.Default.Slug,
					Path:             append(current.Path, deviceBay.Name),
					Ordinal:          ordinal,
					OrdinalPath:      append(current.OrdinalPath, ordinal),
					HardwareTypePath: current.HardwareTypePath,
				})
			}
		}

		results = append(results, current)
	}

	return results[1:], nil
}
