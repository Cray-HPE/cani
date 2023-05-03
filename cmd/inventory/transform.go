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
package inventory

import (
	"errors"
	"net"
	"reflect"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/rs/zerolog/log"
)

// TransformExtraProperties transforms the SLS ExtraPropertiesRaw more useable structures in the 'Hardware' type
func TransformExtraProperties(epr interface{}) (ep SlsExtraProperties, err error) {
	// The SLS common names become the Names in the new Inventory
	names, err := GetCommonNames(epr)
	if err != nil {
		return ep, err
	}
	if len(names) == 0 {
		ep.Names = names
	}

	// The SLS brand becomes the Manufacturer in the new Inventory
	brand, err := GetBrand(epr)
	if err != nil {
		return ep, err
	}
	if brand != "" {
		ep.Manufacturer = brand
	}

	// The SLS IP becomes the IP in the new Inventory
	ip, err := GetIPAddress(epr)
	if err != nil {
		return ep, err
	}
	if ip != nil {
		ep.IP = ip
	}

	// The SLS IP becomes the IP in the new Inventory
	class, err := GetClass(epr)
	if err != nil {
		return ep, err
	}
	if class != "" {
		ep.Class = class
	}

	return ep, nil
}

// TransformCabinets transforms the SLS Cabinets into a 'Hardware' type
func (s *Hardware) TransformAllHw() (err error) {
	for _, hw := range s.Extract.SlsConfig.Hardware {
		// Create a new Cabinet struct for each SLS Cabinet
		hardware := Hardware{}

		// Transform SLS values to their respective fields
		hardware.Xname = xnames.FromString(hw.Xname)
		// cabinet.Class = string(sls_client)
		hardware.Type = string(hw.TypeString)
	}
	return nil
}

// TransformCabinets transforms the SLS Cabinets into a 'Hardware' type
func (s *Hardware) TransformCabinets() (err error) {
	for _, hw := range s.Extract.SlsConfig.Hardware {
		if hw.TypeString == string(hsm_client.CABINET_HmsType100) {
			// Create a new Cabinet struct for each SLS Cabinet
			cabinet := Cabinet{}

			// Transform SLS values to their respective fields
			cabinet.Xname = xnames.FromString(hw.Xname)
			// cabinet.Class = string(sls_client)
			cabinet.Type = string(hw.TypeString)
			// append the cabinet to the system's cabinets
			s.Cabinets = append(s.Cabinets, cabinet)
		}
	}
	return nil
}

// TransformChassis transforms the SLS Chassis into a 'Hardware' type
func (s *Hardware) TransformChassis() (err error) {
	log.Debug().Msgf("Transforming SLS Chassis")
	for _, hw := range s.Extract.SlsConfig.Hardware {
		if hw.TypeString == sls_common.Chassis.String() {
			// Create a new chassis struct for each SLS chassis
			chassis := Chassis{}
			// Transform SLS values to their respective fields
			chassis.Xname = xnames.FromString(hw.Xname)
			log.Debug().Msgf("Chassis xname set to %s from %s", chassis.Xname, hw.Xname)
			// Check if the chassis is in a cabinet
			parent := chassis.Xname.ParentInterface()
			log.Debug().Msgf("Chassis parent is %+v %s", reflect.TypeOf(parent), parent)
			// If the parent is a cabinet, then add the chassis to the cabinet
			if XnameInSlice(parent, s.Cabinets) {
				log.Debug().Msgf("Parent %s exists in current system", parent)
				for _, cabinet := range s.Cabinets {
					if cabinet.Xname == parent {
						cabinet.Chassis = append(cabinet.Chassis, chassis)
						log.Debug().Msgf("Added %s as a chassis in %s cabinet", chassis.Xname, cabinet.Xname)
					}
				}
			} else {
				return errors.New("Chassis " + chassis.Xname.String() + " is not in a cabinet")
			}
			chassis.Type = string(hw.TypeString)
			// Append the chassis to the list of chassis
			s.Chassis = append(s.Chassis, chassis)
			log.Debug().Msgf("Added %s to the []Chassis", chassis.Xname)
		}
	}
	return nil
}

// TransformComputeModule transforms the SLS ComputeModule into a 'Hardware' type
func (s *Hardware) TransformComputeModule() (err error) {
	for _, hw := range s.Extract.SlsConfig.Hardware {
		if hw.Type_ == string(hsm_client.COMPUTE_MODULE_HmsType100) {
			// Create a new hardware struct
			cm := Blade{}
			// Transform SLS values to their respective fields
			cm.Xname = xnames.FromString(hw.Xname)
			log.Debug().Msgf("ComputeModule xname set to %s from %s", cm.Xname, hw.Xname)
			parent := cm.Xname.ParentInterface()
			log.Debug().Msgf("ComputeModule parent is %+v %s", reflect.TypeOf(parent), parent)

			// If the parent is a chassis, then add the cm to the chassis
			if XnameInSlice(parent, s.Chassis) {
				for _, chassis := range s.Chassis {
					if chassis.Xname == parent {
						s.Blades = append(chassis.Blades, cm)
						log.Debug().Msgf("Added %s as a compute module in %s chassis", cm.Xname, chassis.Xname)
					}
				}

				cm.Type = string(hw.TypeString)

				// append the slot to the list of slots
				s.Blades = append(s.Blades, cm)
				log.Debug().Msgf("Added %s to the []Blades", cm.Xname)
			}
		}
	}

	return nil
}

// // TransformNodeBmc transforms the SLS NodeBMC into a 'Hardware' type
// func (s *Hardware) TransformNodeBmc() (err error) {
// 	for _, hw := range s.Extract.SlsConfig.Hardware {
// 		if hw.Type_ == string(hsm_client.COMPUTE_MODULE_HmsType100) {
// 			// Create a new hardware struct
// 			cn := Hardware{}
// 			// Transform SLS values to their respective fields
// 			cn.Xname = xnames.FromString(hw.Xname)
// 			log.Debug().Msgf("NodeBMC xname set to %s from %s", cn.Xname, hw.Xname)
// 			parent := cn.Xname.ParentInterface()
// 			log.Debug().Msgf("NodeBMC parent is %+v %s", reflect.TypeOf(parent), parent)

// 			// If the parent is a chassis, then add the cm to the chassis
// 			if XnameInSlice(parent, s.Blades) {
// 				for _, blade := range s.Blades {
// 					if blade.Xname == parent {
// 						s.Blades = append(blade.Blades, cn)
// 						log.Debug().Msgf("Added %s as a nodebmc in %s blade", cn.Xname, blade.Xname)
// 					}
// 				}

// 				cn.Type = string(hw.TypeString)

// 				// append the slot to the list of slots
// 				s.Blades = append(s.Blades, cn)
// 				log.Debug().Msgf("Added %s to the []Blades", cn.Xname)
// 			}
// 		}
// 	}

// 	return nil
// }

// TransformSlsExtract transforms the SLS extract into the new Inventory
func (s *Hardware) TransformSlsExtract() (err error) {
	err = s.TransformCabinets()
	if err != nil {
		return err
	}
	err = s.TransformChassis()
	if err != nil {
		return err
	}
	err = s.TransformComputeModule()
	if err != nil {
		return err
	}
	// err = s.TransformNodeBmc()
	// if err != nil {
	// 	return err
	// }
	return nil
}

// TransformCanuExtract extracts a CANU config and transforms it into a Hardware struct
func (i *Hardware) TransformCanuExtract() (err error) {
	// for _, canu := range i.Extract.CanuConfig.Topology {
	// 	// Create a new Hardware struct for each CANU Hardware
	// 	hw := Hardware{}
	// 	hw.Architechture = canu.Architecture
	// 	hw.Manufacturer = canu.Vendor
	// 	hw.Vendor = canu.Vendor
	// }
	return nil
}

// TransformCsmExtract extracts a CSI config and transforms it into a Hardware struct
func (i *Hardware) TransformCsiExtract() (err error) {
	// Create a new Hardware struct for each SLS Hardware
	// hw := Hardware{}
	// hw.CsmVersion = i.Extract.CsiConfig.CsmVersion
	// if i.Extract.CsiConfig.SiteDNS != "" {
	// 	hw.Networking.SiteDNS = []net.IP{net.ParseIP(i.Extract.CsiConfig.SiteDNS)}
	// }
	// if i.Extract.CsiConfig.SiteDomain != "" {
	// 	hw.Networking.SiteDomain = i.Extract.CsiConfig.SiteDomain
	// }

	// if i.Extract.CsiConfig.CanGateway != "" {
	// 	hw.Networking.CanGW = net.ParseIP(i.Extract.CsiConfig.CanGateway)
	// }

	// if i.Extract.CsiConfig.SiteIP != "" {
	// 	hw.Networking.SiteIP = net.ParseIP(i.Extract.CsiConfig.SiteIP)
	// }
	return nil
}

// GetBrand extracts the brand from the ExtraPropertiesRaw interface from SLS
func GetBrand(epr interface{}) (brand string, err error) {
	if epr != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := epr.(map[string]interface{}); !ok {
			return "", errors.New("Type assertion error: getting brand")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := epr.(map[string]interface{})
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
func GetCommonNames(epr interface{}) (names []string, err error) {
	if epr != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := epr.(map[string]interface{}); !ok {
			return []string{}, errors.New("Type assertion error: getting common names")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := epr.(map[string]interface{})
		// If there are aliases,
		if ep["Aliases"] != nil {
			// type assert again that this is an interface
			if _, ok := ep["Aliases"].([]interface{}); !ok {
				return []string{}, errors.New("Type assertion error: getting aliases")
			}
			// Append all aliases to the slice, type asserting each one and converting to a string
			for _, alias := range ep["Aliases"].([]interface{}) {
				name := alias.(string)
				names = append(names, name)
			}
		}
	}
	return names, nil
}

// GetIPAddress transforms the ExtraPropertiesRaw interface from SLS into a net.IP
func GetIPAddress(epr interface{}) (ip net.IP, err error) {
	if epr != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := epr.(map[string]interface{}); !ok {
			return nil, errors.New("Type assertion error: ExtraProperties")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := epr.(map[string]interface{})
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
func GetClass(epr interface{}) (class string, err error) {
	if epr != nil {
		// type assert the ExtraPropertiesRaw interface and error if it fails
		if _, ok := epr.(map[string]interface{}); !ok {
			return "", errors.New("Type assertion error: ExtraProperties")
		}
		// set an easier-to-use variable for "ExtraProperties"
		ep := epr.(map[string]interface{})
		if ep["Class"] != nil {
			// type assert that it is a string
			if _, ok := ep["Class"].(string); !ok {
				return "", errors.New("Type assertion error: Class")
			}
			// set an easier-to-use variable
			class = ep["Class"].(string)
		}
	}

	return class, nil
}
