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

package checks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/rs/zerolog/log"
)

const (
	UniqueAlias                      common.ValidationCheck = "unique-alias"
	UniqueNid                        common.ValidationCheck = "unique-nid"
	HardwareNodeCheck                common.ValidationCheck = "hardware-node"
	HardwareMgmtSwitchConnectorCheck common.ValidationCheck = "hardware-mgmt-switch-connector"
	HardwareMgmtSwitchCheck          common.ValidationCheck = "hardware-mgmt-switch"
	SwitchBrandCheck                 common.ValidationCheck = "switch-brand"
	SwitchCredentialsCheck           common.ValidationCheck = "switch-credentials"
	SwitchVendorCheck                common.ValidationCheck = "switch-vendor"
	SwitchConnectorNodeNicsCheck     common.ValidationCheck = "switch-connector-node-nics"
	SwitchConnectorCheck             common.ValidationCheck = "switch-connector"
	SwitchSNMPPropertiesCheck        common.ValidationCheck = "switch-snmp-properties"
	SwitchIpAddressCheck             common.ValidationCheck = "switch-ip-address"
	HardwareMgmtHLSwitchCheck        common.ValidationCheck = "hardware-mgmt-hl-switch"
	HardwareRouterBMC                common.ValidationCheck = "hardware-router-bmc"
	CabinetNetworkCheck              common.ValidationCheck = "hardware-cabinet-network"
)

type HardwareCheck struct {
	hardware          map[string]sls_client.Hardware
	typeToHardware    map[string][]*sls_client.Hardware
	parentHasChildren map[string]struct{}
	networks          map[string]sls_client.Network // only the HMN network is used in the MgmtSwitch and MgmtHLSwitch checks
	validRoles        []string
	validSubRoles     []string
}

func NewHardwareCheck(
	hardware map[string]sls_client.Hardware,
	typeToHardware map[string][]*sls_client.Hardware,
	parentHasChildren map[string]struct{},
	networks map[string]sls_client.Network,
	validRoles []string,
	validSubRoles []string) *HardwareCheck {
	hardwareCheck := HardwareCheck{
		hardware:          hardware,
		typeToHardware:    typeToHardware,
		parentHasChildren: parentHasChildren,
		networks:          networks,
		validRoles:        validRoles,
		validSubRoles:     validSubRoles,
	}
	return &hardwareCheck
}

func (c *HardwareCheck) Validate(results *common.ValidationResults) {

	aliasToHardware := make(map[string]*sls_client.Hardware)
	nidToHardware := make(map[int64]*sls_client.Hardware)
	cabinetNetworkChecker := NewHardwareCabinetNetworkSubCheck(c.networks)

	for _, h := range c.hardware {
		hardware := h
		props := getProps(&h)

		validateUniqueAlias(results, aliasToHardware, &hardware, props)
		validateUniqueNid(results, nidToHardware, &hardware, props)

		switch h.TypeString {
		case xnametypes.Node:
			validateNode(results, &h, props, c.validRoles, c.validSubRoles)
		case xnametypes.MgmtSwitchConnector:
			validateMgmtSwitchConnector(results, &h, props, c.parentHasChildren, c.hardware)
		case xnametypes.MgmtSwitch:
			validateMgmtSwitch(results, &h, props, c.networks)
		case xnametypes.MgmtHLSwitch:
			validateMgmtHLSwitch(results, &h, props, c.networks)
		case xnametypes.RouterBMC:
			validateRouterBMC(results, &h, props)
		case xnametypes.Cabinet:
			cabinetNetworkChecker.Validate(results, &h, props)
		}
	}
}

func getProps(h *sls_client.Hardware) map[string]interface{} {
	// todo improve this
	// common.GetMap uses the structs.Map library,
	// however, this library does not convert the substructs to maps.
	// something is needed to convert all levels of the data to map[string]interface{}
	_, ok := h.ExtraProperties.(sls_client.HardwareExtraPropertiesCabinet)
	if ok {
		props := make(map[string]interface{})
		rawProps, err := json.Marshal(h.ExtraProperties)
		if err != nil {
			log.Debug().Msgf("Failed to marshal ExtraProperties for hardware: %s %s, struct type: %T, error: %s",
				h.Xname, h.Type, h.ExtraProperties, err)
			return props
		}
		if err := json.Unmarshal(rawProps, &props); err != nil {
			log.Debug().Msgf("Failed to unmarshal ExtraProperties for hardware: %s %s, struct type: %T, error: %s",
				h.Xname, h.Type, h.ExtraProperties, err)
			return props
		}
		return props
	}
	props, _ := common.GetMap(h.ExtraProperties)
	return props
}

func validateUniqueAlias(
	results *common.ValidationResults,
	aliasToHardware map[string]*sls_client.Hardware,
	hardware *sls_client.Hardware,
	props map[string]interface{}) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)

	aliases, _ := common.GetSliceOfStrings(props, "Aliases")

	found := false
	for _, alias := range aliases {
		if strings.TrimSpace(alias) == "" {
			results.Fail(
				UniqueAlias,
				componentId,
				fmt.Sprintf("Empty alias '%s' for %s.", alias, hardware.Xname))
			break
		}
		found = true
		otherHardware, ok := aliasToHardware[alias]
		if ok {
			results.Fail(
				UniqueAlias,
				componentId,
				fmt.Sprintf("The alias %s for %s is not unique. It conflicts with %s.", alias, hardware.Xname, otherHardware.Xname))
		}
	}

	if !found {
		if hardware.TypeString == xnametypes.Node ||
			hardware.TypeString == xnametypes.MgmtSwitch ||
			hardware.TypeString == xnametypes.CDUMgmtSwitch {
			results.Fail(
				UniqueAlias,
				componentId,
				fmt.Sprintf("%s %s does not have an alias.", hardware.Xname, hardware.TypeString))
		}
		return
	}
}

func validateUniqueNid(
	results *common.ValidationResults,
	nidToHardware map[int64]*sls_client.Hardware,
	hardware *sls_client.Hardware,
	props map[string]interface{}) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)

	nid, ok := common.ToInt(props["NID"])
	if hardware.TypeString != "Node" {
		if ok {
			results.Fail(
				UniqueNid,
				componentId,
				fmt.Sprintf("%s should not have a NID %v, because it does not have the Node Type.", hardware.Xname, props["NID"]))
		}
		return
	}

	role, _ := common.GetString(props, "Role")
	if role == "Application" {
		if ok {
			results.Fail(
				UniqueNid,
				componentId,
				fmt.Sprintf("%s should not have a NID %v, because it is has the Application Role.", hardware.Xname, props["NID"]))
		}
		return
	}

	if !ok {
		results.Fail(
			UniqueNid,
			componentId,
			fmt.Sprintf("%s does not have a NID", hardware.Xname))
		return
	}
	// todo report both hardware objects as having a non unique nid
	otherHardware, ok := nidToHardware[nid]
	if ok {
		results.Fail(
			UniqueNid,
			componentId,
			fmt.Sprintf("The NID %d for %s is not unique. It conflicts with %s.", nid, hardware.Xname, otherHardware.Xname))
		return
	}
	nidToHardware[nid] = hardware
	results.Pass(
		UniqueNid,
		componentId,
		fmt.Sprintf("The NID %d for %s %s is unique.", nid, hardware.Xname, hardware.TypeString))
}

func validateNode(results *common.ValidationResults, hardware *sls_client.Hardware, props map[string]interface{}, validRoles, validSubRoles []string) {
	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	role, found := common.GetString(props, "Role")
	if found {
		if contains(role, validRoles) {
			results.Pass(
				HardwareNodeCheck,
				componentId,
				fmt.Sprintf("%s %s has a valid Role: %s", hardware.Xname, hardware.TypeString, role))
		} else {
			results.Fail(
				HardwareNodeCheck,
				componentId,
				fmt.Sprintf("%s %s has an invalid Role: %s, Valid Roles are: %v", hardware.Xname, hardware.TypeString, role, validRoles))
		}
	} else {
		results.Fail(
			HardwareNodeCheck,
			componentId,
			fmt.Sprintf("%s %s is missing the Role field", hardware.Xname, hardware.TypeString))
	}

	subRole, found := common.GetString(props, "SubRole")
	if found {
		if subRole == "" || contains(subRole, validSubRoles) {
			results.Pass(
				HardwareNodeCheck,
				componentId,
				fmt.Sprintf("%s %s has a valid SubRole: %s", hardware.Xname, hardware.TypeString, subRole))
		} else {
			results.Fail(
				HardwareNodeCheck,
				componentId,
				fmt.Sprintf("%s %s has an invalid SubRole: %s, Valid SubRoles are: %v", hardware.Xname, hardware.TypeString, subRole, validSubRoles))
		}
	}
}

func validateMgmtSwitchConnector(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	parentHasChildren map[string]struct{},
	hardwareMap map[string]sls_client.Hardware) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)

	nodeNics, found := common.GetSliceOfStrings(props, "NodeNics")
	if !found || len(nodeNics) == 0 {
		results.Fail(
			SwitchConnectorNodeNicsCheck,
			componentId,
			fmt.Sprintf("%s %s is missing NodeNics.", hardware.Xname, hardware.TypeString))
	} else {
		for _, nodeNic := range nodeNics {
			t := xnametypes.GetHMSType(nodeNic)

			foundNodeNic := false
			message := "in the hardware list"
			if xnametypes.IsHMSTypeController(t) {
				_, found := hardwareMap[nodeNic]
				if !found {
					if t == xnametypes.NodeBMC && !strings.HasSuffix(nodeNic, "999") {
						// a NodeNic of type NodeBMC will not exist in the hardware list.
						// but it will exist as a parent of some hardware
						// With the exception of NodeBMC's who's xname ends in 999, these should be in the hardware list.
						_, found = parentHasChildren[nodeNic]
						message = "in the hardware list as the parent of at least one hardware entry"
					}
				}
				foundNodeNic = found
			} else {
				types := strings.Join([]string{xnametypes.ChassisBMC.String(), xnametypes.NodeBMC.String(), xnametypes.RouterBMC.String()}, ",")
				results.Fail(
					SwitchConnectorNodeNicsCheck,
					componentId,
					fmt.Sprintf("%s %s a NodeNic %s is of the type %s when it should be of one of the following types %s.",
						hardware.Xname, hardware.TypeString, nodeNic, t, types))
			}

			if foundNodeNic {
				results.Pass(
					SwitchConnectorNodeNicsCheck,
					componentId,
					fmt.Sprintf("%s %s NodeNic %s exists %s.",
						hardware.Xname, hardware.TypeString, nodeNic, message))
			} else {
				results.Fail(
					SwitchConnectorNodeNicsCheck,
					componentId,
					fmt.Sprintf("%s %s NodeNic %s is missing. There should be a hardware entry for the NodeNic, or hardware entry with the NodeNic as a parent.",
						hardware.Xname, hardware.TypeString, nodeNic))
			}
		}
	}

	parent, found := hardwareMap[hardware.Parent]
	if found {
		results.Pass(
			SwitchConnectorCheck,
			componentId,
			fmt.Sprintf("%s %s parent %s exists. ", hardware.Xname, hardware.TypeString, hardware.Parent))
	} else {
		results.Fail(
			SwitchConnectorCheck,
			componentId,
			fmt.Sprintf("%s %s parent %s is missing. ", hardware.Xname, hardware.TypeString, hardware.Parent))
		return
	}
	parentProps, _ := common.GetMap(parent.ExtraProperties)
	brand, exists := common.GetString(parentProps, "Brand")
	if !exists {
		// this will be checked by the MgmtSwitch checks
		return
	}

	arubaPattern := regexp.MustCompile("^[0-9]+/[0-9]+/[0-9]+$")
	dellPattern := regexp.MustCompile("^ethernet[0-9]+/[0-9]+/[0-9]+$")
	fieldName := "VendorName"
	vendorName, exists := common.GetString(props, fieldName)
	if exists {
		valid := false
		switch brand {
		case "Aruba":
			if arubaPattern.MatchString(vendorName) {
				results.Pass(
					SwitchVendorCheck,
					componentId,
					fmt.Sprintf("%s %s the value %s in the %s property is correct",
						hardware.Xname, hardware.TypeString, vendorName, fieldName))
				valid = true
			}
		case "Dell":
			if dellPattern.MatchString(vendorName) {
				results.Pass(
					SwitchVendorCheck,
					componentId,
					fmt.Sprintf("%s %s the value %s in the %s property is correct",
						hardware.Xname, hardware.TypeString, vendorName, fieldName))
				valid = true
			}
		default:
			if arubaPattern.MatchString(vendorName) || dellPattern.MatchString(vendorName) {
				results.Pass(
					SwitchVendorCheck,
					componentId,
					fmt.Sprintf("%s %s the value %s in the %s property is correct",
						hardware.Xname, hardware.TypeString, vendorName, fieldName))
				valid = true
			}
		}
		if !valid {
			results.Fail(
				SwitchVendorCheck,
				componentId,
				fmt.Sprintf("%s %s the %s property is missing",
					hardware.Xname, hardware.TypeString, fieldName))
		}
	}
}

func validateMgmtSwitch(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	networks map[string]sls_client.Network) {

	validateSwitchBrand(results, hardware, props)

	validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "SNMPAuthProtocol")
	if value, found := validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "SNMPAuthPassword"); found {
		validateVaultField(results, hardware, props, "SNMPAuthPassword", value)
	}

	validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "SNMPPrivProtocol")
	if value, found := validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "SNMPPrivPassword"); found {
		validateVaultField(results, hardware, props, "SNMPPrivPassword", value)
	}

	validateIP4addr(results, hardware, props, networks)
}

func validateMgmtHLSwitch(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	networks map[string]sls_client.Network) {

	validateSwitchBrand(results, hardware, props)
	validateIP4addr(results, hardware, props, networks)
}

func validateRouterBMC(results *common.ValidationResults, hardware *sls_client.Hardware, props map[string]interface{}) {
	if value, found := validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "Password"); found {
		validateVaultField(results, hardware, props, "Password", value)
	}
	if value, found := validateFieldExists(results, hardware, props, SwitchCredentialsCheck, "Username"); found {
		validateVaultField(results, hardware, props, "Username", value)
	}
}

func validateSwitchBrand(results *common.ValidationResults, hardware *sls_client.Hardware, props map[string]interface{}) {
	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	brand, found := common.GetString(props, "Brand")
	brands := getValidBrands()
	if found {
		if contains(brand, brands) {
			results.Pass(
				SwitchBrandCheck,
				componentId,
				fmt.Sprintf("%s %s has a valid Brand %s", hardware.Xname, hardware.TypeString, brand))
		} else {
			results.Fail(
				SwitchBrandCheck,
				componentId,
				fmt.Sprintf("%s %s has an invalid Brand %s, valid brands are: %s", hardware.Xname, hardware.TypeString, brand, strings.Join(brands, ",")))
		}
	} else {
		results.Fail(
			SwitchBrandCheck,
			componentId,
			fmt.Sprintf("%s %s is missing the Brand field, valid brands are: %s", hardware.Xname, hardware.TypeString, strings.Join(brands, ",")))
	}
}

func validateFieldExists(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	check common.ValidationCheck,
	fieldName string) (field string, exists bool) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	field, exists = common.GetString(props, fieldName)
	if exists {
		results.Pass(
			check,
			componentId,
			fmt.Sprintf("%s %s the %s property exists", hardware.Xname, hardware.TypeString, fieldName))
	} else {
		results.Fail(
			check,
			componentId,
			fmt.Sprintf("%s %s the %s property is missing", hardware.Xname, hardware.TypeString, fieldName))
	}
	return field, exists
}

func validateIP4addr(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	networks map[string]sls_client.Network) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	fieldName := "IP4addr"
	ip, exists := common.GetString(props, fieldName)
	if !exists {
		results.Fail(
			SwitchIpAddressCheck,
			componentId,
			fmt.Sprintf("%s %s the %s property is missing", hardware.Xname, hardware.TypeString, ip))
		return
	}

	hmnNetwork, found := networks["HMN"]
	if !found {
		// todo should this fail silently. If the HMN is missing other checks will generate failures
		results.Fail(
			SwitchIpAddressCheck,
			componentId,
			fmt.Sprintf("%s %s cannot validate the IP Address %s because the HMN network is not defined", hardware.Xname, hardware.TypeString, ip))
		return
	}

	subnetName := "network_hardware"
	hasIP := false
	for _, subnet := range hmnNetwork.ExtraProperties.Subnets {
		if subnet.Name != subnetName {
			continue
		}
		for _, res := range subnet.IPReservations {
			if ip == res.IPAddress {
				hasIP = true
				break
			}
		}
		if hasIP {
			break
		}
	}

	if hasIP {
		results.Pass(
			SwitchIpAddressCheck,
			componentId,
			fmt.Sprintf("%s %s the IP4addr %s has an IP reservation in the %s subnet of the HMN network",
				hardware.Xname, hardware.TypeString, ip, subnetName))
	} else {
		results.Fail(
			SwitchIpAddressCheck,
			componentId,
			fmt.Sprintf("%s %s the IP4addr %s does not have an IP reservation in the %s subnet of the HMN network",
				hardware.Xname, hardware.TypeString, ip, subnetName))
	}

}

func validateVaultField(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{},
	fieldName string,
	value string) {
	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	if isValidVault(hardware.Xname, value) {
		results.Pass(
			SwitchCredentialsCheck,
			componentId,
			fmt.Sprintf("%s %s the %s property is correct", hardware.Xname, hardware.TypeString, fieldName))
	} else {
		results.Fail(
			SwitchCredentialsCheck,
			componentId,
			fmt.Sprintf("%s %s the %s property has incorrect incorrect value %s. It should be vault://hms-creds/%s",
				hardware.Xname, hardware.TypeString, fieldName, value, hardware.Xname))
	}
}

func isValidVault(xname string, field string) bool {
	index := strings.LastIndex(field, "/")
	prefix := field[:index]
	suffix := field[index+1:]
	return prefix == "vault://hms-creds" && suffix == xname
}

func getValidBrands() []string {
	return []string{"Arista", "Aruba", "Dell", "Mellanox"}
}

func contains(str string, list []string) bool {
	for _, item := range list {
		if str == item {
			return true
		}
	}
	return false
}
