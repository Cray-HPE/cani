package csm

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// xnameRe matches xname strings like x8000, x8000c0, x8000c0s0, etc.
var xnameRe = regexp.MustCompile(`^x(\d+)(?:c(\d+))?(?:s(\d+))?`)

// parseXnameParts extracts cabinet, chassis, and slot numbers from an
// xname string. Returns (0, 0, 0) for unrecognised formats.
func parseXnameParts(xn string) (cabinet, chassis, slot int) {
	m := xnameRe.FindStringSubmatch(xn)
	if m == nil {
		return 0, 0, 0
	}
	cabinet, _ = strconv.Atoi(m[1])
	if m[2] != "" {
		chassis, _ = strconv.Atoi(m[2])
	}
	if m[3] != "" {
		slot, _ = strconv.Atoi(m[3])
	}
	return cabinet, chassis, slot
}

// DescribeStagedDevice implements provider.StagedDeviceDescriber.
// It returns Cabinet/Chassis/Blade lines derived from the device's CSM
// xname metadata, or nil when no xname is present.
func (p *Csm) DescribeStagedDevice(device *devicetypes.CaniDeviceType) []string {
	sub, ok := device.GetProviderSubMap(p.Slug())
	if !ok {
		return nil
	}
	xn, _ := sub["xname"].(string)
	if xn == "" {
		return nil
	}
	cab, chas, slot := parseXnameParts(xn)
	return []string{
		fmt.Sprintf("Cabinet: %d", cab),
		fmt.Sprintf("Chassis: %d", chas),
		fmt.Sprintf("Blade: %d", slot),
	}
}
