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

type Controller struct {
	Name                 string                 `json:"name,omitempty"`
	Id                   int64                  `json:"id,omitempty"`
	Uuid                 string                 `json:"uuid,omitempty"`
	Etag                 string                 `json:"etag,omitempty"`
	CreationTime         time.Time              `json:"creationTime,omitempty"`
	ModificationTime     time.Time              `json:"modificationTime,omitempty"`
	DeletionTime         time.Time              `json:"deletionTime,omitempty"`
	Links                map[string]string      `json:"links,omitempty"`
	Nodes                []Node                 `json:"nodes,omitempty"`
	Nic                  *Nic                   `json:"nic,omitempty"`
	Type_                string                 `json:"type,omitempty"`
	TemplateName         string                 `json:"templateName,omitempty"`
	IpAddress            string                 `json:"ipAddress,omitempty"`
	MacAddress           string                 `json:"macAddress,omitempty"`
	Protocol             string                 `json:"protocol,omitempty"`
	Channel              int32                  `json:"channel,omitempty"`
	Username             string                 `json:"username,omitempty"`
	Password             string                 `json:"password,omitempty"`
	AdministrativeStatus int32                  `json:"administrativeStatus,omitempty"`
	Location             *LocationSettings      `json:"location,omitempty"`
	Inventory            *interface{}           `json:"inventory,omitempty"`
	Attributes           map[string]interface{} `json:"attributes,omitempty"`
}
