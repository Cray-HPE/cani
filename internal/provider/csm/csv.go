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
	"errors"
	"fmt"
	"strconv"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
)

func (csm *CSM) GetFields(hw *inventory.Hardware, fieldNames []string) (values []string, err error) {
	values = make([]string, len(fieldNames))

	csmProps, ok := hw.ProviderMetadata["csm"]
	if !ok {
		csmProps = make(map[string]interface{})
	}

	for i, name := range fieldNames {
		switch name {
		case "ID":
			values[i] = fmt.Sprintf("%v", hw.ID)
		case "Location":
			values[i] = fmt.Sprintf("%v", hw.LocationPath)
		case "Name":
			values[i] = fmt.Sprintf("%v", hw.Name)
		case "Type":
			values[i] = fmt.Sprintf("%v", hw.Type)
		case "DeviceTypeSlug":
			values[i] = fmt.Sprintf("%v", hw.DeviceTypeSlug)
		case "Status":
			values[i] = fmt.Sprintf("%v", hw.Status)
		case "Vlan":
			values[i] = toString(csmProps["HMNVlan"])
		case "Role":
			values[i] = toString(csmProps["Role"])
		case "SubRole":
			values[i] = toString(csmProps["SubRole"])
		case "Alias":
			values[i] = getStringFromArray(csmProps["Alias"], 0)
		case "Nid":
			values[i] = toString(csmProps["Nid"])
		default:
			// This case should never be hit.
			// The call to normalize should return an error for unknown headers
			// todo return error
			log.Error().Msgf("Unknown field name %s", name)
			values[i] = ""
		}
	}
	return
}

func (csm *CSM) SetFields(hw *inventory.Hardware, values map[string]string) (result provider.SetFieldsResult, err error) {
	csmMetadata, err := DecodeProviderMetadata(*hw)
	if err != nil {
		return result, err
	}

	if csmMetadata.Cabinet == nil && csmMetadata.Node == nil {
		log.Debug().Msgf("Skipping %v of the type %v. It does not have writable properties", hw.ID, hw.Type)
		return
	}

	if csmMetadata.Node != nil {
		for key, value := range values {
			switch key {
			case "Role":
				modified := setRole(value, csmMetadata.Node)
				if modified {
					result.ModifiedFields = append(result.ModifiedFields, "Role")
				}
			case "SubRole":
				modified := setSubRole(value, csmMetadata.Node)
				if modified {
					result.ModifiedFields = append(result.ModifiedFields, "SubRole")
				}
			case "Alias":
				modified := setAlias(value, csmMetadata.Node)
				if modified {
					result.ModifiedFields = append(result.ModifiedFields, "Alias")
				}
			case "Nid":
				modified, err := setNid(value, csmMetadata.Node)
				if err != nil {
					return result, err
				}
				if modified {
					result.ModifiedFields = append(result.ModifiedFields, "Nid")
				}
			}
		}
	} else if csmMetadata.Cabinet != nil {
		for key, value := range values {
			switch key {
			case "Vlan":
				modified, err := setVlan(value, csmMetadata.Cabinet)
				if err != nil {
					return result, err
				}
				if modified {
					result.ModifiedFields = append(result.ModifiedFields, "Vlan")
				}
			}
		}
	}

	return
}

func setVlan(vlanStr string, cabinetMetadata *CabinetMetadata) (bool, error) {
	modified := false
	if vlanStr == "" {
		if cabinetMetadata.HMNVlan != nil {
			cabinetMetadata.HMNVlan = nil
			modified = true
		}
	} else {
		// todo should vlanStr == "" cause the "HMNVlan" field to be removed?
		vlan, err := strconv.ParseInt(vlanStr, 10, 64)
		if err != nil {
			return modified, err
		}
		vlanInt := int(vlan)
		if cabinetMetadata.HMNVlan == nil || *cabinetMetadata.HMNVlan != vlanInt {
			cabinetMetadata.HMNVlan = &vlanInt
			modified = true
		}
	}
	return modified, nil
}

func setRole(role string, nodeMetadata *NodeMetadata) bool {
	modified := false
	if role == "" {
		if nodeMetadata.Role != nil {
			nodeMetadata.Role = nil
			modified = true
		}
	} else {
		if nodeMetadata.Role == nil || role != *nodeMetadata.Role {
			nodeMetadata.Role = &role
			modified = true
		}
	}
	return modified
}

func setSubRole(subRole string, nodeMetadata *NodeMetadata) bool {
	modified := false
	if subRole == "" {
		if nodeMetadata.SubRole != nil {
			nodeMetadata.SubRole = nil
			modified = true
		}
	} else {
		if nodeMetadata.SubRole == nil || subRole != *nodeMetadata.SubRole {
			nodeMetadata.SubRole = &subRole
			modified = true
		}
	}
	return modified
}

func setNid(nidStr string, nodeMetadata *NodeMetadata) (bool, error) {
	modified := false
	if nidStr == "" {
		if nodeMetadata.Nid != nil {
			nodeMetadata.Nid = nil
			modified = true
		}
	} else {
		nid, err := strconv.ParseInt(nidStr, 10, 64)
		if err != nil {
			return modified, errors.Join(fmt.Errorf("failed to parse nid %v", nidStr), err)
		}
		nidInt := int(nid)
		if nodeMetadata.Nid == nil || nidInt != *nodeMetadata.Nid {
			nodeMetadata.Nid = &nidInt
			modified = true
		}
	}
	return modified, nil
}

func setAlias(alias string, nodeMetadata *NodeMetadata) bool {
	modified := false
	// todo what should be done with an empty string
	//   should it remove all alias, or remove the first element, or do nothing
	if alias != "" {
		if len(nodeMetadata.Alias) > 0 {
			if nodeMetadata.Alias[0] != alias {
				nodeMetadata.Alias[0] = alias
				modified = true
			}
		} else {
			nodeMetadata.Alias = append(nodeMetadata.Alias, alias)
			modified = true
		}
	}
	return modified
}

func getStringFromArray(value interface{}, i int) string {
	if value == nil || i < 0 {
		return ""
	}
	v, ok := value.([]interface{})
	if !ok {
		return ""
	}
	if len(v) <= i {
		return ""
	}
	return fmt.Sprintf("%v", v[i])
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
