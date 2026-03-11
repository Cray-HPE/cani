package transform

import (
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

// geolocRe captures cabinet and chassis numbers from an xname string.
var geolocRe = regexp.MustCompile(`^x([0-9]{1,5})c([0-9]+)`)

// GeolocInfo holds cabinet and chassis extracted from a geoloc xname.
type GeolocInfo struct {
	Raw     string
	Cabinet int
	Chassis int
	Valid   bool
}

// ParseGeoloc extracts cabinet and chassis from a geoloc xname string.
// Returns Valid=false if the string is not a recognizable xname.
func ParseGeoloc(xname string) GeolocInfo {
	m := geolocRe.FindStringSubmatch(xname)
	if m == nil {
		return GeolocInfo{Raw: xname}
	}
	cab, _ := strconv.Atoi(m[1])
	ch, _ := strconv.Atoi(m[2])
	return GeolocInfo{
		Raw:     xname,
		Cabinet: cab,
		Chassis: ch,
		Valid:   true,
	}
}

// ParentChassisXname returns the chassis-level xname for a geoloc.
// For example, "x9000c1s7b0n0" returns "x9000c1".
func ParentChassisXname(xname string) string {
	info := ParseGeoloc(xname)
	if !info.Valid {
		return ""
	}
	return "x" + strconv.Itoa(info.Cabinet) + "c" + strconv.Itoa(info.Chassis)
}

// nodeGeolocXname extracts the geoloc xname from inventory or aliases.
// Checks inventory["geoloc"] first, then aliases["cm-geo-name"].
func nodeGeolocXname(inventory, aliases map[string]string) string {
	if inventory != nil {
		if geo, ok := inventory["geoloc"]; ok && geo != "" {
			return geo
		}
	}
	if aliases != nil {
		if geo, ok := aliases["cm-geo-name"]; ok && geo != "" {
			return geo
		}
	}
	return ""
}

// resolveGeolocParent finds the parent chassis UUID from a geoloc xname.
// Tries chassisByLoc (rack-chassis key) first, then chassisByXname.
func resolveGeolocParent(geoloc string, chassisByLoc, chassisByXname map[string]uuid.UUID) uuid.UUID {
	if geoloc == "" {
		return uuid.Nil
	}
	info := ParseGeoloc(geoloc)
	if !info.Valid {
		return uuid.Nil
	}
	// Try location-based lookup using cabinet/chassis from xname.
	key := chassisKey(int32(info.Cabinet), int32(info.Chassis))
	if id, ok := chassisByLoc[key]; ok {
		return id
	}
	// Try xname-based lookup.
	xname := ParentChassisXname(geoloc)
	if id, ok := chassisByXname[xname]; ok {
		return id
	}
	return uuid.Nil
}
