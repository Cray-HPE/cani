package csm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/rs/zerolog/log"
)

// Reconcile CANI's inventory state with the external inventory state and apply required changes
func (csm *CSM) Reconcile(datastore inventory.Datastore) (err error) {

	log.Info().Msg("Starting CSM reconcile process")

	// TODO grab the allowed HSM Roles and SubRoles from HSM
	// This is for data validation

	// Retrieve the current SLS state
	// TODO
	currentSLSState := sls_common.SLSState{}

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

	// Modify SLS
	// TODO

	return nil
}

// displayHardwareComparisonReport displays a report of the hardware comparison
// taken from: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go
func displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware []sls_common.GenericHardware, hardwareWithDifferingValues []sls.GenericHardwarePair) error {
	log.Info().Msg("")
	log.Info().Msg("Identical hardware between current and expected states")
	if len(identicalHardware) == 0 {
		log.Info().Msg("  None")
	}
	for _, hardware := range identicalHardware {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	log.Info().Msg("")
	log.Info().Msg("Common hardware between current and expected states with differing class or extra properties")
	if len(hardwareWithDifferingValues) == 0 {
		log.Info().Msg("  None")
	}
	for _, pair := range hardwareWithDifferingValues {
		log.Info().Msgf("  %s\n", pair.Xname)

		// Expected Hardware json
		pair.HardwareA.LastUpdated = 0
		pair.HardwareA.LastUpdatedTime = ""
		hardwareRaw, err := buildHardwareString(pair.HardwareA)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Expected: %-16s\n", hardwareRaw)

		// Actual Hardware json
		pair.HardwareB.LastUpdated = 0
		pair.HardwareB.LastUpdatedTime = ""
		hardwareRaw, err = buildHardwareString(pair.HardwareB)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Actual:   %-16s\n", hardwareRaw)
	}

	log.Info().Msg("")
	log.Info().Msg("Hardware added to the system")
	if len(hardwareAdded) == 0 {
		log.Info().Msg("  None")
	}
	for _, hardware := range hardwareAdded {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	log.Info().Msg("")
	log.Info().Msg("Hardware removed from system")
	if len(hardwareRemoved) == 0 {
		log.Info().Msg("  None")
	}
	for _, hardware := range hardwareRemoved {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s\n", hardware.Xname, hardwareRaw)
	}

	log.Info().Msg("")
	return nil
}

// buildHardwareString builds a string representation of the hardware
func buildHardwareString(hardware sls_common.GenericHardware) (string, error) {
	extraPropertiesRaw, err := sls.DecodeHardwareExtraProperties(hardware)
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
