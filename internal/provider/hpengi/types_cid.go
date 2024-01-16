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
package hpengi

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Cid struct {
	Version         string            `json:"version"`
	TenantID        string            `json:"tenantId"`
	SiteID          string            `json:"siteId"`
	HpeOrderDetails []HpeOrderDetails `json:"hpeOrderDetails"`
	CreatedBy       string            `json:"createdBy"`
	CreatedOn       string            `json:"createdOn"`
	LastUpdatedBy   string            `json:"lastUpdatedBy"`
	LastUpdatedOn   string            `json:"lastUpdatedOn"`
	Updatable       []string          `json:"updatable"`
	Deletable       bool              `json:"deletable"`
	Racks           []Racks           `json:"racks"`
}
type HpeOrderDetails struct {
	UcID                string   `json:"ucId"`
	Updatable           []string `json:"updatable"`
	Deletable           bool     `json:"deletable"`
	HpeSalesOrderNumber string   `json:"hpeSalesOrderNumber"`
	SfdcOpportunityID   string   `json:"sfdcOpportunityId"`
}
type AccessCredentials struct {
	Target    string   `json:"target"`
	UserName  string   `json:"userName"`
	Password  string   `json:"-"`
	Updatable []string `json:"updatable"`
	Deletable bool     `json:"deletable"`
}
type Pdus struct {
	ComponentID       string              `json:"componentId"`
	Type              string              `json:"type"`
	AccessCredentials []AccessCredentials `json:"accessCredentials"`
	Updatable         []string            `json:"updatable"`
	Deletable         bool                `json:"deletable"`
}
type NetworkSwitches struct {
	ComponentID        string              `json:"componentId"`
	Type               string              `json:"type"`
	AccessCredentials  []AccessCredentials `json:"accessCredentials"`
	Updatable          []string            `json:"updatable"`
	Deletable          bool                `json:"deletable"`
	RackElevationStart int                 `json:"rackElevationStart"`
	RackElevationEnd   int                 `json:"rackElevationEnd"`
}
type Servers struct {
	ComponentID        string              `json:"componentId"`
	Type               string              `json:"type"`
	AccessCredentials  []AccessCredentials `json:"accessCredentials"`
	LighthouseModuleID string              `json:"lighthouseModuleId"`
	Updatable          []string            `json:"updatable"`
	Deletable          bool                `json:"deletable"`
	AllocatedFor       string              `json:"allocatedFor"`
	RackElevationStart int                 `json:"rackElevationStart"`
	RackElevationEnd   int                 `json:"rackElevationEnd"`
	BayNumber          int                 `json:"bayNumber"`
}
type Chassis struct {
	Servers            []Servers `json:"servers"`
	ComponentID        string    `json:"componentId"`
	Type               string    `json:"type"`
	Updatable          []string  `json:"updatable"`
	Deletable          bool      `json:"deletable"`
	RackElevationStart int       `json:"rackElevationStart"`
	RackElevationEnd   int       `json:"rackElevationEnd"`
}
type Racks struct {
	ComponentID     string            `json:"componentId"`
	Pdus            []Pdus            `json:"pdus"`
	NetworkSwitches []NetworkSwitches `json:"networkSwitches"`
	Servers         []Servers         `json:"servers"`
	Chassis         []Chassis         `json:"chassis,omitempty"`
	Updatable       bool              `json:"updatable"`
	Deletable       bool              `json:"deletable"`
}

func (cid *Cid) getHardware() (imported map[uuid.UUID]inventory.Hardware, err error) {
	imported = make(map[uuid.UUID]inventory.Hardware, 0)
	for _, rack := range cid.Racks {
		log.Info().Msgf("Validating CID rack %+v", rack.ComponentID)
		cabinet := inventory.Hardware{
			ID:   uuid.New(),
			Name: rack.ComponentID,
			Type: hardwaretypes.Cabinet,
		}
		imported[cabinet.ID] = cabinet

		for _, chassis := range rack.Chassis {
			log.Info().Msgf("  Validating CID chassis %+v", chassis.ComponentID)
			chass := inventory.Hardware{
				ID:             uuid.New(),
				Name:           chassis.ComponentID,
				DeviceTypeSlug: chassis.Type,
				Type:           hardwaretypes.Chassis,
				Parent:         cabinet.ID,
			}
			imported[chass.ID] = chass

			for _, server := range chassis.Servers {
				log.Info().Msgf("    Validating %+v", server.ComponentID)
				hw := inventory.Hardware{
					ID:             uuid.New(),
					Name:           server.ComponentID,
					DeviceTypeSlug: server.Type,
					Type:           hardwaretypes.NodeBlade,
					Parent:         chass.ID,
				}
				imported[hw.ID] = hw
			}
		}
	}
	return imported, nil
}
