package import_

// LocationSettings holds the physical location fields from HPCM.
// Pointer fields preserve the JSON null-vs-0 distinction needed by
// the classification rules (last non-nil field determines depth).
type LocationSettings struct {
	Rack       *int32 `json:"rack"`
	Chassis    *int32 `json:"chassis"`
	Tray       *int32 `json:"tray"`
	Node       *int32 `json:"node"`
	Controller *int32 `json:"controller"`
}

// Int32Ptr returns a pointer to the given int32 value (test helper).
func Int32Ptr(v int32) *int32 { return &v }
