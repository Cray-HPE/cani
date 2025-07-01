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
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

func TestValidateChassisBmc(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c4", "x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 4, 0)
	// expected results:
	// pass: /Hardware/x1000c4: x1000c4b0 ChassisBMC exists for x1000c4 Chassis
	// pass: /Hardware/x1000c5: x1000c5b0 ChassisBMC exists for x1000c5 Chassis
	// pass: /Hardware/x1000c4b0: x1000c4 Chassis exists for x1000c4b0 ChassisBMC
	// pass: /Hardware/x1000c5b0: x1000c5 Chassis exists for x1000c5b0 ChassisBMC
}

func TestValidateMissmatchedClasses(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c4", "x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassHill, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 0, 4)
	// expected results:
	// fail: /Hardware/x1000c4: The classes do not match, x1000c4b0 ChassisBMC has a class of Hill and x1000c4 Chassis has a class of Mountain
	// fail: /Hardware/x1000c5: The classes do not match, x1000c5b0 ChassisBMC has a class of Hill and x1000c5 Chassis has a class of Mountain
	// fail: /Hardware/x1000c4b0: The classes do not match, x1000c4 Chassis has a class of Mountain and x1000c4b0 ChassisBMC has a class of Hill
	// fail: /Hardware/x1000c5b0: The classes do not match, x1000c5 Chassis has a class of Mountain and x1000c5b0 ChassisBMC has a class of Hill

}

func TestValidateMissingChassisBmc(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c4", "x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 2, 1)
	// expected results:
	// pass: /Hardware/x1000c4: x1000c4b0 ChassisBMC exists for x1000c4 Chassis
	// fail: /Hardware/x1000c5: The Hardware entry x1000c5b0 ChassisBMC is missing, every Chassis needs a corresponding ChassisBMC
	// pass: /Hardware/x1000c4b0: x1000c4 Chassis exists for x1000c4b0 ChassisBMC

}

func TestValidateMissingChassis(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassHill, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassHill, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 2, 1)
	// expected results:
	// pass: /Hardware/x1000c5: x1000c5b0 ChassisBMC exists for x1000c5 Chassis
	// fail: /Hardware/x1000c4b0: The Hardware entry x1000c4 Chassis is missing, every ChassisBMC needs a corresponding Chassis.
	// pass: /Hardware/x1000c5b0: x1000c5 Chassis exists for x1000c5b0 ChassisBMC
}

func TestValidateChassisXnameOfWrongType(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	wrongChassisXnames := []string{"x1000c4"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisBmcXnames...)
	addCabinet(hardware, typeToHardware, sls_client.HardwareClassMountain, wrongChassisXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 2, 1)
	// expected results:
	// pass: /Hardware/x1000c5: x1000c5b0 ChassisBMC exists for x1000c5 Chassis
	// fail: /Hardware/x1000c4b0: x1000c4 should be of type Chassis, but is instead of type Cabinet
	// pass: /Hardware/x1000c5b0: x1000c5 Chassis exists for x1000c5b0 ChassisBMC
}

func TestValidateChassisBmcXnameOfWrongType(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c5", "x1000c4"}
	chassisBmcXnames := []string{"x1000c4b0"}
	wrongChassisBmcXnames := []string{"x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisBmcXnames...)
	addCabinet(hardware, typeToHardware, sls_client.HardwareClassMountain, wrongChassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 2, 1)
	// expected results:
	// fail: /Hardware/x1000c5: x1000c5b0 should be of type ChassisBMC, but is instead of type Cabinet
	// pass: /Hardware/x1000c4: x1000c4b0 ChassisBMC exists for x1000c4 Chassis
	// pass: /Hardware/x1000c4b0: x1000c4 Chassis exists for x1000c4b0 ChassisBMC
}

func TestValidateRiver(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c4", "x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassRiver, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassRiver, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	// River hardware is not checked. There should no passes or failures
	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 0, 0)
}

func TestValidateBadChassisXname(t *testing.T) {
	hardware := make(map[string]sls_client.Hardware)
	typeToHardware := make(map[string][]*sls_client.Hardware)
	chassisXnames := []string{"x1000c4b0", "x1000c5"}
	chassisBmcXnames := []string{"x1000c4b0", "x1000c5b0"}
	addChassis(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisXnames...)
	addChassisBmc(hardware, typeToHardware, sls_client.HardwareClassMountain, chassisBmcXnames...)

	checker := NewHardwareChassisBmcCheck(hardware, typeToHardware)
	results := common.NewValidationResults()

	checker.Validate(results)

	reportResultsForChassisBmcCheck(t, chassisXnames, chassisBmcXnames, results, 2, 2)
	// fail: /Hardware/x1000c4b0: The Xname x1000c4b0 is for the type ChassisBMC but it instead has the TypeString Chassis
	// pass: /Hardware/x1000c5: x1000c5b0 ChassisBMC exists for x1000c5 Chassis
	// fail: /Hardware/x1000c4b0: The Hardware entry x1000c4 Chassis is missing, every ChassisBMC needs a corresponding Chassis
	// pass: /Hardware/x1000c5b0: x1000c5 Chassis exists for x1000c5b0 ChassisBMC
}

func addChassis(
	hardwareMap map[string]sls_client.Hardware,
	typeToHardware map[string][]*sls_client.Hardware,
	class sls_client.HardwareClass,
	xnames ...string) {
	addHardware(
		hardwareMap,
		typeToHardware,
		class,
		sls_client.CHASSIS_HardwareType,
		xnametypes.Chassis,
		xnames...)
}

func addChassisBmc(
	hardwareMap map[string]sls_client.Hardware,
	typeToHardware map[string][]*sls_client.Hardware,
	class sls_client.HardwareClass,
	xnames ...string) {
	addHardware(
		hardwareMap,
		typeToHardware,
		class,
		sls_client.CHASSIS_BMC_HardwareType,
		xnametypes.ChassisBMC,
		xnames...)
}

func addCabinet(
	hardwareMap map[string]sls_client.Hardware,
	typeToHardware map[string][]*sls_client.Hardware,
	class sls_client.HardwareClass,
	xnames ...string) {
	addHardware(
		hardwareMap,
		typeToHardware,
		class,
		sls_client.CABINET_HardwareType,
		xnametypes.Cabinet,
		xnames...)
}

func addHardware(
	hardwareMap map[string]sls_client.Hardware,
	typeToHardware map[string][]*sls_client.Hardware,
	class sls_client.HardwareClass,
	t sls_client.HardwareType,
	typeString xnametypes.HMSType,
	xnames ...string) {

	for _, xname := range xnames {
		h := sls_client.Hardware{
			Xname:      xname,
			Class:      class,
			Type:       t,
			TypeString: typeString,
		}
		hardwareMap[xname] = h

		typeToHardware[typeString.String()] = append(typeToHardware[typeString.String()], &h)
	}
}

func reportResultsForChassisBmcCheck(t *testing.T, chassisXnames []string, chassisBmcXnames []string, results *common.ValidationResults, expectedPass int, expectedFail int) {
	passCount, warnCount, failCount := resultsCount(results.GetResults())
	if failCount != expectedFail {
		t.Errorf("Expected %d failures for chassis xnames %v and chassis bmc xnames %v, pass: %d, warn: %d, fail: %d, results:\n%s",
			expectedFail, chassisXnames, chassisBmcXnames, passCount, warnCount, failCount, results.ToString())
	}
	if passCount != expectedPass {
		t.Errorf("Expected %d passing results for chassis xnames %v and chassis bmc xnames %v, pass: %d, warn: %d, fail: %d, results:\n%s",
			expectedPass, chassisXnames, chassisBmcXnames, passCount, warnCount, failCount, results.ToString())
	}
}
