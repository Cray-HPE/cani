package sls

import (
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type Dumpstate sls_client.SlsState

func (ds *Dumpstate) GetHardware() map[string]sls_client.Hardware {
	return ds.Hardware
}

func (ds *Dumpstate) GetNetworks() map[string]sls_client.Network {
	return ds.Networks
}
