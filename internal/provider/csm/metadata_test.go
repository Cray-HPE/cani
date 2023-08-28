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
package csm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/cani/pkg/pointers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type MetadataStructTagSuite struct {
	suite.Suite

	structuresUnderTest []interface{}
}

func (suite *MetadataStructTagSuite) SetupSuite() {
	suite.structuresUnderTest = []interface{}{
		Metadata{},
		NodeMetadata{},
		CabinetMetadata{},
	}
}

func (suite *MetadataStructTagSuite) verifyUniqueStructTags(structure interface{}, tagName string) {
	for _, structure := range suite.structuresUnderTest {
		suite.T().Logf("Validating structure: %T", structure)

		foundTagValues := map[string]bool{}

		// Iterate over the fields of the structure to find all tags with name
		structType := reflect.TypeOf(structure)
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			suite.T().Logf("Checking field: %v", field.Name)

			// Verify tag exists
			tagValue, exists := field.Tag.Lookup("json")
			suite.Truef(exists, "Structure %T is missing %s struct tag on field %v", structure, tagName, field)
			if !exists {
				continue
			}
			tagValue = strings.TrimSuffix(tagValue, ",omitempty")

			// Check for duplicates
			if _, exists := foundTagValues[tagValue]; exists {
				suite.Fail(fmt.Sprintf("Structure %T has a non-unique %s struct tag value of %v", structure, tagName, tagValue))
			}
			foundTagValues[tagValue] = true
		}
	}
}

func (suite *MetadataStructTagSuite) TestVerifyUniqueJSONTags() {
	for _, structure := range suite.structuresUnderTest {
		suite.T().Logf("Validating structure: %T", structure)
		suite.verifyUniqueStructTags(structure, "json")
	}
}

func (suite *MetadataStructTagSuite) TestVerifyUniqueMapstructureTags() {
	for _, structure := range suite.structuresUnderTest {
		suite.verifyUniqueStructTags(structure, "mapstructure")
	}
}

func (suite *MetadataStructTagSuite) TestVerifyJSONMapstructureTagsMatch() {
	for _, structure := range suite.structuresUnderTest {
		// Iterate over the fields of the structure to find all tags with name
		structType := reflect.TypeOf(structure)
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			suite.T().Logf("Checking field: %v", field.Name)

			// Grab JSON tag value
			jsonTagValue, exists := field.Tag.Lookup("json")
			suite.Truef(exists, "Structure %T is missing json struct tag on field %v", structure, field)
			if !exists {
				continue
			}
			jsonTagValue = strings.TrimSuffix(jsonTagValue, ",omitempty")

			// Grab the mapstructure tag
			mapstructTagValue, exists := field.Tag.Lookup("mapstructure")
			suite.Truef(exists, "Structure %T is missing mapstructure struct tag on field %v", structure, field)
			if !exists {
				continue
			}
			mapstructTagValue = strings.TrimSuffix(mapstructTagValue, ",omitempty")

			// Verify they are the same
			suite.Equal(jsonTagValue, mapstructTagValue, "Structure %T has mis-match json/mapstructure tag values on field %v")
		}
	}
}

func TestMetadataStructTagSuite(t *testing.T) {
	suite.Run(t, new(MetadataStructTagSuite))
}

type BuildHardwareMetadataTestSuite struct {
	suite.Suite

	csm *CSM
}

func (suite *BuildHardwareMetadataTestSuite) SetupSuite() {
	hardwareTypeLibrary, err := hardwaretypes.NewEmbeddedLibrary("")
	suite.NoError(err)

	// Stand up a minimal CSM provider to run BuildHardwareMetadata
	suite.csm = &CSM{
		hardwareLibrary: hardwareTypeLibrary,
	}
}

func (suite *BuildHardwareMetadataTestSuite) TestCabinet() {
	rawProperties := map[string]interface{}{
		ProviderMetadataVlanId: 1234,
	}

	hardware := inventory.Hardware{
		ID:             uuid.New(),
		Type:           hardwaretypes.Cabinet,
		DeviceTypeSlug: "hpe-ex4000",
		Vendor:         "HPE",
		Model:          "EX4000",
		Status:         inventory.HardwareStatusStaged,
	}

	err := suite.csm.BuildHardwareMetadata(&hardware, rawProperties)
	suite.NoError(err)

	suite.NotNil(hardware.ProviderMetadata)
	suite.Contains(hardware.ProviderMetadata, inventory.CSMProvider, "CSM Metadata is missing from cabinet")

	csmMetadata, err := DecodeProviderMetadata(hardware)
	suite.NoError(err)

	suite.Nil(csmMetadata.Node, "Cabinet has Node metadata set")
	suite.NotNil(csmMetadata.Cabinet, "Cabinet is missing node metadata")

	expectedMetadata := Metadata{
		Cabinet: &CabinetMetadata{
			HMNVlan: pointers.IntPtr(1234),
		},
	}
	suite.EqualValues(expectedMetadata, csmMetadata)
}

func (suite *BuildHardwareMetadataTestSuite) TestNode() {

}

func (suite *BuildHardwareMetadataTestSuite) TestNodeUpdateOneFieldExistingData() {

}

func TestBuildHardwareMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(BuildHardwareMetadataTestSuite))
}

type EncodeProviderMetadataTestSuite struct {
	suite.Suite
}

func (suite *EncodeProviderMetadataTestSuite) TestCabinet() {
	metadata := Metadata{
		Cabinet: &CabinetMetadata{
			HMNVlan: pointers.IntPtr(4321),
		},
	}

	metadataRaw, err := EncodeProviderMetadata(metadata)
	suite.NoError(err)

	expectedMetadata := map[string]interface{}{
		"Cabinet": map[string]interface{}{
			"HMNVlan": pointers.IntPtr(4321),
		},
	}
	suite.EqualValues(expectedMetadata, metadataRaw)
}

func (suite *EncodeProviderMetadataTestSuite) TestNode() {
	metadata := Metadata{
		Node: &NodeMetadata{
			Role:    pointers.StringPtr("Management"),
			SubRole: pointers.StringPtr("Worker"),
			Nid:     pointers.IntPtr(10000),
			Alias: []string{
				"ncn-w001",
			},
		},
	}

	metadataRaw, err := EncodeProviderMetadata(metadata)
	suite.NoError(err)

	expectedMetadata := map[string]interface{}{
		"Node": map[string]interface{}{
			"Role":    pointers.StringPtr("Management"),
			"SubRole": pointers.StringPtr("Worker"),
			"Nid":     pointers.IntPtr(10000),
			"Alias": []string{
				"ncn-w001",
			},
			"AdditionalProperties": map[string]interface{}(nil),
		},
	}
	suite.EqualValues(expectedMetadata, metadataRaw)
}

func TestEncodeProviderMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(EncodeProviderMetadataTestSuite))
}
