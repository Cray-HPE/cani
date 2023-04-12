/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"net"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

// TransformSlsExtract transforms the SLS extract into the new Inventory
func (i *Inventory) TransformSlsExtract() (err error) {
	var names []string
	for _, s := range i.Extract.SlsConfig.Hardware {
		// Create a new Hardware struct for each SLS Hardware
		hw := Hardware{}

		// The SLS type becomes the Type in the new Inventory
		hw.Type = s.Type.String()
		names, err = GetCommonNames(s)
		if err != nil {
			return err
		}
		hw.Names = names

		// The SLS brand becomes the Manufacturer in the new Inventory
		brand, err := GetBrand(s)
		if err != nil {
			return err
		}
		if brand != "" {
			hw.Manufacturer = brand
		}

		// The SLS IP becomes the IP in the new Inventory
		ip, err := GetIPAddress(s)
		if err != nil {
			return err
		}
		if ip != nil {
			hw.IP = ip
		}

		// The SLS IP becomes the IP in the new Inventory
		class, err := GetClass(s)
		if err != nil {
			return err
		}
		hw.Class = class

		// hw.Model = ""
		// The GUID comes from Redfish (or is generated)
		// hw.GUID = ""
		// fmt.Printf("FROM SLS: %s %+v\n", i, hw)
	}
	return nil
}

// TransformCanuExtract extracts a CANU config and transforms it into a Hardware struct
func (i *Inventory) TransformCanuExtract() (err error) {
	for _, canu := range i.Extract.CanuConfig.Topology {
		// Create a new Hardware struct for each CANU Hardware
		hw := Hardware{}
		hw.Architechture = canu.Architecture
		hw.Manufacturer = canu.Vendor
		hw.Vendor = canu.Vendor
	}
	return nil
}

// TransformCsmExtract extracts a CSI config and transforms it into a Hardware struct
func (i *Inventory) TransformCsiExtract() (err error) {
	// Create a new Hardware struct for each SLS Hardware
	hw := Hardware{}
	hw.CsmVersion = i.Extract.CsiConfig.CsmVersion
	if i.Extract.CsiConfig.SiteDNS != "" {
		hw.Networking.SiteDNS = []net.IP{net.ParseIP(i.Extract.CsiConfig.SiteDNS)}
	}
	if i.Extract.CsiConfig.SiteDomain != "" {
		hw.Networking.SiteDomain = i.Extract.CsiConfig.SiteDomain
	}

	if i.Extract.CsiConfig.CanGateway != "" {
		hw.Networking.CanGW = net.ParseIP(i.Extract.CsiConfig.CanGateway)
	}

	if i.Extract.CsiConfig.SiteIP != "" {
		hw.Networking.SiteIP = net.ParseIP(i.Extract.CsiConfig.SiteIP)
	}
	return nil
}

// GetBrand extracts the brand from the ExtraPropertiesRaw interface from SLS
func GetBrand(s sls_common.GenericHardware) (brand string, err error) {
	if s.ExtraPropertiesRaw != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := s.ExtraPropertiesRaw.(map[string]interface{}); !ok {
			return "", errors.New("Type assertion error: getting brand")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := s.ExtraPropertiesRaw.(map[string]interface{})
		// If there is a brand,

		if ep["Brand"] != nil {
			// type assert again that this is an interface
			if _, ok := ep["Brand"].(interface{}).(string); !ok {
				return "", errors.New("Type assertion error: Brand")
			}
			brand = ep["Brand"].(interface{}).(string)
		}
	}
	return brand, nil
}

// GetCommonNames transforms the ExtraPropertiesRaw interface from SLS into a slice of strings
func GetCommonNames(s sls_common.GenericHardware) (names []string, err error) {
	if s.ExtraPropertiesRaw != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := s.ExtraPropertiesRaw.(map[string]interface{}); !ok {
			return []string{}, errors.New("Type assertion error: getting common names")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := s.ExtraPropertiesRaw.(map[string]interface{})
		// If there are aliases,
		if ep["Aliases"] != nil {
			// type assert again that this is an interface
			if _, ok := ep["Aliases"].([]interface{}); !ok {
				return []string{}, errors.New("Type assertion error: getting aliases")
			}
			// Append all aliases to the slice, type asserting each one and converting to a string
			for _, alias := range ep["Aliases"].([]interface{}) {
				// fmt.Println("sttt", alias.(string))
				name := alias.(string)
				names = append(names, name)
			}
		}
	}
	return names, nil
}

// GetIPAddress transforms the ExtraPropertiesRaw interface from SLS into a net.IP
func GetIPAddress(s sls_common.GenericHardware) (ip net.IP, err error) {
	if s.ExtraPropertiesRaw != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := s.ExtraPropertiesRaw.(map[string]interface{}); !ok {
			return nil, errors.New("Type assertion error: ExtraProperties")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := s.ExtraPropertiesRaw.(map[string]interface{})
		if ep["IP4addr"] != nil {
			// type assert that it is a string
			if _, ok := ep["IP4addr"].(string); !ok {
				return nil, errors.New("Type assertion error: IP4addr")
			}
			// set an easier-to-use variable
			ip = net.ParseIP(ep["IP4addr"].(string))
		}
	}
	// Convert the string to a net.IPAddr
	return ip, nil
}

// GetClass transforms the ExtraPropertiesRaw interface from SLS into a string
func GetClass(s sls_common.GenericHardware) (class string, err error) {
	if s.Class != "" {
		class = string(s.GetClass())
	}
	return class, nil
}
