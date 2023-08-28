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

package checks

type CabinetNetworkName string

const (
	cn  CabinetNetworkName = "cn"
	ncn CabinetNetworkName = "ncn"
)

func (n CabinetNetworkName) String() string {
	return string(n)
}

type NetworkName string

const (
	HMN_MTN NetworkName = "HMN_MTN"
	HMN_RVR NetworkName = "HMN_RVR"
	NMN_MTN NetworkName = "NMN_MTN"
	NMN_RVR NetworkName = "NMN_RVR"
	HMN     NetworkName = "HMN"
	NMN     NetworkName = "NMN"
)

func (n NetworkName) String() string {
	return string(n)
}

type NetworkId string

const (
	hmnMtnId NetworkId = "/Networks/HMN_MTN"
	nmnMtnId NetworkId = "/Networks/NMN_MTN"
	hmnRvrId NetworkId = "/Networks/HMN_RVR"
	nmnRvrId NetworkId = "/Networks/NMN_RVR"
)

func (n NetworkId) String() string {
	return string(n)
}

type CabinetNetworkField string

const (
	CIDR    CabinetNetworkField = "CIDR"
	Gateway CabinetNetworkField = "Gateway"
	VLan    CabinetNetworkField = "VLan"
)

func (n CabinetNetworkField) String() string {
	return string(n)
}
