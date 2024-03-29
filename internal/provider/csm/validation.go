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
package csm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/util/uuidutil"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Validate the external services of the inventory provider are correct
func (csm *CSM) ValidateExternal(cmd *cobra.Command, args []string) error {

	// Get the dumpate from SLS
	slsState, reps, err := csm.slsClient.DumpstateApi.DumpstateGet(cmd.Context())
	if err != nil {
		return fmt.Errorf("SLS dumpstate failed. %v\n", err)
	}

	// Validate the dumpstate returned from SLS
	_, err = csm.TBV.ValidateHTTPResponse(&slsState, reps)
	if err != nil {
		return fmt.Errorf("Validation failed. %v\n", err)
	}
	return nil
}

// Validate the representation of the inventory data into the destination inventory system
// is consistent. The default set of checks will verify all currently provided data is valid.
// If enableRequiredDataChecks is set to true, additional checks focusing on missing data will be ran.
func (csm *CSM) ValidateInternal(cmd *cobra.Command, args []string, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]provider.HardwareValidationResult, error) {
	log.Debug().Msg("Validating datastore contents against the CSM Provider")

	allHardware, err := datastore.List()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to list hardware from the datastore"),
			err,
		)
	}

	// Build up the validation results map
	results := map[uuid.UUID]provider.HardwareValidationResult{}
	for _, cHardware := range allHardware.Hardware {
		results[cHardware.ID] = provider.HardwareValidationResult{
			Hardware: cHardware,
		}
	}

	// Perform validations
	if err := csm.validateInternalNode(allHardware.Hardware, enableRequiredDataChecks, results); err != nil {
		return nil, err
	}
	if err := csm.validateInternalCabinet(allHardware.Hardware, enableRequiredDataChecks, results); err != nil {
		return nil, err
	}

	//
	// Build results
	//
	resultsWithErrors := map[uuid.UUID]provider.HardwareValidationResult{}
	for id, result := range results {
		if len(result.Errors) > 0 {
			resultsWithErrors[id] = result
		}
	}

	if len(resultsWithErrors) > 0 {
		return resultsWithErrors, provider.ErrDataValidationFailure
	}

	return nil, nil
}

func (csm *CSM) validateInternalNode(allHardware map[uuid.UUID]inventory.Hardware, enableRequiredDataChecks bool, results map[uuid.UUID]provider.HardwareValidationResult) error {
	validRoles := map[string]bool{}
	for _, role := range csm.Options.ValidRoles {
		validRoles[role] = true
	}
	validSubRoles := map[string]bool{}
	for _, subRole := range csm.Options.ValidSubRoles {
		validSubRoles[subRole] = true
	}

	// Verify all specified Node metadata is valid
	nodeNIDLookup := map[int][]uuid.UUID{}
	nodeAliasLookup := map[string][]uuid.UUID{}
	for _, cHardware := range allHardware {
		if cHardware.Type != hardwaretypes.Node {
			continue
		}
		log.Trace().Msgf("Validating %s: %v", cHardware.ID, cHardware)

		metadata, err := DecodeProviderMetadata(cHardware)
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to get provider metadata from hardware (%s)", cHardware.ID),
				err,
			)
		}

		// There is no metadata for this node
		if metadata.Node == nil {
			log.Debug().Msgf("No metadata found for %s", cHardware.ID)
			metadata.Node = &NodeMetadata{}
		}

		//
		// Uniques checks
		// The following are checks that can be performed all the time
		// as it just verifies that the data being added is unique.
		// Ideally should be ran as early as possible
		//

		validationResult := results[cHardware.ID]

		// Verify all specified Roles are valid
		if metadata.Node.Role != nil {
			if !validRoles[*metadata.Node.Role] {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified role (%s) is invalid, choose from: %s", *metadata.Node.Role, strings.Join(csm.Options.ValidRoles, ", ")),
				)
			}
		}

		// Verify all specified SubRoles are valid
		if metadata.Node.SubRole != nil {
			if !validSubRoles[*metadata.Node.SubRole] {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified sub-role (%s) is invalid, choose from: %s", *metadata.Node.SubRole, strings.Join(csm.Options.ValidSubRoles, ", ")),
				)
			}
		}

		// Verify NID is valid
		if metadata.Node.Nid != nil {
			nodeNIDLookup[*metadata.Node.Nid] = append(nodeNIDLookup[*metadata.Node.Nid], cHardware.ID)
			if *metadata.Node.Nid <= 0 {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified NID (%d) invalid, needs to be positive integer", *metadata.Node.Nid),
				)
			}
		}

		// Verify Alias is valid
		if metadata.Node.Alias != nil {
			for _, alias := range metadata.Node.Alias {
				nodeAliasLookup[alias] = append(nodeAliasLookup[alias], cHardware.ID)

				if metadata.Node.Alias != nil && len(alias) == 0 {
					validationResult.Errors = append(validationResult.Errors, "Specified Alias is empty")
				}

				// TODO a regex here might be better
				if strings.Contains(alias, " ") {
					validationResult.Errors = append(validationResult.Errors,
						fmt.Sprintf("Specified alias (%d) is invalid, alias contains spaces", *metadata.Node.Nid),
					)
				}
			}
		}

		if enableRequiredDataChecks {
			//
			// Missing data checks
			// These checks should be ran at reconcile time or via a command line options
			// to ensure all of the required data is present in the datastore before
			//

			// Required Node data checks. All nodes require
			// - Alias
			// - NID
			// - Role
			if metadata.Node.Role == nil {
				validationResult.Errors = append(validationResult.Errors, "Missing required information: Role is not set")
			}
			if metadata.Node.Nid == nil {
				validationResult.Errors = append(validationResult.Errors, "Missing required information: NID is not set")
			}
			if metadata.Node.Alias == nil {
				validationResult.Errors = append(validationResult.Errors, "Missing required information: Alias is not set")
			}

		}

		results[cHardware.ID] = validationResult
	}

	//
	// Uniqueness checks
	//

	// Verify all specified NIDs are unique
	for nid, matchingHardware := range nodeNIDLookup {
		if len(matchingHardware) > 1 {
			// We found hardware with duplicate NIDs
			for _, id := range matchingHardware {
				validationResult := results[id]
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified NID (%d) is not unique, shared by: %s", nid, uuidutil.Join(matchingHardware, ", ", id)),
				)
				results[id] = validationResult
			}
		}
	}

	// Verify all specified Aliases are unique
	for alias, matchingHardware := range nodeAliasLookup {
		if len(matchingHardware) > 1 {
			// We found hardware with duplicate NIDs
			for _, id := range matchingHardware {
				validationResult := results[id]
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified alias (%s) is not unique, shared by: %s", alias, uuidutil.Join(matchingHardware, ", ", id)),
				)
				results[id] = validationResult
			}
		}
	}

	return nil
}

func (csm *CSM) validateInternalCabinet(allHardware map[uuid.UUID]inventory.Hardware, enableRequiredDataChecks bool, results map[uuid.UUID]provider.HardwareValidationResult) error {
	// Verify all specified Cabinet metadata is valid
	cabinetVLANLookup := map[int][]uuid.UUID{}
	for _, cHardware := range allHardware {
		if cHardware.Type != hardwaretypes.Cabinet {
			continue
		}

		log.Trace().Msgf("Validating %s: %v", cHardware.ID, cHardware)

		metadata, err := DecodeProviderMetadata(cHardware)
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to get provider metadata from hardware (%s)", cHardware.ID),
				err,
			)
		}

		// There is no metadata for this cabinet
		if metadata.Cabinet == nil {
			log.Debug().Msgf("No metadata found for %s", cHardware.ID)
			metadata.Cabinet = &CabinetMetadata{}
		}

		validationResult := results[cHardware.ID]

		if metadata.Cabinet.HMNVlan != nil {
			// Verify the vlan is within the allowed range
			if !(0 <= *metadata.Cabinet.HMNVlan && *metadata.Cabinet.HMNVlan <= 4094) {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified HMN Vlan (%d) is invalid, must be in range: 0-4094", *metadata.Cabinet.HMNVlan),
				)
			}

			cabinetVLANLookup[*metadata.Cabinet.HMNVlan] = append(cabinetVLANLookup[*metadata.Cabinet.HMNVlan], cHardware.ID)
		}

		if enableRequiredDataChecks {
			// Check for missing required data

			// The HMN vlan is only required if we are populating a Cray EX managed cabinet, so lets check to see if a CEC is a child of the cabinet
			// TODO what if the datastore has a get relative function?
			cecManagedCabinet := false
			for _, childID := range cHardware.Children {
				childHardware, ok := allHardware[childID]
				if !ok {
					// This should not happen
					return fmt.Errorf("unable to find hardware object with ID (%s)", childID)
				}

				if childHardware.Type == hardwaretypes.CabinetEnvironmentalController {
					cecManagedCabinet = true
					break
				}

			}

			if cecManagedCabinet {
				if metadata.Cabinet.HMNVlan == nil {
					validationResult.Errors = append(validationResult.Errors, "Missing required information: HMN Vlan is not set")
				}
			}
		}

		results[cHardware.ID] = validationResult
	}

	//
	// Uniqueness checks
	//

	// Verify all specified Cabinet VLANs are unique
	for nid, matchingHardware := range cabinetVLANLookup {
		if len(matchingHardware) > 1 {
			// We found hardware with duplicate NIDs
			for _, id := range matchingHardware {
				validationResult := results[id]
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified HMN Vlan (%d) is not unique, shared by: %s", nid, uuidutil.Join(matchingHardware, ", ", id)),
				)
				results[id] = validationResult
			}
		}
	}

	return nil
}
