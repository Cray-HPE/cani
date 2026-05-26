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
package import_

import (
	"github.com/Cray-HPE/cani/pkg/provider/hpcm/import/cmdb"
)

// Cmdb holds the full set of resources fetched from the HPCM CMDB REST API.
type Cmdb struct {
	Controllers     []cmdb.Group          `json:"controllers,omitempty"`
	CustomGroups    []cmdb.Group          `json:"customGroups,omitempty"`
	ImageGroups     []cmdb.Group          `json:"imageGroups,omitempty"`
	ManagementCards []cmdb.ManagementCard `json:"managementCards,omitempty"`
	NetworkGroups   []cmdb.Group          `json:"networkGroups,omitempty"`
	Networks        []cmdb.Network        `json:"networks,omitempty"`
	Nics            []cmdb.Nic            `json:"nics,omitempty"`
	NodeTemplates   []cmdb.NodeTemplate   `json:"nodeTemplates,omitempty"`
	Nodes           []cmdb.Node           `json:"nodes,omitempty"`
	SystemGroups    []cmdb.Group          `json:"systemGroups,omitempty"`
}
