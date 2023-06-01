package sls

import (
	"encoding/json"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

func CopyState(in sls_client.SlsState) (sls_client.SlsState, error) {
	// This is a hack to easily create a copy of the SLS state
	raw, err := json.Marshal(in)
	if err != nil {
		return sls_client.SlsState{}, err
	}

	var out sls_client.SlsState
	if err := json.Unmarshal(raw, &out); err != nil {
		return sls_client.SlsState{}, err
	}

	return out, nil
}
