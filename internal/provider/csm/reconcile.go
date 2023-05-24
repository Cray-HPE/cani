package csm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/rs/zerolog/log"
)

// Reconcile CANI's inventory state with the external inventory state and apply required changes
// TODO perhaps Reconcole should return a ReconcileResult struct, that can contain what the provider wants to do
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

	// Retrieve the current SLS state
	// TODO
	currentSLSState, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to perform SLS dumpstate"),
			err,
		)
	}
	// Build up the expected SLS state
	expectedSLSState, err := BuildExpectedHardwareState(datastore)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to build expected SLS state"),
			err,
		)
	}

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
	identicalHardware, hardwareWithDifferingValues, err := sls.HardwareUnion(currentSLSState, expectedSLSState)
	if err != nil {
		return err
	}

	if err := displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware, hardwareWithDifferingValues); err != nil {
		return err
	}

	//
	// Modify SLS
	//

	// TODO simulate changes to the SLS state and validate them, and then make the changes

	// Sort hardware so children are deleted before their parents
	sls.SortHardware(hardwareRemoved)
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
				fmt.Errorf("failed to delete hardware (%s) from SLS", hardware.Xname),
				err,
			)
		}
		log.Info().Int("status", r.StatusCode).Msg("Added hardware to SLS")
	}

	// Update existing hardware
	for _, hardware := range hardwareWithDifferingValues {
		log.Info().Str("xname", hardware.Xname).Msg("Updating")
		// Put into transaction log with old and new value
		// TODO

		// Perform a PUT against SLS
		// TODO
	}

	return nil
}

//
// The following is taken from: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go
//

func displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware []sls_client.Hardware, hardwareWithDifferingValues []sls.GenericHardwarePair) error {
	logFunc := log.Info().Msgf

	logFunc("")
	logFunc("Identical hardware between current and expected states")
	if len(identicalHardware) == 0 {
		logFunc("  None")
	}
	for _, hardware := range identicalHardware {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		logFunc("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	logFunc("")
	logFunc("Common hardware between current and expected states with differing class or extra properties")
	if len(hardwareWithDifferingValues) == 0 {
		log.Info().Msg("  None")
	}
	for _, pair := range hardwareWithDifferingValues {
		logFunc("  %s\n", pair.Xname)

		// Expected Hardware json
		pair.HardwareA.LastUpdated = 0
		pair.HardwareA.LastUpdatedTime = ""
		hardwareRaw, err := buildHardwareString(pair.HardwareA)
		if err != nil {
			return err
		}
		logFunc("  - Expected: %-16s\n", hardwareRaw)

		// Actual Hardware json
		pair.HardwareB.LastUpdated = 0
		pair.HardwareB.LastUpdatedTime = ""
		hardwareRaw, err = buildHardwareString(pair.HardwareB)
		if err != nil {
			return err
		}
		logFunc("  - Actual:   %-16s\n", hardwareRaw)
	}

	logFunc("")
	logFunc("Hardware added to the system")
	if len(hardwareAdded) == 0 {
		logFunc("  None")
	}
	for _, hardware := range hardwareAdded {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		logFunc("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	logFunc("")
	logFunc("Hardware removed from system")
	if len(hardwareRemoved) == 0 {
		logFunc("  None")
	}
	for _, hardware := range hardwareRemoved {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		logFunc("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	logFunc("")
	return nil
}

func buildHardwareString(hardware sls_client.Hardware) (string, error) {
	extraPropertiesRaw, err := hardware.DecodeExtraProperties()
	if err != nil {
		return "", err
	}

	var tokens []string
	tokens = append(tokens, fmt.Sprintf("Type: %s", hardware.TypeString))

	switch hardware.TypeString {
	case xnametypes.Cabinet:
		// Nothing to do
	case xnametypes.CabinetPDUController:
		// Nothing to do
	case xnametypes.RouterBMC:
		// Nothing to do
	case xnametypes.NodeBMC:
		// Nothing to do
	case xnametypes.Node:
		if extraProperties, ok := extraPropertiesRaw.(sls_common.ComptypeNode); ok {
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
		if extraProperties, ok := extraPropertiesRaw.(sls_common.ComptypeMgmtSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4Addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4Addr: %s", extraProperties.IP4Addr))
			}
			if extraProperties.IP6Addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6Addr: %s", extraProperties.IP6Addr))
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
		if extraProperties, ok := extraPropertiesRaw.(sls_common.ComptypeMgmtHLSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4Addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4Addr: %s", extraProperties.IP4Addr))
			}
			if extraProperties.IP6Addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6Addr: %s", extraProperties.IP6Addr))
			}
		}
	case xnametypes.MgmtSwitchConnector:
		if extraProperties, ok := extraPropertiesRaw.(sls_common.ComptypeMgmtSwitchConnector); ok {
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
