/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

// Package cmdb defines the data types returned by the HPCM CMDB REST API.
// These mirror the models from the swagger-generated hpcm-client package
// so that the import package can deserialize CMDB JSON without depending
// on the generated client.
package cmdb

import "time"

// Group represents a controller, custom group, image group, network group,
// or system group resource from the CMDB.
type Group struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	Nodes            []Node                 `json:"nodes,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}

// ManagementCard represents a BMC / management card resource from the CMDB.
type ManagementCard struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	Scannable        bool                   `json:"scannable,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}

// Network represents a network resource from the CMDB.
type Network struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	BaseIp           string                 `json:"baseIp,omitempty"`
	Broadcast        string                 `json:"broadcast,omitempty"`
	Gateway          string                 `json:"gateway,omitempty"`
	Type             string                 `json:"type,omitempty"`
	Vlan             int32                  `json:"vlan,omitempty"`
	Netmask          string                 `json:"netmask,omitempty"`
	RackNetmask      string                 `json:"rackNetmask,omitempty"`
	MgmtServerIp     string                 `json:"mgmtServerIp,omitempty"`
	Mtu              int32                  `json:"mtu,omitempty"`
	Rack             int32                  `json:"rack,omitempty"`
	Nics             []Nic                  `json:"nics,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}

// NodeTemplate represents a node template resource from the CMDB.
type NodeTemplate struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	NicTemplates     []NicTemplate          `json:"nicTemplates,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}

// NicTemplate represents a NIC template attached to a NodeTemplate.
type NicTemplate struct {
	Network       string   `json:"network"`
	BondingMaster string   `json:"bondingMaster,omitempty"`
	BondingMode   string   `json:"bondingMode,omitempty"`
	Bridge        string   `json:"bridge,omitempty"`
	Interfaces    []string `json:"interfaces,omitempty"`
}

// Node is a lightweight representation of a CMDB node for JSON deserialization.
// The authoritative Node type with all business-logic fields lives in the
// parent import_ package; this copy exists solely to break the import cycle.
type Node struct {
	Name                 string                 `json:"name,omitempty"`
	Aliases              map[string]string      `json:"aliases,omitempty"`
	Id                   int64                  `json:"id,omitempty"`
	Uuid                 string                 `json:"uuid,omitempty"`
	Etag                 string                 `json:"etag,omitempty"`
	CreationTime         time.Time              `json:"creationTime,omitempty"`
	ModificationTime     time.Time              `json:"modificationTime,omitempty"`
	DeletionTime         time.Time              `json:"deletionTime,omitempty"`
	Links                map[string]string      `json:"links,omitempty"`
	Network              *NetworkSettings       `json:"network,omitempty"`
	Image                *ImageSettings         `json:"image,omitempty"`
	Platform             *PlatformSettings      `json:"platform,omitempty"`
	Management           *ManagementSettings    `json:"management,omitempty"`
	Controller           *ControllerSettings    `json:"controller,omitempty"`
	Location             *LocationSettings      `json:"location,omitempty"`
	InternalName         string                 `json:"internalName,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	ImageTransport       string                 `json:"imageTransport,omitempty"`
	ImagePending         bool                   `json:"imagePending,omitempty"`
	TemplateName         string                 `json:"templateName,omitempty"`
	RootFs               string                 `json:"rootFs,omitempty"`
	OperationalStatus    int32                  `json:"operationalStatus,omitempty"`
	AdministrativeStatus int32                  `json:"administrativeStatus,omitempty"`
	Managed              bool                   `json:"managed,omitempty"`
	Monitoring           string                 `json:"monitoring,omitempty"`
	RootSlot             int32                  `json:"rootSlot,omitempty"`
	BiosBootMode         string                 `json:"biosBootMode,omitempty"`
	BootOrder            int32                  `json:"bootOrder,omitempty"`
	IscsiRoot            string                 `json:"iscsiRoot,omitempty"`
	Inventory            *interface{}           `json:"inventory,omitempty"`
	NodeController       string                 `json:"nodeController,omitempty"`
	Attributes           map[string]interface{} `json:"attributes,omitempty"`
}

// Nic is a lightweight representation of a CMDB NIC for JSON deserialization.
// The authoritative Nic type lives in the parent import_ package.
type Nic struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	Network          string                 `json:"network,omitempty"`
	IpAddress        string                 `json:"ipAddress,omitempty"`
	Ipv6Address      string                 `json:"ipv6Address,omitempty"`
	MacAddress       string                 `json:"macAddress,omitempty"`
	BondingMaster    string                 `json:"bondingMaster,omitempty"`
	BondingMode      string                 `json:"bondingMode,omitempty"`
	InterfaceName    string                 `json:"interfaceName,omitempty"`
	Managed          bool                   `json:"managed,omitempty"`
	Type             string                 `json:"type,omitempty"`
	Node             string                 `json:"node,omitempty"`
	NodeName         string                 `json:"nodeName,omitempty"`
	Controller       string                 `json:"controller,omitempty"`
	ControllerName   string                 `json:"controllerName,omitempty"`
	NetworkName      string                 `json:"networkName,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}

// --- Sub-types referenced by Node ---

// NetworkSettings holds network configuration embedded in a Node.
type NetworkSettings struct {
	Name           string `json:"name,omitempty"`
	DefaultGateway string `json:"defaultGateway,omitempty"`
	Nics           []Nic  `json:"nics,omitempty"`
	IpAddress      string `json:"ipAddress,omitempty"`
	MacAddress     string `json:"macAddress,omitempty"`
	SubnetMask     string `json:"subnetMask,omitempty"`
	MgmtServerIp   string `json:"mgmtServerIp,omitempty"`
}

// ImageSettings holds image configuration embedded in a Node.
type ImageSettings struct {
	Name               string    `json:"name,omitempty"`
	Kernel             string    `json:"kernel,omitempty"`
	CloningBlockDevice string    `json:"cloningBlockDevice,omitempty"`
	CloningDate        time.Time `json:"cloningDate,omitempty"`
}

// PlatformSettings holds platform configuration embedded in a Node.
type PlatformSettings struct {
	Name            string `json:"name,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	SerialPort      string `json:"serialPort,omitempty"`
	SerialPortSpeed string `json:"serialPortSpeed,omitempty"`
	VendorsArgs     string `json:"vendorsArgs,omitempty"`
}

// ManagementSettings holds management card configuration embedded in a Node.
type ManagementSettings struct {
	CardType       string `json:"cardType,omitempty"`
	CardIpAddress  string `json:"cardIpAddress,omitempty"`
	CardMacAddress string `json:"cardMacAddress,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Channel        int32  `json:"channel,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
}

// ControllerSettings holds controller configuration embedded in a Node.
type ControllerSettings struct {
	Type       string `json:"type,omitempty"`
	IpAddress  string `json:"ipAddress,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Channel    int32  `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

// LocationSettings holds the physical location fields embedded in a Node.
type LocationSettings struct {
	Rack       int32 `json:"rack,omitempty"`
	Chassis    int32 `json:"chassis,omitempty"`
	Tray       int32 `json:"tray,omitempty"`
	Node       int32 `json:"node,omitempty"`
	Controller int32 `json:"controller,omitempty"`
}
