package import_

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Properties struct {
	RedfishURI string `json:"redfish_uri"`
}

type JSONDeviceRecord struct {
	DeviceType         string     `json:"deviceType"`
	SerialNumber       string     `json:"serialNumber"`
	Manufacturer       string     `json:"manufacturer,omitempty"`
	PartNumber         string     `json:"partNumber,omitempty"`
	ParentSerialNumber string     `json:"parentSerialNumber,omitempty"`
	Properties         Properties `json:"properties"`
}

type DiscoverySnapshotSpec struct {
	RawData []JSONDeviceRecord `json:"rawData"`
}

type DiscoverySnapshotMetadata struct {
	Name string `json:"name"`
}

type DiscoverySnapshot struct {
	APIVersion string                    `json:"apiVersion"`
	Kind       string                    `json:"kind"`
	Metadata   DiscoverySnapshotMetadata `json:"metadata"`
	Spec       DiscoverySnapshotSpec     `json:"spec"`
}

func ParseJson(filepath string) ([]JSONDeviceRecord, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Json file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file: %w", err)
	}

	var discoverySnapshot DiscoverySnapshot
	json.Unmarshal([]byte(fileBytes), &discoverySnapshot)

	return discoverySnapshot.Spec.RawData, nil
}
