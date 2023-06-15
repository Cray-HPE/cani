package csm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/rs/zerolog/log"
)

// Reconcile CANI's inventory state with the external inventory state and apply required changes
// TODO perhaps Reconcile should return a ReconcileResult struct, that can contain what the provider wants to do
// This would enable these two things
//   - Provide a way to pass downwards the result, and allow for a custom string/Presentation function to
//     format the results in a human readable way
//   - Allow for a process like the following:
//     1. Figure out what has changed
//     2. Validate the changes
//     3. Display what changed
//     4. Make changes
func (csm *CSM) Reconcile(ctx context.Context, datastore inventory.Datastore) (err error) {
	// TODO should we have a presentation callback to confirm the removal of hardware?

	log.Info().Msg("Starting CSM reconcile process")

	// TODO grab the allowed HSM Roles and SubRoles from HSM
	// This is for data validation

	//
	// Retrieve the current SLS state
	//
	currentSLSState, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to perform SLS dumpstate"),
			err,
		)
	}

	// DEBUG BEGIN
	currentSLSStateRaw, err := json.MarshalIndent(currentSLSState, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("reconcile_sls_current.json", currentSLSStateRaw, 0600)
	// DEBUG END

	//
	// Build up the expected SLS state
	//
	expectedSLSState, hardwareMapping, err := BuildExpectedHardwareState(datastore)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to build expected SLS state"),
			err,
		)
	}

	// DEBUG BEGIN
	expectedSLSStateRaw, err := json.MarshalIndent(expectedSLSState, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("reconcile_sls_expected.json", expectedSLSStateRaw, 0600)
	// DEBUG END

	// HACK Prune non-supported hardware from the current SLS state.
	// For right only remove all objects from the current SLS state that the SLS generator
	// has no idea about
	currentSLSState.Hardware, _ = sls.FilterHardware(currentSLSState.Hardware, func(hardware sls_client.Hardware) (bool, error) {
		_, exists := hardwareMapping[hardware.Xname]
		return exists, nil
	})

	//
	// Compare the current hardware state with the expected hardware state
	//

	hardwareRemoved, err := sls.HardwareSubtract(currentSLSState, expectedSLSState)
	if err != nil {
		return err
	}

	hardwareAdded, err := sls.HardwareSubtract(expectedSLSState, currentSLSState)
	if err != nil {
		return err
	}

	// Identify hardware present in both states
	// Does not take into account differences in Class/ExtraProperties, just by the primary key of xname
	identicalHardware, hardwareWithDifferingValues, err := sls.HardwareUnion(expectedSLSState, currentSLSState)
	if err != nil {
		return err
	}

	if err := displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware, hardwareWithDifferingValues); err != nil {
		return err
	}

	// debug
	identicalHardwareRaw, err := json.MarshalIndent(identicalHardware, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("reconcile_sls_identical_hardware.json", identicalHardwareRaw, 0600)

	hardwareWithDifferingValuesRaw, err := json.MarshalIndent(hardwareWithDifferingValues, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("reconcile_sls_hardware_with_differing_values.json", hardwareWithDifferingValuesRaw, 0600)
	// debug

	//
	// Verify expected hardware actions are taking place.
	// This can detect drift from when hardware was removed/added outside of CANI after the session was started
	//
	unexpectedHardwareRemoval := []sls_client.Hardware{}
	for _, hardware := range hardwareRemoved {
		if hardwareMapping[hardware.Xname].Status != inventory.HardwareStatusDecommissioned {
			// This piece of hardware wasn't flagged for removal from the system, but a
			// the reconcile logic wants to remove it and this is bad
			unexpectedHardwareRemoval = append(unexpectedHardwareRemoval, hardware)
		}

		// This piece of hardware is allowed to be removed from the system, as it has the right
		// inventory status
	}

	unexpectedHardwareAdditions := []sls_client.Hardware{}
	for _, hardware := range hardwareAdded {
		if hardwareMapping[hardware.Xname].Status != inventory.HardwareStatusStaged {
			// This piece of hardware wasn't flagged to be added to the system, but a
			// the reconcile logic wants to remove it and this is bad
			unexpectedHardwareAdditions = append(unexpectedHardwareAdditions, hardware)
		}
		// This piece of hardware is allowed to be added from the system, as it has the right
		// inventory status
	}

	// TODO need a good way to signal in the inventory structure that we need to support
	// update metadata for a piece of hardware, for now just not handle it
	// for _, hardware := range hardwareWithDifferingValues {
	// }

	if len(unexpectedHardwareRemoval) != 0 || len(unexpectedHardwareAdditions) != 0 {
		displayUnwantedChanges(unexpectedHardwareRemoval, unexpectedHardwareAdditions)
		return fmt.Errorf("detected unexpected hardware changes between current and expected system states")
	}

	//
	// Simulate and validate SLS actions
	//
	modifiedState, err := sls.CopyState(currentSLSState)
	if err != nil {
		return errors.Join(fmt.Errorf("unable to copy SLS state"), err)
	}
	for _, hardware := range hardwareRemoved {
		delete(modifiedState.Hardware, hardware.Xname)
	}
	for _, hardware := range hardwareAdded {
		modifiedState.Hardware[hardware.Xname] = hardware
	}
	for _, hardwarePair := range hardwareWithDifferingValues {
		updatedHardware := hardwarePair.HardwareB
		modifiedState.Hardware[updatedHardware.Xname] = updatedHardware
	}

	// DEBUG BEGIN
	modifiedStateRaw, err := json.MarshalIndent(modifiedState, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("reconcile_sls_modified_state.json", modifiedStateRaw, 0600)
	// DEBUG END

	// _, err = validate.Validate(&modifiedState)
	// if err != nil {
	// 	return fmt.Errorf("Validation failed. %v\n", err)
	// }

	//
	// Modify the System's SLS instance
	//

	// Sort hardware so children are deleted before their parents
	sls.SortHardwareReverse(hardwareRemoved)
	// Remove hardware that no longer exists
	for _, hardware := range hardwareRemoved {
		log.Info().Str("xname", hardware.Xname).Msg("Removing")
		// Put into transaction log with old and new value
		// TODO

		// Perform a DELETE against SLS
		r, err := csm.slsClient.HardwareApi.HardwareXnameDelete(ctx, hardware.Xname)
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to delete hardware (%s) from SLS", hardware.Xname),
				err,
			)
		}
		log.Info().Int("status", r.StatusCode).Msg("Deleted hardware from SLS")
	}

	// Add hardware new hardware
	for _, hardware := range hardwareAdded {
		log.Info().Str("xname", hardware.Xname).Msg("Adding")
		// Put into transaction log with old and new value
		// TODO

		// Perform a POST against SLS
		_, r, err := csm.slsClient.HardwareApi.HardwarePost(ctx, sls.NewHardwarePostOpts(hardware))
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to add hardware (%s) to SLS", hardware.Xname),
				err,
			)
		}
		log.Info().Int("status", r.StatusCode).Msg("Added hardware to SLS")
	}

	// Update existing hardware
	for _, hardwarePair := range hardwareWithDifferingValues {
		updatedHardware := hardwarePair.HardwareB
		log.Info().Str("xname", updatedHardware.Xname).Msg("Updating")
		// Put into transaction log with old and new value
		// TODO

		// Perform a PUT against SLS
		_, r, err := csm.slsClient.HardwareApi.HardwareXnamePut(ctx, updatedHardware.Xname, sls.NewHardwareXnamePutOpts(updatedHardware))
		if err != nil {
			return errors.Join(
				fmt.Errorf("failed to update hardware (%s) from SLS", updatedHardware.Xname),
				err,
			)
		}
		log.Info().Int("status", r.StatusCode).Msg("Updated hardware to SLS")
	}

	return nil
}

//
// The following is taken from: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go
//

func displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware []sls_client.Hardware, hardwareWithDifferingValues []sls.GenericHardwarePair) error {
	log.Info().Msgf("")
	log.Info().Msgf("Identical hardware between current and expected states")
	if len(identicalHardware) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range identicalHardware {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Common hardware between current and expected states with differing class or extra properties")
	if len(hardwareWithDifferingValues) == 0 {
		log.Info().Msg("  None")
	}
	for _, pair := range hardwareWithDifferingValues {
		log.Info().Msgf("  %s", pair.Xname)

		// Expected Hardware json
		pair.HardwareA.LastUpdated = 0
		pair.HardwareA.LastUpdatedTime = ""
		hardwareRaw, err := buildHardwareString(pair.HardwareA)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Expected: %-16s", hardwareRaw)

		// Actual Hardware json
		pair.HardwareB.LastUpdated = 0
		pair.HardwareB.LastUpdatedTime = ""
		hardwareRaw, err = buildHardwareString(pair.HardwareB)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Actual:   %-16s", hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Hardware added to the system")
	if len(hardwareAdded) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range hardwareAdded {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Hardware removed from system")
	if len(hardwareRemoved) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range hardwareRemoved {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	return nil
}

func displayUnwantedChanges(unwantedHardwareRemoved, unwantedHardwareAdded []sls_client.Hardware) error {
	if len(unwantedHardwareAdded) != 0 {
		log.Error().Msgf("")
		log.Error().Msgf("Unexpected Hardware detected added to the system")
		for _, hardware := range unwantedHardwareAdded {
			hardwareRaw, err := buildHardwareString(hardware)
			if err != nil {
				return err
			}

			log.Error().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
		}
	}

	if len(unwantedHardwareRemoved) != 0 {
		log.Error().Msgf("")
		log.Error().Msgf("Unexpected Hardware detected removed from the system")
		for _, hardware := range unwantedHardwareRemoved {
			hardwareRaw, err := buildHardwareString(hardware)
			if err != nil {
				return err
			}

			log.Error().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
		}
	}

	log.Info().Msgf("")
	return nil
}

func buildHardwareString(hardware sls_client.Hardware) (string, error) {
	// TODO include CANU UUID

	extraPropertiesRaw, err := hardware.DecodeExtraProperties()
	if err != nil {
		return "", err
	}

	var tokens []string
	tokens = append(tokens, fmt.Sprintf("Type: %s", hardware.TypeString))

	switch hardware.TypeString {
	// case xnametypes.Cabinet:
	// 	// If we don't know how to pretty print it, lets just do the raw JSON
	// 	extraPropertiesRaw, err := json.Marshal(hardware.ExtraProperties)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	tokens = append(tokens, string(extraPropertiesRaw))
	case xnametypes.Chassis:
		// Nothing to do
	case xnametypes.ChassisBMC:
		// Nothing to do
	case xnametypes.CabinetPDUController:
		// Nothing to do
	case xnametypes.RouterBMC:
		// Nothing to do
	case xnametypes.NodeBMC:
		// Nothing to do
	case xnametypes.Node:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesNode); ok {
			tokens = append(tokens, fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")))
			if extraProperties.Role != "" {
				tokens = append(tokens, fmt.Sprintf("Role: %s", extraProperties.Role))
			}
			if extraProperties.SubRole != "" {
				tokens = append(tokens, fmt.Sprintf("SubRole: %s", extraProperties.SubRole))
			}
			if extraProperties.NID != 0 {
				tokens = append(tokens, fmt.Sprintf("NID: %d", extraProperties.NID))
			}
		}
	case xnametypes.MgmtSwitch:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4addr: %s", extraProperties.IP4addr))
			}
			if extraProperties.IP6addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6addr: %s", extraProperties.IP6addr))
			}

			tokens = append(tokens,
				fmt.Sprintf("SNMPUsername: %s", extraProperties.SNMPUsername),
				fmt.Sprintf("SNMPAuthProtocol: %s", extraProperties.SNMPAuthProtocol),
				fmt.Sprintf("SNMPAuthPassword: %s", extraProperties.SNMPAuthPassword),
				fmt.Sprintf("SNMPPrivProtocol: %s", extraProperties.SNMPPrivProtocol),
				fmt.Sprintf("SNMPPrivPassword: %s", extraProperties.SNMPPrivPassword),
			)
		}
	case xnametypes.MgmtHLSwitch:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtHlSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4addr: %s", extraProperties.IP4addr))
			}
			if extraProperties.IP6addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6addr: %s", extraProperties.IP6addr))
			}
		}
	case xnametypes.MgmtSwitchConnector:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtSwitchConnector); ok {
			tokens = append(tokens,
				fmt.Sprintf("VendorName: %s", extraProperties.VendorName),
				fmt.Sprintf("NodeNics: [%s]", strings.Join(extraProperties.NodeNics, ",")),
			)
		}
	default:
		// If we don't know how to pretty print it, lets just do the raw JSON
		hardwareRaw, err := json.Marshal(hardware)
		if err != nil {
			return "", err
		}
		tokens = append(tokens, string(hardwareRaw))
	}

	return strings.Join(tokens, ", "), nil
}
