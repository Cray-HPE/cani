package devicetypes

import (
	"regexp"
)

func Racks() (devices map[string]DeviceType) {
	for _, device := range allDeviceTypes {
		re := regexp.MustCompile(`rack`)
		if re.MatchString(device.Slug) {
			if devices == nil {
				devices = make(map[string]DeviceType)
			}
			devices[device.Slug] = device
		}
	}

	return devices
}
