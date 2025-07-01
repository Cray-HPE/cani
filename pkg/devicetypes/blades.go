package devicetypes

import (
	"regexp"

	"github.com/google/uuid"
)

func Blades() (devices map[string]DeviceType) {
	for _, device := range allDeviceTypes {
		re := regexp.MustCompile(`node|server|proliant`)
		if re.MatchString(device.Slug) {
			if devices == nil {
				devices = make(map[string]DeviceType)
			}
			devices[device.Slug] = device
		}
	}

	return devices
}

func NewBlade(s string) (device *CaniDeviceType) {
	if d, ok := allDeviceTypes[s]; ok {
		device = &CaniDeviceType{
			ID:   uuid.New(),
			Name: d.Model,
		}
	} else {
		device = &CaniDeviceType{
			ID:   uuid.New(),
			Name: s,
		}
	}

	return device
}
