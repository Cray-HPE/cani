/*
 * HPCM cmdb REST API Documentation
 *
 * HPE Performance Cluster Manager 'cmdb' service features a REST API. This section describes its implementation.  Standard REST API concepts (such as HTTP verbs, return codes, JSON, etc.) are not covered here.
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package hpcm_client

import (
	"time"
)

type Nic struct {
	Name             string            `json:"name,omitempty"`
	Id               int64             `json:"id,omitempty"`
	Uuid             string            `json:"uuid,omitempty"`
	Etag             string            `json:"etag,omitempty"`
	CreationTime     time.Time         `json:"creationTime,omitempty"`
	ModificationTime time.Time         `json:"modificationTime,omitempty"`
	DeletionTime     time.Time         `json:"deletionTime,omitempty"`
	Links            map[string]string `json:"links,omitempty"`
	// Write-only field to configure the network this nic is attached to at creation time
	Network       string `json:"network,omitempty"`
	IpAddress     string `json:"ipAddress,omitempty"`
	Ipv6Address   string `json:"ipv6Address,omitempty"`
	MacAddress    string `json:"macAddress,omitempty"`
	BondingMaster string `json:"bondingMaster,omitempty"`
	BondingMode   string `json:"bondingMode,omitempty"`
	InterfaceName string `json:"interfaceName,omitempty"`
	Managed       bool   `json:"managed,omitempty"`
	Type_         string `json:"type,omitempty"`
	// Write-only field to configure the node this nic is attached to at creation time
	Node     string `json:"node,omitempty"`
	NodeName string `json:"nodeName,omitempty"`
	// Write-only field to configure the controller this nic is attached to at creation time
	Controller     string                 `json:"controller,omitempty"`
	ControllerName string                 `json:"controllerName,omitempty"`
	NetworkName    string                 `json:"networkName,omitempty"`
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
}
