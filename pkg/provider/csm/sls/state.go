/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
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

func CopyNetworks(in map[string]sls_client.Network) (map[string]sls_client.Network, error) {
	// This is a hack to easily create a copy of the SLS Networks
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var out map[string]sls_client.Network
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	return out, nil
}
