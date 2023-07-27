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

var (
	csvAllowedHeaders = map[string]string{
		"id":             "ID",
		"name":           "Name",
		"type":           "Type",
		"devicetypeslug": "DeviceTypeSlug",
		"status":         "Status",
		"vlan":           "Vlan",
		"role":           "Role",
		"subrole":        "SubRole",
		"alias":          "Alias",
		"nid":            "Nid"}
)

func (csm *CSM) GetFields(hw *inventory.Hardware, fieldNames []string) (values []string, err error) {
	values = make([]string, len(fieldNames))

	rawCsmProps := hw.ProviderProperties["csm"]
	csmProps, ok := rawCsmProps.(map[string]interface{})
	if !ok {
		csmProps = make(map[string]interface{})
	}

	for i, name := range fieldNames {
		switch name {
		case "ID":
			values[i] = fmt.Sprintf("%v", hw.ID)
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
	rawCsmProps := hw.ProviderProperties["csm"]
	csmProps, ok := rawCsmProps.(map[string]interface{})
	if !ok {
		// NodeCard's do not have csm props
		// todo possibly verify that the writable columns are empty in the csv input
		log.Debug().Msgf("Skipping %v of the type %v. It does not have writable properties", hw.ID, hw.Type)
		return
	}

	for key, value := range values {
		switch key {
		case "Vlan":
			modified, err := setVlan(value, csmProps)
			if err != nil {
				return result, err
			}
			if modified {
				result.ModifiedFields = append(result.ModifiedFields, "Vlan")
			}
		case "Role":
			modified := setRole(value, csmProps)
			if modified {
				result.ModifiedFields = append(result.ModifiedFields, "Role")
			}
		case "SubRole":
			modified := setSubRole(value, csmProps)
			if modified {
				result.ModifiedFields = append(result.ModifiedFields, "SubRole")
			}
		case "Alias":
			modified, err := setAlias(value, csmProps, hw)
			if err != nil {
				return result, err
			}
			if modified {
				result.ModifiedFields = append(result.ModifiedFields, "Alias")
			}
		case "Nid":
			modified, err := setNid(value, csmProps)
			if err != nil {
				return result, err
			}
			if modified {
				result.ModifiedFields = append(result.ModifiedFields, "Nid")
			}
		}
	}

	return
}

func setVlan(vlanStr string, csmProperties map[string]interface{}) (bool, error) {
	modified := false
	if vlanStr != "" {
		// todo should vlanStr == "" cause the "HMNVlan" field to be removed?
		vlan, err := strconv.ParseInt(vlanStr, 10, 64)
		if err != nil {
			return modified, err
		}
		current := csmProperties["HMNVlan"]
		vlanFloat := float64(vlan)
		if current != vlanFloat {
			csmProperties["HMNVlan"] = vlanFloat
			modified = true
		}
	}
	return modified, nil
}

func setRole(role string, csmProperties map[string]interface{}) bool {
	modified := false
	if role != "" {
		if role != csmProperties["Role"] {
			csmProperties["Role"] = role
			modified = true
		}
	}
	return modified
}

func setSubRole(subRole string, csmProperties map[string]interface{}) bool {
	modified := false
	currentSubRole, ok := csmProperties["SubRole"]
	if subRole == "" {
		if ok {
			if nil != currentSubRole && subRole != currentSubRole {
				csmProperties["SubRole"] = nil
				modified = true
			}
		}
	} else {
		if subRole != currentSubRole {
			csmProperties["SubRole"] = subRole
			modified = true
		}
	}
	return modified
}

func setNid(nidStr string, csmProperties map[string]interface{}) (bool, error) {
	modified := false
	if nidStr != "" {
		nid, err := strconv.ParseInt(nidStr, 10, 64)
		if err != nil {
			return modified, errors.Join(fmt.Errorf("failed to parse nid %v", nidStr), err)
		}
		currentNidRaw := csmProperties["Nid"]
		currentNid, ok := currentNidRaw.(float64)
		nidFloat := float64(nid)
		if !ok || nidFloat != currentNid {
			csmProperties["Nid"] = nidFloat
			modified = true
		}
	}
	return modified, nil
}

func setAlias(alias string, csmProperties map[string]interface{}, hw *inventory.Hardware) (bool, error) {
	modified := false
	if alias != "" {
		rawAlias, ok := csmProperties["Alias"]
		if !ok {
			a := make([]interface{}, 0)
			a = append(a, alias)
			csmProperties["Alias"] = a
			modified = true
		} else {
			v, ok := rawAlias.([]interface{})
			if !ok {
				return modified, fmt.Errorf("expected the Alias field to be an array in the hardware %v", hw)
			}
			if len(v) > 0 {
				if v[0] != alias {
					v[0] = alias
					csmProperties["Alias"] = v
					modified = true
				}
			} else {
				v = append(v, alias)
				csmProperties["Alias"] = v
				modified = true
			}
		}
	}
	return modified, nil
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
