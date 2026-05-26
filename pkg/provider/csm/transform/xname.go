package transform

import (
	"regexp"
	"strconv"
	"unicode"
)

// XnameType constants matching SLS TypeString values.
const (
	XnameTypeSystem               = "System"
	XnameTypeCabinet              = "Cabinet"
	XnameTypeChassis              = "Chassis"
	XnameTypeChassisBMC           = "ChassisBMC"
	XnameTypeComputeModule        = "ComputeModule"
	XnameTypeNodeBMC              = "NodeBMC"
	XnameTypeNode                 = "Node"
	XnameTypeRouterModule         = "RouterModule"
	XnameTypeRouterBMC            = "RouterBMC"
	XnameTypeMgmtSwitch           = "MgmtSwitch"
	XnameTypeMgmtHLSwitch         = "MgmtHLSwitch"
	XnameTypeMgmtSwitchConnector  = "MgmtSwitchConnector"
	XnameTypeNodeEnclosure        = "NodeEnclosure"
	XnameTypeHSNBoard             = "HSNBoard"
	XnameTypeCabinetPDUController = "CabinetPDUController"
	XnameTypeMgmtCDUSwitch        = "MgmtCDUSwitch"
)

// XnameInfo holds the parsed components of an xname string.
type XnameInfo struct {
	Raw     string
	Type    string
	Cabinet int
	Chassis int
	Slot    int // s=ComputeModule, w=MgmtSwitch, h=HSNConnEnclosure, r=RouterModule
	BMC     int
	Node    int
	Port    int // j=MgmtSwitchConnector port
}

// Parent returns the parent xname string using HMS-style derivation.
// Per hms-xname: trim trailing digits, then trim trailing letters.
func (x XnameInfo) Parent() string {
	return GetParentXname(x.Raw)
}

// GetParentXname derives the parent xname from any valid xname string.
// Cabinets return "s0". For all others, trim trailing digits then
// trailing letters (matches hms-xname GetHMSCompParent logic).
func GetParentXname(xname string) string {
	if xname == "" {
		return ""
	}
	// Cabinets (x<digits>) have system as parent
	if reCabinet.MatchString(xname) {
		return "s0"
	}
	// Trim trailing digits, then trailing letters
	pstr := trimRightFunc(xname, unicode.IsNumber)
	pstr = trimRightFunc(pstr, unicode.IsLetter)
	if pstr == "" {
		return ""
	}
	return pstr
}

// reCabinet matches a bare cabinet xname.
var reCabinet = regexp.MustCompile(`^x[0-9]{1,4}$`)

// trimRightFunc trims runes from the right while f returns true.
func trimRightFunc(s string, f func(rune) bool) string {
	runes := []rune(s)
	i := len(runes) - 1
	for i >= 0 && f(runes[i]) {
		i--
	}
	return string(runes[:i+1])
}

func formatXname(format string, args ...any) string {
	return sprintf(format, args...)
}

// xnamePattern defines a regex pattern to match and the type it identifies.
type xnamePattern struct {
	re       *regexp.Regexp
	typeName string
	extract  func(matches []string) XnameInfo
}

// Compiled patterns ordered from most specific (longest) to least specific.
// Regex patterns match hms-xname canonical definitions.
var xnamePatterns = []xnamePattern{
	// MgmtSwitchConnector: x<cab>c<ch>w<slot>j<port>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)j([1-9][0-9]*)$`),
		typeName: XnameTypeMgmtSwitchConnector,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), Port: atoi(m[4]),
			}
		},
	},
	// Node: x<cab>c<ch>s<slot>b<bmc>n<node>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)n([0-9]+)$`),
		typeName: XnameTypeNode,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]), Node: atoi(m[5]),
			}
		},
	},
	// NodeBMC: x<cab>c<ch>s<slot>b<bmc>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])s([0-9]+)b([0-9]+)$`),
		typeName: XnameTypeNodeBMC,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]),
			}
		},
	},
	// NodeEnclosure: x<cab>c<ch>s<slot>e<ordinal>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])s([0-9]+)e([0-9]+)$`),
		typeName: XnameTypeNodeEnclosure,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]),
			}
		},
	},
	// ComputeModule: x<cab>c<ch>s<slot>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])s([0-9]+)$`),
		typeName: XnameTypeComputeModule,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]),
			}
		},
	},
	// RouterBMC: x<cab>c<ch>r<slot>b<bmc>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])r([0-9]+)b([0-9]+)$`),
		typeName: XnameTypeRouterBMC,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]),
			}
		},
	},
	// HSNBoard: x<cab>c<ch>r<slot>e<ordinal>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])r([0-9]+)e([0-9]+)$`),
		typeName: XnameTypeHSNBoard,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]),
			}
		},
	},
	// RouterModule: x<cab>c<ch>r<slot>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])r([0-9]+)$`),
		typeName: XnameTypeRouterModule,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]),
			}
		},
	},
	// MgmtHLSwitch: x<cab>c<ch>h<slot>s<ordinal>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])h([1-9][0-9]*)s([1-9])$`),
		typeName: XnameTypeMgmtHLSwitch,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]), BMC: atoi(m[4]),
			}
		},
	},
	// MgmtSwitch: x<cab>c<ch>w<slot>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])w([1-9][0-9]*)$`),
		typeName: XnameTypeMgmtSwitch,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				Slot: atoi(m[3]),
			}
		},
	},
	// ChassisBMC: x<cab>c<ch>b<bmc>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])b([0])$`),
		typeName: XnameTypeChassisBMC,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
				BMC: atoi(m[3]),
			}
		},
	},
	// Chassis: x<cab>c<ch>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})c([0-7])$`),
		typeName: XnameTypeChassis,
		extract: func(m []string) XnameInfo {
			return XnameInfo{
				Cabinet: atoi(m[1]), Chassis: atoi(m[2]),
			}
		},
	},
	// Cabinet: x<cab>
	{
		re:       regexp.MustCompile(`^x([0-9]{1,4})$`),
		typeName: XnameTypeCabinet,
		extract: func(m []string) XnameInfo {
			return XnameInfo{Cabinet: atoi(m[1])}
		},
	},
}

// ParseXname extracts type and ordinals from an xname string.
func ParseXname(xname string) XnameInfo {
	for _, p := range xnamePatterns {
		matches := p.re.FindStringSubmatch(xname)
		if matches == nil {
			continue
		}
		info := p.extract(matches)
		info.Raw = xname
		info.Type = p.typeName
		return info
	}
	return XnameInfo{Raw: xname}
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// sprintf is a thin wrapper to avoid importing fmt for simple formatting.
func sprintf(format string, args ...any) string {
	// Use strconv-based approach for the simple "x%d" patterns we need.
	// For the xname formats, all args are ints so we build manually.
	result := make([]byte, 0, 32)
	argIdx := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' {
			if argIdx < len(args) {
				result = strconv.AppendInt(result, int64(args[argIdx].(int)), 10)
				argIdx++
			}
			i++ // skip 'd'
		} else {
			result = append(result, format[i])
		}
	}
	return string(result)
}
