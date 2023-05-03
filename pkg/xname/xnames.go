package xname

import (
	"regexp"
	"strconv"
	"unicode"
)

type Xname string

type Context struct {
	Letters string
	Type    string
	Regexp  *regexp.Regexp
}

// IsValid returns true if the xname is valid, false otherwise.
// iterates through the input string and checks if each character is a letter.
// If it encounters a letter, it proceeds to the next character and starts looking for digits.
// If it finds a digit, it continues to the next character until it reaches a non-digit character or the end of the string.
// Then, it converts the consecutive digit characters into an integer using strconv.Atoi().
// If the conversion is successful, it continues with the next pair; otherwise, it returns false.
// If all pairs are valid, the method returns true.
func (x Xname) IsValid() bool {
	xname := string(x)

	i := 0
	// iterate through each character in the input string
	for i < len(xname) {
		// check if each character is a letter
		if unicode.IsLetter(rune(xname[i])) {
			// if a letter is encountered,
			i++
			// proceed to the next character and starts looking for digits.
			start := i
			// if a digit is found, continue to the next character until it reaches a non-digit character or the end of the string
			for i < len(xname) && unicode.IsDigit(rune(xname[i])) {
				i++
			}
			// Then, convert the consecutive digit characters into an integer using strconv.Atoi().
			if i > start {
				_, err := strconv.Atoi(xname[start:i])
				if err != nil {
					return false
				}
				// If the conversion is successful, continue with the next pair; otherwise, it returns false
			} else {
				return false
			}
		} else {
			return false
		}
	}
	// If all pairs are valid, return true
	return true
}

func (x Xname) Type() Context {
	for _, c := range definitionSets {
		if c.Regexp.MatchString(string(x)) {
			return c
		}
	}
	return Context{}
}

// systemDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func systemDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^s\\d+$"), Type: "System"}
}

// cduDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cduDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^d\\d+$"), Type: "CDU"}
}

// cduMgmtSwitchDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cduMgmtSwitchDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^d\\d+w\\d+$"), Type: "CDUMgmtSwitch"}
}

// cabinetDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+$"), Type: "Cabinet"}
}

// cecDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cecDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+e\\d+$"), Type: "CEC"}
}

// cabinetBmcDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetBmcDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+b\\d*$"), Type: "CabBMC"}
}

// cabinetCduDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetCduDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+d\\d*$"), Type: "CabinetCDU"}
}

// cabinetPduControllerDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetPduControllerDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+m\\d*$"), Type: "CabinetPDUController"}
}

// cabinetPduDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetPduDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+m\\d+p\\d*$"), Type: "CabinetPDU"}
}

// cabinetPduNicDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetPduNicDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+m\\d+p\\d+i\\d*$"), Type: "CabinetPDUNic"}
}

// cabinetPduOutletDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetPduOutletDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+m\\d+p\\d+j\\d*$"), Type: "CabinetPDUOutlet"}
}

// cabinetPduPowerConnectorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func cabinetPduPowerConnectorDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+m\\d+p\\d+v\\d*$"), Type: "CabinetPDUPowerConnector"}
}

// chassisDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d*$"), Type: "Chassis"}
}

// chassisCmmFpgaDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisCmmFpgaDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+f\\d*$"), Type: "ChassisCMMFPGA"}
}

// chassisCmmRectifierfinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisCmmRectifierfinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+t\\d*$"), Type: "ChassisCMMRectifier"}
}

// chassisBmcfinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisBmcDefinitions() Context {
	return Context{Letters: "xXcCbB", Regexp: regexp.MustCompile("^x\\d+c\\d+b\\d*$"), Type: "ChassisBMC"}
}

// chassisBmcNicfinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisBmcNicDefinitions() Context {
	return Context{Letters: "xXcCiI", Regexp: regexp.MustCompile("^x\\d+c\\d+b\\d+i\\d*$"), Type: "NodeBMCNic"}
}

// computeModuleDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func computeModuleDefinitions() Context {
	return Context{Letters: "xXcCsS", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+$"), Type: "ComputeModule"}
}

// nodePowerConnectorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func nodePowerConnectorDefinitions() Context {
	return Context{Letters: "xXcCsSvV", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+v\\d*$"), Type: "NodePowerConnector"}
}

// nodeBmcDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func nodeBmcDefinitions() Context {
	return Context{Letters: "xXcCsSbB", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d*$"), Type: "NodeBMC"}
}

// chassisNodeBmcNicDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeBmcNicDefinitions() Context {
	return Context{Letters: "xXcCsSbBiI", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+i\\d*$"), Type: "NodeBMCNic"}
}

// chassisNodeDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func nodeDefinitions() Context {
	return Context{Letters: "xXcCsSbBnN", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d*$"), Type: "Node"}
}

// chassisNodeMemoryDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func nodeMemoryDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+d\\d*$"), Type: "NodeMemory"}
}

// chassisNodeAccelDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeAccelDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+a\\d*$"), Type: "NodeAccel"}
}

// chassisNodeAccelRiserDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeAccelRiserDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+r\\d*$"), Type: "NodeAccelRiser"}
}

// chassisNodeHsnNicDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeHsnNicDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+h\\d*$"), Type: "NodeHSNNic"}
}

// chassisNodeProcessorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeProcessorDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+p\\d*$"), Type: "NodeProcessor"}
}

// chassisNodeStorageGroupDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeStorageGroupDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+g\\d*$"), Type: "NodeStorageGroup"}
}

// chassisNodeDriveDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeDriveDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+n\\d+g\\d+k\\d*$"), Type: "NodeDrive"}
}

// chassisNodeEnclosureDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeEnclosureDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+e\\d*$"), Type: "NodeEnclosure"}
}

// chassisNodeEnclosurePowerSupplyDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeEnclosurePowerSupplyDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+e\\d+t\\d*$"), Type: "NodeEnclosurePowerSupply"}
}

// chassisNodeEnclosureFpgaDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisNodeEnclosureFpgaDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+b\\d+e\\d+f\\d*$"), Type: "NodeEnclosureFpga"}
}

// chassisMgmtHlSwitchEnclosureDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisMgmtHlSwitchEnclosureDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+h\\d*$"), Type: "MgmtHlSwitchEnclosure"}
}

// chassisMgmtHlSwitchDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisMgmtHlSwitchDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+h\\d*$"), Type: "MgmtHlSwitch"}
}

// chassisMgmtSwitchDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisMgmtSwitchDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+w\\d*$"), Type: "MgmtSwitch"}
}

// chassisMgmtSwitchConnectorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisMgmtSwitchConnectorDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+s\\d+w\\d+j\\d*$"), Type: "MgmtSwitchConnector"}
}

// chassisRouterModuleDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterModuleDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d*$"), Type: "RouterModule"}
}

// chassisRouterHsnAsicDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterHsnAsicDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+a\\d*$"), Type: "RouterHsnAsic"}
}

// chassisRouterHsnLinkDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterHsnLinkDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+a\\d+l\\d*$"), Type: "RouterHsnLink"}
}

// chassisRouterHsnBoardDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterHsnBoardDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+a\\d+b\\d*$"), Type: "RouterHsnBoard"}
}

// chassisRouterHsnConnectorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterHsnConnectorDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+a\\d+b\\d+j\\d*$"), Type: "RouterHsnConnector"}
}

// chassisRouterHsnConnectorPortDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterHsnConnectorPortDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+a\\d+b\\d+j\\d+p\\d*$"), Type: "RouterHsnConnectorPort"}
}

// chassisRouterBmcDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterBmcDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+b\\d*$"), Type: "RouterBmc"}
}

// chassisRouterBmcNicDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterBmcNicDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+b\\d+i\\d*$"), Type: "RouterBmcNic"}
}

// chassisRouterFpgaDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterFpgaDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+f\\d*$"), Type: "RouterFpga"}
}

// chassisRouterPowerConnectorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterPowerConnectorDefinitions() Context {
	return Context{
		Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+v\\d*$"),
		Type:   "RouterPowerConnector"}
}

// chassisRouterTorDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterTorDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+t\\d*$"), Type: "RouterTor"}
}

// chassisRouterTorFpgaDefinitions returns a slice of Contexts that define the possible types of each section of an Xname
// It is intended to be used as a constant
func chassisRouterTorFpgaDefinitions() Context {
	return Context{Letters: "", Regexp: regexp.MustCompile("^x\\d+c\\d+r\\d+t\\d+f\\d*$"), Type: "RouterTorFpga"}
}

var (
	definitionSets = []Context{
		systemDefinitions(),
		cduDefinitions(),
		cduMgmtSwitchDefinitions(),
		cabinetDefinitions(),
		cecDefinitions(),
		cabinetBmcDefinitions(),
		cabinetCduDefinitions(),
		cabinetPduControllerDefinitions(),
		cabinetPduDefinitions(),
		cabinetPduOutletDefinitions(),
		cabinetPduPowerConnectorDefinitions(),
		chassisDefinitions(),
		chassisCmmFpgaDefinitions(),
		chassisCmmRectifierfinitions(),
		chassisBmcDefinitions(),
		chassisBmcNicDefinitions(),
		computeModuleDefinitions(),
		nodePowerConnectorDefinitions(),
		nodeBmcDefinitions(),
		chassisNodeBmcNicDefinitions(),
		nodeDefinitions(),
		nodeMemoryDefinitions(),
		chassisNodeAccelDefinitions(),
		chassisNodeAccelRiserDefinitions(),
		chassisNodeHsnNicDefinitions(),
		chassisNodeProcessorDefinitions(),
		chassisNodeStorageGroupDefinitions(),
		chassisNodeDriveDefinitions(),
		chassisNodeEnclosureDefinitions(),
		chassisNodeEnclosurePowerSupplyDefinitions(),
		chassisNodeEnclosureFpgaDefinitions(),
		chassisMgmtHlSwitchEnclosureDefinitions(),
		chassisMgmtHlSwitchDefinitions(),
		chassisMgmtSwitchDefinitions(),
		chassisRouterModuleDefinitions(),
		chassisRouterHsnAsicDefinitions(),
		chassisRouterHsnLinkDefinitions(),
		chassisRouterHsnBoardDefinitions(),
		chassisRouterHsnConnectorDefinitions(),
		chassisRouterHsnConnectorPortDefinitions(),
		chassisRouterBmcDefinitions(),
		chassisRouterBmcNicDefinitions(),
		chassisRouterFpgaDefinitions(),
		chassisRouterPowerConnectorDefinitions(),
		chassisRouterTorDefinitions(),
		chassisRouterTorFpgaDefinitions(),
	}
)

// // ComputeModule returns the xname leading up to the slot (computemodule) if it exists, otherwise it returns an error.
// func (x Xname) ComputeModule() (string, error) {
// 	xname := string(x)
// 	i := 0
// 	for i < len(xname) {
// 		// set a var to check for invalid xnames (an xname of x1000d0s0 is not a slot)
// 		found := false
// 		// For each definition,
// 		for _, ctx := range chassisComputeModuleDefinitions() {
// 			if xname[i] == ctx.Letters[0] {
// 				// if a valid character is found, set found to true
// 				found = true
// 				if ctx.Type == "computemodule" {
// 					i++
// 					// Continue scanning for digits until none are found.
// 					for i < len(xname) && unicode.IsDigit(rune(xname[i])) {
// 						i++
// 					}
// 					return xname[:i], nil
// 				} else {
// 					// If this section is not a slot, continue scanning for digits until none are found.
// 					i++
// 					for i < len(xname) && unicode.IsDigit(rune(xname[i])) {
// 						i++
// 					}
// 					break
// 				}
// 			}
// 		}
// 		// If no match was found, return an error.
// 		if !found {
// 			return "", errors.New("Not a valid ComputeModule Xname: Unexpected character found")
// 		}
// 	}
// 	// If no slot was found, return an error.
// 	return "", errors.New("no ComputeModule found")
// }

// // NodeBmc returns the xname leading up to the NodeBMC if it exists, otherwise it returns an error.
// // If the xname is not a complete NodeBMC, it returns the children using filterSequentialXnames.
// func (x Xname) NodeBmc(allXnames []string) ([]string, error) {
// 	xname := string(x)
// 	i := 0
// 	for i < len(xname) {
// 		// set a var to check for invalid xnames (an xname of x1000d0s0 is not a slot)
// 		found := false
// 		// For each definition,
// 		for _, ctx := range chassisNodeBmcDefinitions() {
// 			if xname[i] == ctx.Letters[0] {
// 				// if a valid character is found, set found to true
// 				found = true
// 				if ctx.Type == "nodebmc" {
// 					i++
// 					// Continue scanning for digits until none are found.
// 					for i < len(xname) && unicode.IsDigit(rune(xname[i])) {
// 						i++
// 					}
// 					return []string{xname[:i]}, nil
// 				} else {
// 					// If this section is not a slot, continue scanning for digits until none are found.
// 					i++
// 					for i < len(xname) && unicode.IsDigit(rune(xname[i])) {
// 						i++
// 					}
// 					break
// 				}
// 			}
// 		}

// 		// If no match was found, return an error.
// 		if !found {
// 			return nil, errors.New("Not a valid NodeBMC Xname: Unexpected character found")
// 		}
// 	}

// 	// If no slot was found, return the children using filterSequentialXnames.
// 	children := filterSequentialXnames(xname, allXnames, chassisNodeDefinitions())
// 	if len(children) == 0 {
// 		return nil, errors.New("No NodeBMC found")
// 	}
// 	return children, nil
// }

// func containsChar(s string, c rune) bool {
// 	for _, ch := range s {
// 		if ch == c {
// 			return true
// 		}
// 	}
// 	return false
// }

// func filterSequentialXnames(prefix string, xnames []string, definitions []Context) []string {
// 	matchingXnames := []string{}

// 	// Find the index of the last definition matched by the prefix
// 	lastMatchedIndex := -1
// 	for i, def := range definitions {
// 		if containsChar(prefix, rune(def.Letters[0])) {
// 			lastMatchedIndex = i
// 		} else {
// 			break
// 		}
// 	}

// 	pattern := "^" + prefix
// 	for i := lastMatchedIndex + 1; i < len(definitions); i++ {
// 		def := definitions[i]
// 		pattern += fmt.Sprintf("[%s][0-9]+", def.Letters)
// 	}
// 	pattern += "$"
// 	regex := regexp.MustCompile(pattern)

// 	for _, xname := range xnames {
// 		if regex.MatchString(xname) {
// 			matchingXnames = append(matchingXnames, xname)
// 		}
// 	}

// 	return matchingXnames
// }

// func CheckXnameType(xname string) (string, bool) {
// 	for i, def := range definitionSets {
// 		if strings.ContainsAny(xname, def[i].Letters) {
// 			// Check that all letters in the name match the definition
// 			var matched bool
// 			var currentType string
// 			for _, c := range xname {
// 				if strings.ContainsRune(def[i].Letters, c) {
// 					matched = true
// 					currentType = def[i].Type
// 				} else if unicode.IsDigit(c) && matched {
// 					// If the previous character was a valid letter, then this digit
// 					// matches the definition and can be ignored
// 				} else {
// 					// If any other character is encountered, the name doesn't match
// 					return "", false
// 				}
// 			}
// 			// Name matches this definition
// 			return currentType, true
// 		}
// 	}
// 	// No matching definition found
// 	return "", false
// }

// func filterSequentialXnames(prefix string, xnames []string) []string {
// 	matchingXnames := []string{}
// 	prefixLength := len(prefix)
// 	for _, xname := range xnames {
// 		if strings.HasPrefix(xname, prefix) && len(xname) == prefixLength+4 {
// 			matchingXnames = append(matchingXnames, xname)
// 		}
// 	}

// 	return matchingXnames
// }

// func main() {
// 	myxname := Xname("x1000c0s0")

// 	valid := myxname.IsValid()
// 	if !valid {
// 		fmt.Println("Invalid Xname")
// 	} else {
// 		fmt.Printf("Valid Xname: %t\n", valid) // Output: Valid: true
// 	}
// 	slot, err := myxname.NodeBmc()
// 	if err != nil {
// 		fmt.Println(fmt.Sprintf("Not a ComputeModule: %v", err))
// 	} else {
// 		fmt.Printf("ComputeModule Xname: %s\n", slot) // Output: Slot: s
// 	}
// }
