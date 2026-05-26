package import_

import (
	"encoding/json"
	"fmt"
)

// ParseSlsDumpstate parses raw JSON bytes into an SlsDumpstate.
func ParseSlsDumpstate(data []byte) (*SlsDumpstate, error) {
	var state SlsDumpstate
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing SLS dumpstate: %w", err)
	}
	if state.Hardware == nil {
		state.Hardware = make(map[string]SlsHardware)
	}
	if state.Networks == nil {
		state.Networks = make(map[string]SlsNetwork)
	}
	return &state, nil
}

// ParseSmdComponents parses raw JSON bytes into an SmdComponentList.
func ParseSmdComponents(data []byte) (*SmdComponentList, error) {
	var list SmdComponentList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parsing SMD components: %w", err)
	}
	return &list, nil
}

// DecodeExtraProperties converts a generic map into a typed struct
// using a JSON round-trip. This replaces the old mapstructure dependency.
func DecodeExtraProperties[T any](raw map[string]any) (T, error) {
	var result T
	if raw == nil {
		return result, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return result, fmt.Errorf("encoding extra properties: %w", err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("decoding extra properties: %w", err)
	}
	return result, nil
}
