package csminv

import (
	"fmt"
	"log"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
)

func (ci *CSMInventory) ValidateChassisXname(slsCabinet sls_common.GenericHardware, chassis xnames.Chassis) {
	// Verify the chassis exists
	switch slsCabinet.Class {
	case sls_common.ClassRiver:
		// All hardware in a river chassis is present in chassis 0
		if chassis.Chassis != 0 {
			log.Fatalf("River nodes within a River cabinet must be located in chassis 0. Chassis %d was specified instead.", chassis.Chassis)
		}
	case sls_common.ClassHill:
		fallthrough
	case sls_common.ClassMountain:
		// Verify the chassis exists in SLS
		slsChassis, err := ci.SLSClient.GetHardware(ci.Ctx, chassis.String())
		if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for chassis (%s): %s", chassis.String(), err)
		}

		if slsChassis.Class == sls_common.ClassRiver {
			log.Fatalf("Found river chassis in a hill/mountain cabinet: %s", chassis.String())
		}

		// TODO EX2500 river hardware
		// - Verify cabinet model
		// - Verify that only one liquid cooled chassis exists
	default:
	}
}

// TODO this should probably go into a different file
func GetMgmtSwitchConnectorVendorName(mgmtSwitchEP sls_common.ComptypeMgmtSwitch, switchPort xnames.MgmtSwitchConnector) string {
	switch mgmtSwitchEP.Brand {
	case "Aruba":
		return fmt.Sprintf("1/1/%d", switchPort.MgmtSwitchConnector)
	case "Dell":
		fallthrough
	default:
		return fmt.Sprintf("ethernet1/1/%d", switchPort.MgmtSwitchConnector)
	}
}
