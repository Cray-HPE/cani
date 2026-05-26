package import_

import (
	"fmt"
	"net"
	"strings"
)

// ValidateSlsDumpstate performs basic sanity checks on the SLS data.
// It returns a non-nil error if any check fails.
func ValidateSlsDumpstate(sls *SlsDumpstate) error {
	if sls == nil {
		return fmt.Errorf("SLS dumpstate is nil")
	}

	var errs []string

	// Check hardware parent-child consistency.
	for xname, hw := range sls.Hardware {
		if hw.Parent == "" {
			continue
		}
		if _, ok := sls.Hardware[hw.Parent]; !ok {
			errs = append(errs, fmt.Sprintf(
				"hardware %s references parent %s which does not exist",
				xname, hw.Parent))
		}
	}

	// Check network DHCP ranges.
	for netName, network := range sls.Networks {
		if network.ExtraProperties == nil {
			continue
		}
		for _, subnet := range network.ExtraProperties.Subnets {
			if err := validateDHCPRange(subnet); err != nil {
				errs = append(errs, fmt.Sprintf(
					"network %s subnet %s: %s",
					netName, subnet.Name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("SLS validation failed:\n  %s",
			strings.Join(errs, "\n  "))
	}
	return nil
}

// validateDHCPRange checks that DHCPStart <= DHCPEnd when both are set.
func validateDHCPRange(s SlsSubnet) error {
	if s.DHCPStart == "" || s.DHCPEnd == "" {
		return nil
	}
	start := net.ParseIP(s.DHCPStart)
	end := net.ParseIP(s.DHCPEnd)
	if start == nil {
		return fmt.Errorf("invalid DHCPStart %q", s.DHCPStart)
	}
	if end == nil {
		return fmt.Errorf("invalid DHCPEnd %q", s.DHCPEnd)
	}
	// Normalise to 16-byte form for byte comparison.
	start = start.To16()
	end = end.To16()
	for i := range start {
		if start[i] < end[i] {
			return nil
		}
		if start[i] > end[i] {
			return fmt.Errorf(
				"DHCPStart %s is greater than DHCPEnd %s",
				s.DHCPStart, s.DHCPEnd)
		}
	}
	return nil // equal is fine
}
